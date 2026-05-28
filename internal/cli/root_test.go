package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	config "github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestRunBoardCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"title":"Test Board","color":"ff0000","archived":false}`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "create", "--title", "Test Board", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if payload["id"].(float64) != 7 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRunCardCreateReadsDescriptionFile(t *testing.T) {
	descriptionPath := t.TempDir() + "/description.md"
	if err := os.WriteFile(descriptionPath, []byte("line 1\nline 2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/1/stacks/2/cards" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if payload["description"] != "line 1\nline 2\n" {
			t.Fatalf("description = %#v", payload["description"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"title":"Test","description":"line 1\nline 2\n","stackId":2,"type":"plain","order":999,"archived":false}`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"card", "create", "--board", "1", "--stack", "2", "--title", "Test", "--description-file", descriptionPath}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
}

func TestRunCommentCreateReadsCommentStdin(t *testing.T) {
	withCommandStdin(t, "from\nstdin")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/ocs/v2.php/apps/deck/api/v1.0/cards/3/comments" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"message":"from\nstdin"`) {
			t.Fatalf("body = %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ocs":{"meta":{"status":"ok","statuscode":200,"message":"OK"},"data":{"id":9,"objectId":3,"message":"from\nstdin"}}}`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"comment", "create", "--card", "3", "--comment-stdin"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
}

func TestRunBoardListDefaultsToTextAndSupportsJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Test Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(board list) error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tTest Board\n" {
		t.Fatalf("default text output = %q", got)
	}

	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"board", "list", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(board list --json) error = %v; stderr=%s", err, stderr.String())
	}
	var boards []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &boards); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if len(boards) != 1 || boards[0]["title"] != "Test Board" {
		t.Fatalf("unexpected boards: %#v", boards)
	}

	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"board", "list", "--text"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(board list --text) error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tTest Board\n" {
		t.Fatalf("text output = %q", got)
	}
}

func TestRunListFriendlyBoardSelectorsAndAliases(t *testing.T) {
	var writes int32
	var boardLookups int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			atomic.AddInt32(&writes, 1)
			http.Error(w, "unexpected write", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards":
			atomic.AddInt32(&boardLookups, 1)
			_, _ = w.Write([]byte(`[{"id":14,"title":"DAILY TODO","color":"ff0000","archived":false},{"id":27,"title":"Other","color":"00ff00","archived":false}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks":
			_, _ = w.Write([]byte(`[{"id":2,"title":"Todo","boardId":14,"order":1}]`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	tests := []struct {
		name string
		args []string
	}{
		{name: "list flag substring", args: []string{"list", "--board", "daily"}},
		{name: "list board positional", args: []string{"list", "board", "daily"}},
		{name: "stack alias", args: []string{"stack", "--board", "daily"}},
		{name: "stacks alias", args: []string{"stacks", "--board", "daily"}},
		{name: "exact title", args: []string{"list", "--board", "DAILY TODO"}},
		{name: "numeric compatibility", args: []string{"list", "list", "--board", "14"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if got := stdout.String(); got != "2\tTodo\t1\n" {
				t.Fatalf("stdout = %q", got)
			}
		})
	}

	if got := atomic.LoadInt32(&writes); got != 0 {
		t.Fatalf("made %d write requests", got)
	}
	if got, want := atomic.LoadInt32(&boardLookups), int32(5); got != want {
		t.Fatalf("board lookups = %d, want %d", got, want)
	}
}

func TestRunListLaneAndStackSelectorsListCards(t *testing.T) {
	var writes int32
	var stackListReads int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			atomic.AddInt32(&writes, 1)
			http.Error(w, "unexpected write", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks":
			atomic.AddInt32(&stackListReads, 1)
			_, _ = w.Write([]byte(`[{"id":96,"title":"Heute","boardId":14,"order":1},{"id":97,"title":"Morgen","boardId":14,"order":2}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks/96":
			_, _ = w.Write([]byte(`{"id":96,"title":"Heute","boardId":14,"order":1,"cards":[{"id":91,"title":"Call","stackId":96,"type":"plain","order":1,"archived":false}]}`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	tests := []struct {
		name string
		args []string
	}{
		{name: "lane title", args: []string{"list", "--board", "14", "--lane", "Heute", "--json"}},
		{name: "stack title", args: []string{"list", "--board", "14", "--stack", "Heute", "--json"}},
		{name: "stack numeric", args: []string{"list", "--board", "14", "--stack", "96", "--json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			var cards []map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &cards); err != nil {
				t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
			}
			if len(cards) != 1 || cards[0]["id"] != float64(91) || cards[0]["title"] != "Call" {
				t.Fatalf("cards = %#v", cards)
			}
		})
	}

	if got := atomic.LoadInt32(&writes); got != 0 {
		t.Fatalf("made %d write requests", got)
	}
	if got, want := atomic.LoadInt32(&stackListReads), int32(2); got != want {
		t.Fatalf("stack list reads = %d, want %d", got, want)
	}
}

func TestRunCardListBoardAndStackTitleSelectors(t *testing.T) {
	var writes int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			atomic.AddInt32(&writes, 1)
			http.Error(w, "unexpected write", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards":
			_, _ = w.Write([]byte(`[{"id":14,"title":"DAILY TODO","color":"ff0000","archived":false}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks":
			_, _ = w.Write([]byte(`[{"id":96,"title":"Heute","boardId":14,"order":1}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks/96":
			_, _ = w.Write([]byte(`{"id":96,"title":"Heute","boardId":14,"order":1,"cards":[{"id":91,"title":"Call","stackId":96,"type":"plain","order":1,"archived":false}]}`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	var stdout, stderr bytes.Buffer
	args := []string{"card", "list", "--board", "DAILY TODO", "--stack", "Heute", "--json"}
	if err := Run(args, &stdout, &stderr); err != nil {
		t.Fatalf("Run(%v) error = %v; stderr=%s", args, err, stderr.String())
	}
	var cards []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &cards); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if len(cards) != 1 || cards[0]["id"] != float64(91) || cards[0]["title"] != "Call" {
		t.Fatalf("cards = %#v", cards)
	}
	if got := atomic.LoadInt32(&writes); got != 0 {
		t.Fatalf("made %d write requests", got)
	}
}

func TestRunListBoardTitleValidationErrors(t *testing.T) {
	var writes int32
	var stackReads int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			atomic.AddInt32(&writes, 1)
			http.Error(w, "unexpected write", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards":
			_, _ = w.Write([]byte(`[{"id":14,"title":"Daily Todo","color":"ff0000","archived":false},{"id":15,"title":"Daily Work","color":"00ff00","archived":false}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks", "/index.php/apps/deck/api/v1.0/boards/15/stacks":
			atomic.AddInt32(&stackReads, 1)
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "multiple", args: []string{"list", "--board", "daily"}, want: `board "daily" matched 2 boards: 14 "Daily Todo", 15 "Daily Work"; use a numeric board id or a more specific title`},
		{name: "unknown", args: []string{"list", "--board", "missing"}, want: `board "missing" not found; use a board id, exact title, or a unique case-insensitive title substring`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := Run(tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", tt.args, stdout.String(), stderr.String())
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}

	if got := atomic.LoadInt32(&writes); got != 0 {
		t.Fatalf("made %d write requests", got)
	}
	if got := atomic.LoadInt32(&stackReads); got != 0 {
		t.Fatalf("read stacks %d times before resolving a unique board", got)
	}
}

func TestRunStackSelectorValidationErrors(t *testing.T) {
	var writes int32
	var cardReads int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			atomic.AddInt32(&writes, 1)
			http.Error(w, "unexpected write", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks":
			_, _ = w.Write([]byte(`[{"id":96,"title":"Heute","boardId":14,"order":1},{"id":97,"title":"Heute Abend","boardId":14,"order":2}]`))
		case "/index.php/apps/deck/api/v1.0/boards/14/stacks/96":
			atomic.AddInt32(&cardReads, 1)
			_, _ = w.Write([]byte(`{"id":96,"title":"Heute","boardId":14,"order":1,"cards":[]}`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "ambiguous", args: []string{"list", "--board", "14", "--stack", "Heut"}, want: `stack/lane "Heut" matched 2 stacks: 96 "Heute", 97 "Heute Abend"; use a numeric stack id or a more specific title`},
		{name: "unknown", args: []string{"card", "list", "--board", "14", "--stack", "Missing"}, want: `stack/lane "Missing" not found on board 14; use a stack id, exact title, or a unique case-insensitive title substring`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := Run(tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", tt.args, stdout.String(), stderr.String())
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}

	if got := atomic.LoadInt32(&writes); got != 0 {
		t.Fatalf("made %d write requests", got)
	}
	if got := atomic.LoadInt32(&cardReads); got != 0 {
		t.Fatalf("read cards %d times before resolving a unique stack", got)
	}
}

func TestRunListRejectsConflictingLaneAndStackWithoutAPI(t *testing.T) {
	runInvalidCommandDoesNotCallAPI(t, []string{"list", "--board", "14", "--lane", "Heute", "--stack", "Heute"}, "use only one of --lane or --stack")
}

func TestRunListArchivedRejectsLaneAndStackWithoutAPI(t *testing.T) {
	runInvalidCommandDoesNotCallAPI(t, []string{"list", "archived", "--board", "14", "--stack", "Heute"}, "list archived does not support --lane or --stack")
}

func TestRunUnknownListCommandShowsExamples(t *testing.T) {
	clearNextcloudEnv(t)

	var stdout, stderr bytes.Buffer
	err := Run([]string{"list", "bogus"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run(list bogus) succeeded")
	}
	for _, want := range []string{
		`unknown list command "bogus"`,
		"deck list --board <id-or-title>",
		"deck list find --board <id-or-title> --title <list-title>",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want substring %q", err.Error(), want)
		}
	}
}

func TestRunBoardListLoadsSavedConfigWithoutEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "antonia" || pass != "pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Saved Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tSaved Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListEnvOverridesSavedConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "env-user" || pass != "env-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Env Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: "https://saved.example.com", Username: "saved-user", Password: "saved-pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "env-user")
	t.Setenv("NEXTCLOUD_PASSWORD", "env-pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tEnv Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListSelectsSavedProfileWithGlobalFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "work-user" || pass != "work-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Work Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: "https://default.example.com", Username: "default-user", Password: "default-pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := (config.Config{BaseURL: server.URL, Username: "work-user", Password: "work-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	for _, args := range [][]string{
		{"--profile", "work", "board", "list"},
		{"board", "list", "--profile", "work"},
		{"board", "list", "--profile=work"},
	} {
		var stdout, stderr bytes.Buffer
		if err := Run(args, &stdout, &stderr); err != nil {
			t.Fatalf("Run(%v) error = %v; stderr=%s", args, err, stderr.String())
		}
		if got := stdout.String(); got != "7\tWork Board\n" {
			t.Fatalf("stdout = %q", got)
		}
	}
}

func TestRunBoardListSelectsDeckProfileEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "work-user" || pass != "work-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Env Profile Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: server.URL, Username: "work-user", Password: "work-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	t.Setenv("DECK_PROFILE", "work")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tEnv Profile Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListCLIProfileOverridesDeckProfileEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "work-user" || pass != "work-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"CLI Profile Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: "https://personal.example.com", Username: "personal-user", Password: "personal-pw"}).SaveProfile("personal"); err != nil {
		t.Fatalf("SaveProfile(personal) error = %v", err)
	}
	if err := (config.Config{BaseURL: server.URL, Username: "work-user", Password: "work-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile(work) error = %v", err)
	}
	t.Setenv("DECK_PROFILE", "personal")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--profile", "work", "board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tCLI Profile Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListProfileCredentialEnvOverridesSelectedProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "env-user" || pass != "env-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Env Override Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: "https://saved.example.com", Username: "saved-user", Password: "saved-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "env-user")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "env-pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--profile", "work", "board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tEnv Override Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListDefaultProfileAliasIgnoresDeckProfileEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "default-user" || pass != "default-pw" {
			t.Fatalf("basic auth = %q %q %v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Default Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	if err := (config.Config{BaseURL: server.URL, Username: "default-user", Password: "default-pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := (config.Config{BaseURL: "https://work.example.com", Username: "work-user", Password: "work-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	t.Setenv("DECK_PROFILE", "work")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--profile", "default", "board", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "7\tDefault Board\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBoardListMissingProfileError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	if err := (config.Config{BaseURL: "https://default.example.com", Username: "default-user", Password: "default-pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := Run([]string{"board", "list", "--profile", "missing"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), `profile "missing" not found`) {
		t.Fatalf("err = %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}

func TestRunProfileOptionErrors(t *testing.T) {
	for _, args := range [][]string{{"--profile"}, {"--profile", ""}, {"--profile="}} {
		var stdout, stderr bytes.Buffer
		if err := Run(args, &stdout, &stderr); err == nil {
			t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
		}
	}
}

func TestRunAuthSetupSavesConfigWithoutEnvAndOpensSecurityURL(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	withCommandStdin(t, "nextcloud.xhacker.de/\n antonia \n app-pw \n")

	var opened []string
	oldOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		opened = append(opened, target)
		return nil
	}
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"auth", "setup"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth setup) error = %v; stderr=%s", err, stderr.String())
	}
	if len(opened) != 1 || opened[0] != "https://nextcloud.xhacker.de/settings/user/security" {
		t.Fatalf("opened = %#v", opened)
	}
	for _, want := range []string{"Nextcloud base URL: ", "Nextcloud username: ", "Open this URL to create an app password: https://nextcloud.xhacker.de/settings/user/security", "Nextcloud app password: ", "Saved local auth config."} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
		}
	}
	path := defaultTestConfigPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	wantJSON := "{\"base_url\":\"https://nextcloud.xhacker.de\",\"username\":\"antonia\",\"app_password\":\"app-pw\"}\n"
	if got := string(data); got != wantJSON {
		t.Fatalf("saved config = %q, want %q", got, wantJSON)
	}
}

func TestRunAuthSetupAllowsLocalhostHTTPBaseURL(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	withCommandStdin(t, "http://localhost:8080/root\nantonia\napp-pw\n")

	var opened []string
	oldOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		opened = append(opened, target)
		return nil
	}
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"auth", "setup"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth setup) error = %v; stderr=%s", err, stderr.String())
	}
	if len(opened) != 1 || opened[0] != "http://localhost:8080/root/settings/user/security" {
		t.Fatalf("opened = %#v", opened)
	}
	if !strings.Contains(stdout.String(), "Saved local auth config.") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	path := defaultTestConfigPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	wantJSON := "{\"base_url\":\"http://localhost:8080/root\",\"username\":\"antonia\",\"app_password\":\"app-pw\"}\n"
	if got := string(data); got != wantJSON {
		t.Fatalf("saved config = %q, want %q", got, wantJSON)
	}
}

func TestRunAuthSetupSavesNamedProfileWithoutTouchingDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	if err := (config.Config{BaseURL: "https://default.example.com", Username: "default", Password: "default-pw"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	withCommandStdin(t, "nextcloud.xhacker.de/\n antonia \n app-pw \n")

	oldOpenBrowser := openBrowser
	openBrowser = func(string) error { return nil }
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"auth", "setup", "--profile", "work"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth setup --profile) error = %v; stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `Saved local auth profile "work".`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
	defaultConfig, err := config.LoadProfile("default")
	if err != nil {
		t.Fatalf("LoadProfile(default) error = %v", err)
	}
	if defaultConfig.BaseURL != "https://default.example.com" || defaultConfig.Username != "default" || defaultConfig.Password != "default-pw" {
		t.Fatalf("default config = %#v", defaultConfig)
	}
	profile, err := config.LoadProfile("work")
	if err != nil {
		t.Fatalf("LoadProfile(work) error = %v", err)
	}
	if profile.BaseURL != "https://nextcloud.xhacker.de" || profile.Username != "antonia" || profile.Password != "app-pw" {
		t.Fatalf("profile config = %#v", profile)
	}
}

func TestRunAuthSetupReplacesDuplicateNamedProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	if err := (config.Config{BaseURL: "https://old.example.com", Username: "old", Password: "old-pw"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	withCommandStdin(t, "nextcloud.xhacker.de/\n antonia \n app-pw \n")

	oldOpenBrowser := openBrowser
	openBrowser = func(string) error { return nil }
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--profile", "work", "auth", "setup"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth setup) error = %v; stderr=%s", err, stderr.String())
	}
	profile, err := config.LoadProfile("work")
	if err != nil {
		t.Fatalf("LoadProfile(work) error = %v", err)
	}
	if profile.BaseURL != "https://nextcloud.xhacker.de" || profile.Username != "antonia" || profile.Password != "app-pw" {
		t.Fatalf("profile config = %#v", profile)
	}
}

func TestRunAuthProfilesListsProfilesWithoutSecrets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	if err := (config.Config{BaseURL: "https://default.example.com", Username: "default", Password: "default-secret"}).Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := (config.Config{BaseURL: "https://work.example.com", Username: "work", Password: "work-secret"}).SaveProfile("work"); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"auth", "profiles"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth profiles) error = %v; stderr=%s", err, stderr.String())
	}
	for _, want := range []string{"default\thttps://default.example.com\tdefault", "work\thttps://work.example.com\twork"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if strings.Contains(stdout.String(), "secret") {
		t.Fatalf("stdout leaked password: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"auth", "profiles", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth profiles --json) error = %v; stderr=%s", err, stderr.String())
	}
	if strings.Contains(stdout.String(), "secret") {
		t.Fatalf("json leaked password: %q", stdout.String())
	}
	var profiles []config.ProfileSummary
	if err := json.Unmarshal(stdout.Bytes(), &profiles); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if len(profiles) != 2 || profiles[0].Name != "default" || profiles[1].Name != "work" {
		t.Fatalf("profiles = %#v", profiles)
	}
}

func TestRunAuthSetupRejectsExternalHTTPBaseURL(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	withCommandStdin(t, "http://cloud.example.com\n")

	var opened []string
	oldOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		opened = append(opened, target)
		return nil
	}
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	err := Run([]string{"auth", "setup"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unsafe base URL")
	}
	if !strings.Contains(err.Error(), "http://cloud.example.com") {
		t.Fatalf("error = %v", err)
	}
	if len(opened) != 0 {
		t.Fatalf("opened = %#v", opened)
	}
	if strings.Contains(stdout.String(), "Nextcloud username:") || strings.Contains(stdout.String(), "Nextcloud app password:") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunAuthSetupContinuesWhenBrowserOpenFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearNextcloudEnv(t)
	withCommandStdin(t, "https://nextcloud.xhacker.de\nantonia\napp-pw\n")

	oldOpenBrowser := openBrowser
	openBrowser = func(string) error { return errors.New("no browser") }
	t.Cleanup(func() { openBrowser = oldOpenBrowser })

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"auth", "setup"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(auth setup) error = %v; stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Could not open browser: no browser") || !strings.Contains(stdout.String(), "Saved local auth config.") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(defaultTestConfigPath(t)); err != nil {
		t.Fatalf("expected saved config: %v", err)
	}
}

func TestRunOutputFormatOptionsForDataCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Test Board","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	tests := []struct {
		name string
		args []string
		json bool
	}{
		{name: "default text", args: []string{"board", "list"}},
		{name: "prefix --text", args: []string{"--text", "board", "list"}},
		{name: "suffix --text", args: []string{"board", "list", "--text"}},
		{name: "prefix -o text", args: []string{"-o", "text", "board", "list"}},
		{name: "suffix -o text", args: []string{"board", "list", "-o", "text"}},
		{name: "equals -o text", args: []string{"board", "list", "-o=text"}},
		{name: "prefix --output text", args: []string{"--output", "text", "board", "list"}},
		{name: "suffix --output text", args: []string{"board", "list", "--output", "text"}},
		{name: "equals --output text", args: []string{"board", "list", "--output=text"}},
		{name: "table alias", args: []string{"board", "list", "-o", "table"}},
		{name: "prefix --json", args: []string{"--json", "board", "list"}, json: true},
		{name: "suffix --json", args: []string{"board", "list", "--json"}, json: true},
		{name: "prefix -o json", args: []string{"-o", "json", "board", "list"}, json: true},
		{name: "suffix -o json", args: []string{"board", "list", "-o", "json"}, json: true},
		{name: "equals -o json", args: []string{"board", "list", "-o=json"}, json: true},
		{name: "prefix --output json", args: []string{"--output", "json", "board", "list"}, json: true},
		{name: "suffix --output json", args: []string{"board", "list", "--output", "json"}, json: true},
		{name: "equals --output json", args: []string{"board", "list", "--output=json"}, json: true},
		{name: "last option wins text", args: []string{"board", "list", "--json", "--text"}},
		{name: "last option wins json", args: []string{"board", "list", "--text", "--json"}, json: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if tt.json {
				var boards []map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &boards); err != nil {
					t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
				}
				if len(boards) != 1 || boards[0]["id"].(float64) != 7 {
					t.Fatalf("unexpected JSON output: %#v", boards)
				}
				return
			}
			if got := stdout.String(); got != "7\tTest Board\n" {
				t.Fatalf("text output = %q", got)
			}
		})
	}
}

func TestRunDeleteDefaultsToTextAndSupportsJSONStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/7" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "delete", "--board", "7"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(board delete) error = %v; stderr=%s", err, stderr.String())
	}
	if got := stdout.String(); got != "deleted board 7\n" {
		t.Fatalf("default delete output = %q", got)
	}

	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"board", "delete", "--board", "7", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(board delete --json) error = %v; stderr=%s", err, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if payload["status"] != "deleted" || payload["boardId"].(float64) != 7 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRunOutputFormatOptionsForStatusCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/7" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	tests := []struct {
		name string
		args []string
		json bool
	}{
		{name: "default text", args: []string{"board", "delete", "--board", "7"}},
		{name: "--text", args: []string{"board", "delete", "--board", "7", "--text"}},
		{name: "-o text", args: []string{"board", "delete", "--board", "7", "-o", "text"}},
		{name: "--output=text", args: []string{"board", "delete", "--board", "7", "--output=text"}},
		{name: "--json", args: []string{"board", "delete", "--board", "7", "--json"}, json: true},
		{name: "-o json", args: []string{"board", "delete", "--board", "7", "-o", "json"}, json: true},
		{name: "--output=json", args: []string{"board", "delete", "--board", "7", "--output=json"}, json: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if tt.json {
				var payload map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
				}
				if payload["status"] != "deleted" || payload["boardId"].(float64) != 7 {
					t.Fatalf("unexpected JSON status: %#v", payload)
				}
				return
			}
			if got := stdout.String(); got != "deleted board 7\n" {
				t.Fatalf("text status = %q", got)
			}
		})
	}
}

func TestRunOutputFormatOptionErrors(t *testing.T) {
	for _, args := range [][]string{
		{"-o"},
		{"--output"},
		{"-o", "yaml"},
		{"--output=yaml"},
	} {
		var stdout, stderr bytes.Buffer
		if err := Run(args, &stdout, &stderr); err == nil {
			t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
		}
	}
}

func TestRunTimeoutFlagCancelsSlowRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
	t.Setenv("DECK_TIMEOUT", "5m")

	var stdout, stderr bytes.Buffer
	err := Run([]string{"--timeout", "1ms", "board", "list"}, &stdout, &stderr)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want deadline exceeded", err)
	}
}

func TestRunDeckTimeoutEnvCancelsSlowRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
	t.Setenv("DECK_TIMEOUT", "1ms")

	var stdout, stderr bytes.Buffer
	err := Run([]string{"board", "list"}, &stdout, &stderr)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want deadline exceeded", err)
	}
}

func TestRunTimeoutOptionErrors(t *testing.T) {
	for _, args := range [][]string{
		{"--timeout"},
		{"--timeout", "0s"},
		{"--timeout=nope"},
	} {
		var stdout, stderr bytes.Buffer
		if err := Run(args, &stdout, &stderr); err == nil {
			t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
		}
	}
}

func TestMainValidationErrorOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--output", "yaml"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Main() exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got, want := stderr.String(), "error: validation: unsupported output format \"yaml\"; supported formats: json, text\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestMainCommandValidationKeepsJSONStdoutEmpty(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "https://cloud.example.com")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--json", "board", "create"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Main() exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got, want := stderr.String(), "error: validation: board create requires --title\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestMainAPIErrorKinds(t *testing.T) {
	for _, tt := range []struct {
		name       string
		statusCode int
		kind       string
	}{
		{name: "api", statusCode: http.StatusBadRequest, kind: "api"},
		{name: "auth", statusCode: http.StatusUnauthorized, kind: "auth"},
		{name: "server", statusCode: http.StatusInternalServerError, kind: "server"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": tt.statusCode, "message": http.StatusText(tt.statusCode)})
			}))
			defer server.Close()

			t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
			t.Setenv("NEXTCLOUD_USERNAME", "antonia")
			t.Setenv("NEXTCLOUD_PASSWORD", "pw")

			var stdout, stderr bytes.Buffer
			if code := Main([]string{"board", "list", "--json"}, &stdout, &stderr); code != 1 {
				t.Fatalf("Main() exit = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.HasPrefix(stderr.String(), "error: "+tt.kind+": deck api returned status ") {
				t.Fatalf("stderr = %q, want %s kind", stderr.String(), tt.kind)
			}
		})
	}
}

func TestMainNetworkErrorKind(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := server.URL
	server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", baseURL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if code := Main([]string{"board", "list"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Main() exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.HasPrefix(stderr.String(), "error: network: ") {
		t.Fatalf("stderr = %q, want network error", stderr.String())
	}
}

func TestRunRequiredFlagValidationForCommandFamilies(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "label list missing board", args: []string{"label", "list"}, wantErr: "label list requires --board"},
		{name: "label get missing label", args: []string{"label", "get", "--board", "1"}, wantErr: "label get requires --board --label"},
		{name: "label create missing title", args: []string{"label", "create", "--board", "1"}, wantErr: "label create requires --board --title"},
		{name: "label update missing label", args: []string{"label", "update", "--board", "1", "--title", "x"}, wantErr: "label update requires --board --label"},
		{name: "label delete missing label", args: []string{"label", "delete", "--board", "1"}, wantErr: "label delete requires --board --label"},

		{name: "comment list missing card", args: []string{"comment", "list"}, wantErr: "comment list requires --card"},
		{name: "comment create missing message", args: []string{"comment", "create", "--card", "1"}, wantErr: "comment create requires --card --message"},
		{name: "comment update missing comment", args: []string{"comment", "update", "--card", "1", "--message", "x"}, wantErr: "comment update requires --card --comment --message"},
		{name: "comment delete missing comment", args: []string{"comment", "delete", "--card", "1"}, wantErr: "comment delete requires --card --comment"},

		{name: "attachment list missing card", args: []string{"attachment", "list", "--board", "1", "--stack", "2"}, wantErr: "attachment list requires --board --stack --card"},
		{name: "attachment upload missing file", args: []string{"attachment", "upload", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "attachment upload requires --board --stack --card --file"},
		{name: "attachment download missing out", args: []string{"attachment", "download", "--board", "1", "--stack", "2", "--card", "3", "--attachment", "4"}, wantErr: "attachment download requires --board --stack --card --attachment --out"},
		{name: "attachment delete missing attachment", args: []string{"attachment", "delete", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "attachment delete requires --board --stack --card --attachment"},
		{name: "attachment restore missing attachment", args: []string{"attachment", "restore", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "attachment restore requires --board --stack --card --attachment"},

		{name: "share list missing board", args: []string{"share", "list"}, wantErr: "share list requires --board"},
		{name: "share create missing participant", args: []string{"share", "create", "--board", "1"}, wantErr: "share create requires --board --participant"},
		{name: "share update missing share", args: []string{"share", "update", "--board", "1"}, wantErr: "share update requires --board --share-id"},
		{name: "share delete missing share", args: []string{"share", "delete", "--board", "1"}, wantErr: "share delete requires --board --share-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInvalidCommandDoesNotCallAPI(t, tt.args, tt.wantErr)
		})
	}
}

func TestRunRequiredFlagValidationForRemainingWeakPaths(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "board clone missing board", args: []string{"board", "clone"}, wantErr: "board clone requires --board"},
		{name: "card reorder missing order", args: []string{"card", "reorder", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "card reorder requires --board --stack --card --order"},
		{name: "card due get missing ids", args: []string{"card", "due", "get", "--board", "1", "--stack", "2"}, wantErr: "card due get requires --board --stack --card"},
		{name: "card due set missing value", args: []string{"card", "due", "set", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "card due set requires --value"},
		{name: "card create conflicting text sources", args: []string{"card", "create", "--board", "1", "--stack", "2", "--title", "x", "--description", "x", "--description-file", "description.md"}, wantErr: "choose only one text source"},
		{name: "comment create conflicting text sources", args: []string{"comment", "create", "--card", "1", "--message", "x", "--body-stdin"}, wantErr: "choose only one text source"},
		{name: "todo list missing ids", args: []string{"todo", "list", "--board", "1", "--stack", "2"}, wantErr: "todo list requires --board --stack --card"},
		{name: "todo add missing ids", args: []string{"todo", "add", "--text", "x"}, wantErr: "todo add requires --board --stack --card"},
		{name: "todo check missing index", args: []string{"todo", "check", "--board", "1", "--stack", "2", "--card", "3"}, wantErr: "todo check requires --board --stack --card --index"},
		{name: "activity card missing card", args: []string{"activity", "card"}, wantErr: "activity card requires --card"},
		{name: "user search missing term", args: []string{"user", "search"}, wantErr: "user search requires --term"},
		{name: "user get missing user", args: []string{"user", "get"}, wantErr: "user get requires --user"},
		{name: "config set missing key", args: []string{"config", "set", "--value", "true"}, wantErr: "config set requires --key --value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInvalidCommandDoesNotCallAPI(t, tt.args, tt.wantErr)
		})
	}
}

func runInvalidCommandDoesNotCallAPI(t *testing.T, args []string, wantErr string) {
	t.Helper()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	err := Run(args, &stdout, &stderr)
	if err == nil {
		t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("Run(%v) error = %q, want %q", args, err.Error(), wantErr)
	}
	if got := atomic.LoadInt32(&requests); got != 0 {
		t.Fatalf("Run(%v) made %d API requests", args, got)
	}
}

func TestRunHelp(t *testing.T) {
	clearNextcloudEnv(t)

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"help"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(help) error = %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help output")
	}
}

// --- Run basics ---

func TestRun_NoArgs(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_UnknownCommand(t *testing.T) {
	runCLISmoke(t)
}

// --- Board subcommands ---

func TestRun_BoardList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardGet_MissingFlag(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardUpdate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardArchive(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardUnarchive(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardDelete(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardClone(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardExport(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardImport(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardRestore(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardImportSystems(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_BoardImportSchema(t *testing.T) {
	runCLISmoke(t)
}

// --- List subcommands ---

func TestRun_ListList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListArchived(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListRename(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListReorder(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ListDelete(t *testing.T) {
	runCLISmoke(t)
}

// --- Card subcommands ---

func TestRun_CardList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardClone(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDelete(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardMove(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardReorder(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardArchive(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardUnarchive(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDone(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardUndone(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardRename(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDescribe(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDueGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDueSet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardDueClear(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardAssignUser(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardUnassignUser(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardAssignLabel(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CardRemoveLabel(t *testing.T) {
	runCLISmoke(t)
}

// --- Label subcommands ---

func TestRun_LabelList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_LabelGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_LabelCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_LabelUpdate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_LabelDelete(t *testing.T) {
	runCLISmoke(t)
}

// --- Comment subcommands ---

func TestRun_CommentList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CommentCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CommentUpdate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_CommentDelete(t *testing.T) {
	runCLISmoke(t)
}

// --- Attachment subcommands ---

func TestRun_AttachmentList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_AttachmentUpload(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_AttachmentDownload(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_AttachmentDelete(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_AttachmentRestore(t *testing.T) {
	runCLISmoke(t)
}

// --- Share subcommands ---

func TestRun_ShareList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ShareCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ShareUpdate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ShareDelete(t *testing.T) {
	runCLISmoke(t)
}

// --- Config subcommands ---

func TestRun_ConfigGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ConfigSet(t *testing.T) {
	runCLISmoke(t)
}

// --- Other subcommands ---

func TestRun_SearchCards(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_OverviewUpcoming(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_Capabilities(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_UserSearch(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_UserGet(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_ActivityCard(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_SessionCreate(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_SessionSync(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_SessionClose(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_TodoList(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_TodoAdd(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_TodoCheck(t *testing.T) {
	runCLISmoke(t)
}

func TestRun_TodoUncheck(t *testing.T) {
	runCLISmoke(t)
}

func TestRunHelpPathsWithoutCredentials(t *testing.T) {
	clearNextcloudEnv(t)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "root long help", args: []string{"--help"}, want: "deck <command>"},
		{name: "root short help", args: []string{"-h"}, want: "deck <command>"},
		{name: "card help flag", args: []string{"card", "--help"}, want: "deck card list|get|deleted|create"},
		{name: "card short help", args: []string{"card", "-h"}, want: "deck card list|get|deleted|create"},
		{name: "card help subcommand", args: []string{"card", "help"}, want: "deck card list|get|deleted|create"},
		{name: "board help flag", args: []string{"board", "--help"}, want: "deck board list|get|find|create"},
		{name: "board help subcommand", args: []string{"board", "help"}, want: "deck board list|get|find|create"},
		{name: "board short help", args: []string{"board", "-h"}, want: "deck board list|get|find|create"},
		{name: "help board", args: []string{"help", "board"}, want: "deck board list|get|find|create"},
		{name: "board list help flag", args: []string{"board", "list", "--help"}, want: "deck board list"},
		{name: "board list short help", args: []string{"board", "list", "-h"}, want: "deck board list"},
		{name: "nested board list help command", args: []string{"help", "board", "list"}, want: "deck board list"},
		{name: "card due help", args: []string{"card", "due", "--help"}, want: "deck card due get|set|clear"},
		{name: "card due short help", args: []string{"card", "due", "-h"}, want: "deck card due get|set|clear"},
		{name: "card due help subcommand", args: []string{"card", "due", "help"}, want: "deck card due get|set|clear"},
		{name: "nested help command", args: []string{"help", "card", "due"}, want: "deck card due get|set|clear"},
		{name: "auth help flag", args: []string{"auth", "--help"}, want: "deck auth setup"},
		{name: "auth setup help flag", args: []string{"auth", "setup", "--help"}, want: "deck auth setup"},
		{name: "help auth setup", args: []string{"help", "auth", "setup"}, want: "deck auth setup"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), tt.want)
			}
			if strings.Contains(stderr.String(), "missing env") {
				t.Fatalf("stderr = %q, want no credential error", stderr.String())
			}
		})
	}
}

func TestRunUnknownCommandsWithoutCredentials(t *testing.T) {
	clearNextcloudEnv(t)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown root command", args: []string{"bogus"}, want: `unknown command "bogus"`},
		{name: "unknown board command", args: []string{"board", "bogus"}, want: `unknown board command "bogus"`},
		{name: "unknown capabilities command", args: []string{"capabilities", "bogus"}, want: `unknown capabilities command "bogus"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := Run(tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("Run(%v) error = nil, want %q", tt.args, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
			if strings.Contains(err.Error(), "missing env") || strings.Contains(stderr.String(), "missing env") {
				t.Fatalf("err=%v stderr=%q, want no credential error", err, stderr.String())
			}
		})
	}
}

func TestRunMissingSubcommandsWithoutCredentials(t *testing.T) {
	clearNextcloudEnv(t)

	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantErr string
	}{
		{name: "board", args: []string{"board"}, wantOut: "deck board list|get|find|create"},
		{name: "card due", args: []string{"card", "due"}, wantErr: "card due requires get, set, or clear"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := Run(tt.args, &stdout, &stderr)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Run(%v) error = nil, want %q", tt.args, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
				}
			}
			if tt.wantOut != "" && !strings.Contains(stdout.String(), tt.wantOut) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), tt.wantOut)
			}
			if strings.Contains(errString(err), "missing env") || strings.Contains(stderr.String(), "missing env") {
				t.Fatalf("err=%v stderr=%q, want no credential error", err, stderr.String())
			}
		})
	}
}

func clearNextcloudEnv(t *testing.T) {
	t.Helper()
	t.Setenv("NEXTCLOUD_BASE_URL", "")
	t.Setenv("NEXTCLOUD_USERNAME", "")
	t.Setenv("NEXTCLOUD_PASSWORD", "")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "")
	t.Setenv("DECK_PROFILE", "")
}

func defaultTestConfigPath(t *testing.T) string {
	t.Helper()
	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}
	return filepath.Join(dir, "nextcloud-deck-cli", "config.json")
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func withCommandStdin(t *testing.T, input string) {
	t.Helper()
	old := commandStdin
	commandStdin = strings.NewReader(input)
	t.Cleanup(func() {
		commandStdin = old
	})
}

type cliSmokeCase struct {
	args     []string
	want     []string
	setup    func(t *testing.T, tc *cliSmokeCase)
	validate func(t *testing.T)
}

func runCLISmoke(t *testing.T) {
	t.Helper()
	cases := cliSmokeCases()
	tc, ok := cases[t.Name()]
	if !ok {
		t.Fatalf("missing CLI smoke case for %s", t.Name())
	}
	if tc.setup != nil {
		tc.setup(t, &tc)
	}
	if len(tc.want) == 0 {
		if tc.args != nil {
			runCLIWithoutServer(t, tc.args)
		}
		if tc.validate != nil {
			tc.validate(t)
		}
		return
	}
	seen := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		seen[key]++
		writeCLIMockResponse(t, w, r)
	}))
	defer server.Close()
	setCLIEnv(t, server.URL)
	var stdout, stderr bytes.Buffer
	if err := Run(tc.args, &stdout, &stderr); err != nil {
		t.Fatalf("Run(%v) error = %v; stdout=%s stderr=%s", tc.args, err, stdout.String(), stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatalf("Run(%v) produced empty stdout", tc.args)
	}
	for _, want := range tc.want {
		if seen[want] == 0 {
			t.Fatalf("Run(%v) did not call %s; seen=%#v", tc.args, want, seen)
		}
	}
	if tc.validate != nil {
		tc.validate(t)
	}
}

func cliSmokeCases() map[string]cliSmokeCase {
	return map[string]cliSmokeCase{
		"TestRun_NoArgs":         {args: []string{}, validate: wantStdoutContains("deck <command>")},
		"TestRun_UnknownCommand": {args: []string{"bogus"}, validate: wantRunError(`unknown command "bogus"`)},
		"TestRun_BoardList":      {args: []string{"board", "list", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards"}},
		"TestRun_BoardGet":       {args: []string{"board", "get", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_BoardGet_MissingFlag": {validate: func(t *testing.T) {
			runInvalidCommandDoesNotCallAPI(t, []string{"board", "get"}, "board get requires --board")
		}},
		"TestRun_BoardUpdate":           {args: []string{"board", "update", "--board", "7", "--title", "Updated", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7", "PUT /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_BoardArchive":          {args: []string{"board", "archive", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7", "PUT /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_BoardUnarchive":        {args: []string{"board", "unarchive", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7", "PUT /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_BoardDelete":           {args: []string{"board", "delete", "--board", "7", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_BoardClone":            {args: []string{"board", "clone", "--board", "7", "--with-cards", "--json"}, want: []string{"POST /index.php/apps/deck/boards/7/clone"}},
		"TestRun_BoardExport":           exportBoardSmokeCase(),
		"TestRun_BoardImport":           importBoardSmokeCase(),
		"TestRun_BoardImportServer":     importServerBoardSmokeCase(),
		"TestRun_BoardRestore":          {args: []string{"board", "restore", "--board", "7", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/undo_delete"}},
		"TestRun_BoardImportSystems":    {args: []string{"board", "import-systems", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/boards/import/getSystems"}},
		"TestRun_BoardImportSchema":     {args: []string{"board", "import-schema", "--name", "deck", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/boards/import/config/schema/deck"}},
		"TestRun_ListList":              {args: []string{"list", "list", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks"}},
		"TestRun_ListArchived":          {args: []string{"list", "archived", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/archived"}},
		"TestRun_ListGet":               {args: []string{"list", "get", "--board", "7", "--list", "2", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2"}},
		"TestRun_ListCreate":            {args: []string{"list", "create", "--board", "7", "--title", "Doing", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/stacks"}},
		"TestRun_ListRename":            {args: []string{"list", "rename", "--board", "7", "--list", "2", "--title", "Renamed", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2"}},
		"TestRun_ListReorder":           {args: []string{"list", "reorder", "--board", "7", "--list", "2", "--order", "1", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2"}},
		"TestRun_ListDone":              {args: []string{"list", "done", "--board", "7", "--list", "2", "--json"}, want: []string{"PUT /ocs/v2.php/apps/deck/api/v1.0/stacks/2/done"}},
		"TestRun_ListUndone":            {args: []string{"list", "undone", "--board", "7", "--list", "2", "--json"}, want: []string{"PUT /ocs/v2.php/apps/deck/api/v1.0/stacks/2/done"}},
		"TestRun_ListDelete":            {args: []string{"list", "delete", "--board", "7", "--list", "2", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7/stacks/2"}},
		"TestRun_CardList":              {args: []string{"card", "list", "--board", "7", "--stack", "2", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2"}},
		"TestRun_CardGet":               {args: []string{"card", "get", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardDeleted":           {args: []string{"card", "deleted", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/7/cards/deleted"}},
		"TestRun_CardCreate":            {args: []string{"card", "create", "--board", "7", "--stack", "2", "--title", "Card", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards"}},
		"TestRun_CardClone":             {args: []string{"card", "clone", "--card", "9", "--to-stack", "2", "--json"}, want: []string{"GET /index.php/apps/deck/cards/9", "POST /ocs/v2.php/apps/deck/api/v1.0/cards/9/clone"}},
		"TestRun_CardDelete":            {args: []string{"card", "delete", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardMove":              {args: []string{"card", "move", "--board", "7", "--from-stack", "2", "--to-stack", "3", "--card", "9", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/reorder"}},
		"TestRun_CardReorder":           {args: []string{"card", "reorder", "--board", "7", "--stack", "2", "--card", "9", "--order", "1", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/reorder"}},
		"TestRun_CardArchive":           {args: []string{"card", "archive", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/archive"}},
		"TestRun_CardUnarchive":         {args: []string{"card", "unarchive", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/unarchive"}},
		"TestRun_CardDone":              {args: []string{"card", "done", "--card", "9", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/done"}},
		"TestRun_CardUndone":            {args: []string{"card", "undone", "--card", "9", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/undone"}},
		"TestRun_CardRename":            {args: []string{"card", "rename", "--board", "7", "--stack", "2", "--card", "9", "--title", "New", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardDescribe":          {args: []string{"card", "describe", "--board", "7", "--stack", "2", "--card", "9", "--description", "Desc", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardUpdate":            {args: []string{"card", "update", "--board", "7", "--stack", "2", "--card", "9", "--type", "text", "--color", "00ff00", "--start", "2026-05-20", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardDueGet":            {args: []string{"card", "due", "get", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardDueSet":            {args: []string{"card", "due", "set", "--board", "7", "--stack", "2", "--card", "9", "--value", "2026-05-09", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardDueClear":          {args: []string{"card", "due", "clear", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_CardAssignUser":        {args: []string{"card", "assign-user", "--board", "7", "--stack", "2", "--card", "9", "--user", "alice", "--json"}, want: []string{"POST /index.php/apps/deck/cards/9/assign"}},
		"TestRun_CardUnassignUser":      {args: []string{"card", "unassign-user", "--board", "7", "--stack", "2", "--card", "9", "--user", "alice", "--json"}, want: []string{"PUT /index.php/apps/deck/cards/9/unassign"}},
		"TestRun_CardAssignLabel":       {args: []string{"card", "assign-label", "--board", "7", "--stack", "2", "--card", "9", "--label", "4", "--json"}, want: []string{"PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/assignLabel"}},
		"TestRun_CardRemoveLabel":       {args: []string{"card", "remove-label", "--board", "7", "--stack", "2", "--card", "9", "--label", "4", "--json"}, want: []string{"PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/removeLabel"}},
		"TestRun_CardAssignDependent":   {args: []string{"card", "assign-dependent", "--board", "7", "--stack", "2", "--card", "9", "--dependent-card", "10", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/dependentCards/10"}},
		"TestRun_CardRemoveDependent":   {args: []string{"card", "remove-dependent", "--board", "7", "--stack", "2", "--card", "9", "--dependent-card", "10", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/dependentCards/10"}},
		"TestRun_LabelList":             {args: []string{"label", "list", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_LabelGet":              {args: []string{"label", "get", "--board", "7", "--label", "4", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/labels/4"}},
		"TestRun_LabelCreate":           {args: []string{"label", "create", "--board", "7", "--title", "Bug", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/labels"}},
		"TestRun_LabelUpdate":           {args: []string{"label", "update", "--board", "7", "--label", "4", "--title", "New", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/labels/4", "PUT /index.php/apps/deck/api/v1.0/boards/7/labels/4"}},
		"TestRun_LabelDelete":           {args: []string{"label", "delete", "--board", "7", "--label", "4", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7/labels/4"}},
		"TestRun_CommentList":           {args: []string{"comment", "list", "--card", "9", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/cards/9/comments"}},
		"TestRun_CommentCreate":         {args: []string{"comment", "create", "--card", "9", "--message", "Hi", "--json"}, want: []string{"POST /ocs/v2.php/apps/deck/api/v1.0/cards/9/comments"}},
		"TestRun_CommentReply":          {args: []string{"comment", "create", "--card", "9", "--reply-to", "6", "--message", "Reply", "--json"}, want: []string{"POST /ocs/v2.php/apps/deck/api/v1.0/cards/9/comments"}},
		"TestRun_CommentUpdate":         {args: []string{"comment", "update", "--card", "9", "--comment", "6", "--message", "Hi", "--json"}, want: []string{"PUT /ocs/v2.php/apps/deck/api/v1.0/cards/9/comments/6"}},
		"TestRun_CommentDelete":         {args: []string{"comment", "delete", "--card", "9", "--comment", "6", "--json"}, want: []string{"DELETE /ocs/v2.php/apps/deck/api/v1.0/cards/9/comments/6"}},
		"TestRun_AttachmentList":        {args: []string{"attachment", "list", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"GET /index.php/apps/deck/cards/9/attachments"}},
		"TestRun_AttachmentUpload":      uploadAttachmentSmokeCase(),
		"TestRun_AttachmentDownload":    downloadAttachmentSmokeCase(),
		"TestRun_AttachmentDelete":      {args: []string{"attachment", "delete", "--board", "7", "--stack", "2", "--card", "9", "--attachment", "8", "--json"}, want: []string{"GET /index.php/apps/deck/cards/9/attachments", "DELETE /index.php/apps/deck/cards/9/attachment/deck_file:8"}},
		"TestRun_AttachmentTypedDelete": {args: []string{"attachment", "delete", "--board", "7", "--stack", "2", "--card", "9", "--attachment", "8", "--type", "deck_file", "--json"}, want: []string{"DELETE /ocs/v2.php/apps/deck/api/v1.0/cards/9/attachments/deck_file:8"}},
		"TestRun_AttachmentRestore":     {args: []string{"attachment", "restore", "--board", "7", "--stack", "2", "--card", "9", "--attachment", "8", "--json"}, want: []string{"GET /index.php/apps/deck/cards/9/attachments", "GET /index.php/apps/deck/cards/9/attachment/deck_file:8/restore"}},
		"TestRun_ShareList":             {args: []string{"share", "list", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7"}},
		"TestRun_SharePermissions":      {args: []string{"share", "permissions", "--board", "7", "--json"}, want: []string{"GET /index.php/apps/deck/boards/7/permissions"}},
		"TestRun_ShareCreate":           {args: []string{"share", "create", "--board", "7", "--participant", "alice", "--json"}, want: []string{"POST /index.php/apps/deck/api/v1.0/boards/7/acl"}},
		"TestRun_ShareUpdate":           {args: []string{"share", "update", "--board", "7", "--share-id", "3", "--json"}, want: []string{"PUT /index.php/apps/deck/api/v1.0/boards/7/acl/3"}},
		"TestRun_ShareDelete":           {args: []string{"share", "delete", "--board", "7", "--share-id", "3", "--json"}, want: []string{"DELETE /index.php/apps/deck/api/v1.0/boards/7/acl/3"}},
		"TestRun_ShareLeave":            {args: []string{"share", "leave", "--board", "7", "--json"}, want: []string{"POST /index.php/apps/deck/boards/7/leave"}},
		"TestRun_ShareTransferOwner":    {args: []string{"share", "transfer-owner", "--board", "7", "--new-owner", "alice", "--json"}, want: []string{"PUT /index.php/apps/deck/boards/7/transferOwner"}},
		"TestRun_ConfigGet":             {args: []string{"config", "get", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/config"}},
		"TestRun_ConfigSet":             {args: []string{"config", "set", "--key", "calendar", "--value", "true", "--json"}, want: []string{"POST /ocs/v2.php/apps/deck/api/v1.0/config/calendar"}},
		"TestRun_SearchCards":           {args: []string{"search", "cards", "--term", "hello", "--cursor", "5", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/search"}},
		"TestRun_OverviewUpcoming":      {args: []string{"overview", "upcoming", "--json"}, want: []string{"GET /ocs/v2.php/apps/deck/api/v1.0/overview/upcoming"}},
		"TestRun_Capabilities":          {args: []string{"capabilities", "--json"}, want: []string{"GET /ocs/v2.php/cloud/capabilities"}},
		"TestRun_UserSearch":            {args: []string{"user", "search", "--term", "alice", "--json"}, want: []string{"GET /ocs/v2.php/apps/files_sharing/api/v1/sharees"}},
		"TestRun_UserGet":               {args: []string{"user", "get", "--user", "alice", "--json"}, want: []string{"GET /ocs/v2.php/cloud/users/alice"}},
		"TestRun_ActivityCard":          {args: []string{"activity", "card", "--card", "9", "--json"}, want: []string{"GET /ocs/v2.php/apps/activity/api/v2/activity/filter"}},
		"TestRun_ActivityList":          {args: []string{"activity", "list", "--object-type", "deck_card", "--object-id", "9", "--limit", "5", "--json"}, want: []string{"GET /ocs/v2.php/apps/activity/api/v2/activity/filter"}},
		"TestRun_SessionCreate":         {args: []string{"session", "create", "--board", "7", "--json"}, want: []string{"PUT /ocs/v2.php/apps/deck/api/v1.0/session/create"}},
		"TestRun_SessionSync":           {args: []string{"session", "sync", "--board", "7", "--token", "tok", "--json"}, want: []string{"POST /ocs/v2.php/apps/deck/api/v1.0/session/sync"}},
		"TestRun_SessionClose":          {args: []string{"session", "close", "--board", "7", "--token", "tok", "--json"}, want: []string{"POST /ocs/v2.php/apps/deck/api/v1.0/session/close"}},
		"TestRun_TodoList":              {args: []string{"todo", "list", "--board", "7", "--stack", "2", "--card", "9", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_TodoAdd":               {args: []string{"todo", "add", "--board", "7", "--stack", "2", "--card", "9", "--text", "new", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_TodoCheck":             {args: []string{"todo", "check", "--board", "7", "--stack", "2", "--card", "9", "--index", "1", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
		"TestRun_TodoUncheck":           {args: []string{"todo", "uncheck", "--board", "7", "--stack", "2", "--card", "9", "--index", "2", "--json"}, want: []string{"GET /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", "PUT /index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9"}},
	}
}

func importServerBoardSmokeCase() cliSmokeCase {
	return cliSmokeCase{
		setup: func(t *testing.T, tc *cliSmokeCase) {
			dir := t.TempDir()
			config := filepath.Join(dir, "config.json")
			data := filepath.Join(dir, "data.json")
			if err := os.WriteFile(config, []byte(`{"owner":"alice"}`), 0o600); err != nil {
				t.Fatalf("write config file: %v", err)
			}
			if err := os.WriteFile(data, []byte(`{"title":"Imported"}`), 0o600); err != nil {
				t.Fatalf("write data file: %v", err)
			}
			tc.args = []string{"board", "import-server", "--system", "DeckJson", "--config-file", config, "--data-file", data, "--json"}
			tc.want = []string{"POST /ocs/v2.php/apps/deck/api/v1.0/boards/import"}
		},
	}
}

func setCLIEnv(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("NEXTCLOUD_BASE_URL", baseURL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
}

func runCLIWithoutServer(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	clearNextcloudEnv(t)
	var stdout, stderr bytes.Buffer
	err := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

func wantStdoutContains(want string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		stdout, stderr, err := runCLIWithoutServer(t, cliSmokeCases()[t.Name()].args)
		if err != nil {
			t.Fatalf("Run() error = %v; stderr=%s", err, stderr)
		}
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want substring %q", stdout, want)
		}
	}
}

func wantRunError(want string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		stdout, stderr, err := runCLIWithoutServer(t, cliSmokeCases()[t.Name()].args)
		if err == nil {
			t.Fatalf("Run() succeeded; stdout=%s stderr=%s", stdout, stderr)
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("err = %q, want substring %q", err.Error(), want)
		}
	}
}

func exportBoardSmokeCase() cliSmokeCase {
	var out string
	return cliSmokeCase{
		setup: func(t *testing.T, tc *cliSmokeCase) {
			out = t.TempDir() + "/board.json"
			tc.args = []string{"board", "export", "--board", "7", "--out", out, "--json"}
			tc.want = []string{"GET /index.php/apps/deck/boards/7/export"}
		},
		validate: func(t *testing.T) {
			if data, err := os.ReadFile(out); err != nil || string(data) == "" {
				t.Fatalf("exported data = %q err=%v", string(data), err)
			}
		},
	}
}

func importBoardSmokeCase() cliSmokeCase {
	return cliSmokeCase{setup: func(t *testing.T, tc *cliSmokeCase) {
		file := t.TempDir() + "/board.json"
		if err := os.WriteFile(file, []byte(`{"title":"Board"}`), 0o600); err != nil {
			t.Fatalf("write import file: %v", err)
		}
		tc.args = []string{"board", "import", "--file", file, "--json"}
		tc.want = []string{"POST /index.php/apps/deck/boards/import"}
	}}
}

func uploadAttachmentSmokeCase() cliSmokeCase {
	return cliSmokeCase{setup: func(t *testing.T, tc *cliSmokeCase) {
		file := t.TempDir() + "/note.txt"
		if err := os.WriteFile(file, []byte("note"), 0o600); err != nil {
			t.Fatalf("write upload file: %v", err)
		}
		tc.args = []string{"attachment", "upload", "--board", "7", "--stack", "2", "--card", "9", "--file", file, "--json"}
		tc.want = []string{"POST /index.php/apps/deck/cards/9/attachment"}
	}}
}

func downloadAttachmentSmokeCase() cliSmokeCase {
	var out string
	return cliSmokeCase{
		setup: func(t *testing.T, tc *cliSmokeCase) {
			out = t.TempDir() + "/download.txt"
			tc.args = []string{"attachment", "download", "--board", "7", "--stack", "2", "--card", "9", "--attachment", "8", "--out", out, "--json"}
			tc.want = []string{"GET /index.php/apps/deck/cards/9/attachments", "GET /index.php/apps/deck/cards/9/attachment/deck_file:8"}
		},
		validate: func(t *testing.T) {
			if data, err := os.ReadFile(out); err != nil || string(data) != "download" {
				t.Fatalf("download data = %q err=%v", string(data), err)
			}
		},
	}
}

func writeCLIMockResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	path := r.URL.Path
	if strings.HasPrefix(path, "/ocs/v2.php/") {
		writeCLIOCSResponse(w, cliOCSDataForPath(r.Method, path))
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.Contains(path, "/dependentCards/") {
		writeCLIJSONResponse(w, map[string]any{"id": 9, "title": "Card", "stackId": 2, "dependentCards": []int64{10}})
		return
	}
	if strings.HasSuffix(path, "/export") {
		_, _ = w.Write([]byte(`{"title":"Board"}`))
		return
	}
	if strings.Contains(path, "/attachment/deck_file:8") && !strings.HasSuffix(path, "/restore") {
		_, _ = w.Write([]byte("download"))
		return
	}
	if r.Method == http.MethodPost && strings.HasSuffix(path, "/stacks") {
		writeCLIJSONResponse(w, map[string]any{"id": 2, "boardId": 7, "title": "Doing", "order": 1})
		return
	}
	writeCLIJSONResponse(w, cliDataForPathWithMethod(r.Method, path))
}

func cliOCSDataForPath(method, path string) any {
	switch {
	case strings.Contains(path, "/comments"):
		if method != http.MethodGet || strings.Contains(path, "/comments/6") {
			return map[string]any{"id": 6, "objectId": 9, "message": "Hi"}
		}
		return []map[string]any{{"id": 6, "objectId": 9, "message": "Hi"}}
	case strings.Contains(path, "/session/create"):
		return map[string]any{"token": "tok"}
	case strings.Contains(path, "/session/"):
		return []any{}
	case strings.Contains(path, "/search"), strings.Contains(path, "/overview/upcoming"):
		return []map[string]any{{"id": 9, "title": "Card", "stackId": 2}}
	case strings.Contains(path, "/activity/"):
		return []map[string]any{{"activity_id": 1, "subject": "updated"}}
	case strings.Contains(path, "/boards/import/getSystems"):
		return []string{"deck"}
	case strings.Contains(path, "/boards/import/config/schema"):
		return map[string]any{"type": "object"}
	default:
		return map[string]any{"calendar": true, "users": []any{map[string]any{"label": "alice"}}, "uid": "alice"}
	}
}

func cliDataForPathWithMethod(method, path string) any {
	switch {
	case strings.HasSuffix(path, "/boards"):
		if strings.Contains(path, "/api/v1.0/boards") {
			return []map[string]any{{"id": 7, "title": "Board", "color": "ff0000", "archived": false}}
		}
	case strings.Contains(path, "/boards/import"):
		return map[string]any{"id": 7, "title": "Imported", "color": "ff0000"}
	case strings.Contains(path, "/boards/7/permissions"):
		return map[string]any{"PERMISSION_READ": true, "PERMISSION_EDIT": true, "PERMISSION_MANAGE": true, "PERMISSION_SHARE": true}
	case strings.Contains(path, "/boards/7/stacks/2/cards/9"):
		return map[string]any{"id": 9, "title": "Card", "description": "- [ ] first\n- [x] second", "stackId": 2, "type": "plain", "order": 1, "archived": false}
	case strings.Contains(path, "/boards/7/stacks/2/cards"):
		return map[string]any{"id": 9, "title": "Card", "stackId": 2, "type": "plain", "order": 1, "archived": false}
	case strings.Contains(path, "/boards/7/stacks/2"):
		return map[string]any{"id": 2, "boardId": 7, "title": "Doing", "order": 1, "cards": []map[string]any{{"id": 9, "title": "Card", "stackId": 2}}}
	case strings.Contains(path, "/boards/7/stacks"):
		return []map[string]any{{"id": 2, "boardId": 7, "title": "Doing", "order": 1}}
	case strings.Contains(path, "/labels/4"):
		return map[string]any{"id": 4, "title": "Bug", "color": "31CC7C"}
	case strings.Contains(path, "/labels"):
		return map[string]any{"id": 4, "title": "Bug", "color": "31CC7C"}
	case strings.Contains(path, "/acl"):
		return []map[string]any{{"id": 3, "type": 0, "permissionEdit": true, "permissionShare": false, "permissionManage": false, "participant": map[string]any{"uid": "alice"}}}
	case strings.Contains(path, "/cards/9/attachments/deck_file:8"):
		return map[string]any{"id": 8, "cardId": 9, "type": "deck_file", "data": "note.txt"}
	case strings.Contains(path, "/cards/9/attachments"):
		return []map[string]any{{"id": 8, "cardId": 9, "type": "", "data": "note.txt"}}
	case strings.Contains(path, "/cards/9/attachment"):
		return map[string]any{"id": 8, "cardId": 9, "type": "deck_file", "data": "note.txt"}
	case strings.Contains(path, "/cards/9/assign"):
		return map[string]any{"id": 5, "participant": map[string]any{"uid": "alice"}}
	case strings.Contains(path, "/cards/9"):
		return map[string]any{"id": 9, "title": "Card", "description": "- [ ] first\n- [x] second", "stackId": 2, "type": "plain", "order": 1, "archived": false}
	case strings.Contains(path, "/boards/7/clone"):
		return map[string]any{"id": 17, "title": "Board copy", "color": "ff0000"}
	case strings.Contains(path, "/boards/7"):
		return map[string]any{"id": 7, "title": "Board", "color": "ff0000", "archived": false, "labels": []map[string]any{{"id": 4, "title": "Bug", "color": "31CC7C"}}, "acl": []map[string]any{{"id": 3, "type": 0, "permissionEdit": true, "permissionShare": false, "permissionManage": false, "participant": map[string]any{"uid": "alice"}}}}
	}
	return map[string]any{"id": 1, "title": "ok"}
}

func writeCLIJSONResponse(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(data)
}

func writeCLIOCSResponse(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"ocs": map[string]any{"meta": map[string]any{"status": "ok", "statuscode": 200}, "data": data}})
}
