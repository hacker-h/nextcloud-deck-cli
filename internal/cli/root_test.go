package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
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
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"help"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(help) error = %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help output")
	}
}
