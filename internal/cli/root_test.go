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
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
	t.Skip("TODO: implement")
}

func TestRun_UnknownCommand(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Board subcommands ---

func TestRun_BoardList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardGet_MissingFlag(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardArchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardUnarchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardClone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardExport(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImport(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardRestore(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImportSystems(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImportSchema(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- List subcommands ---

func TestRun_ListList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListArchived(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListRename(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListReorder(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Card subcommands ---

func TestRun_CardList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardClone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardMove(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardReorder(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardArchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUnarchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUndone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardRename(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDescribe(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueSet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueClear(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardAssignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUnassignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardAssignLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardRemoveLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Label subcommands ---

func TestRun_LabelList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Comment subcommands ---

func TestRun_CommentList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Attachment subcommands ---

func TestRun_AttachmentList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentUpload(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentDownload(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentRestore(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Share subcommands ---

func TestRun_ShareList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Config subcommands ---

func TestRun_ConfigGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ConfigSet(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Other subcommands ---

func TestRun_SearchCards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_OverviewUpcoming(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_Capabilities(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_UserSearch(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_UserGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ActivityCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionSync(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionClose(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoAdd(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoCheck(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoUncheck(t *testing.T) {
	t.Skip("TODO: implement")
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
		{name: "card help flag", args: []string{"card", "--help"}, want: "deck card list|get|create"},
		{name: "card short help", args: []string{"card", "-h"}, want: "deck card list|get|create"},
		{name: "card help subcommand", args: []string{"card", "help"}, want: "deck card list|get|create"},
		{name: "board help flag", args: []string{"board", "--help"}, want: "deck board list|get|create"},
		{name: "board help subcommand", args: []string{"board", "help"}, want: "deck board list|get|create"},
		{name: "board short help", args: []string{"board", "-h"}, want: "deck board list|get|create"},
		{name: "help board", args: []string{"help", "board"}, want: "deck board list|get|create"},
		{name: "board list help flag", args: []string{"board", "list", "--help"}, want: "deck board list"},
		{name: "board list short help", args: []string{"board", "list", "-h"}, want: "deck board list"},
		{name: "nested board list help command", args: []string{"help", "board", "list"}, want: "deck board list"},
		{name: "card due help", args: []string{"card", "due", "--help"}, want: "deck card due get|set|clear"},
		{name: "card due short help", args: []string{"card", "due", "-h"}, want: "deck card due get|set|clear"},
		{name: "card due help subcommand", args: []string{"card", "due", "help"}, want: "deck card due get|set|clear"},
		{name: "nested help command", args: []string{"help", "card", "due"}, want: "deck card due get|set|clear"},
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
		{name: "board", args: []string{"board"}, wantOut: "deck board list|get|create"},
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
