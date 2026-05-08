package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunFindCommandsReturnJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/index.php/apps/deck/api/v1.0/boards":
			_, _ = w.Write([]byte(`[{"id":7,"title":"Project","color":"ff0000","archived":false}]`))
		case "/index.php/apps/deck/api/v1.0/boards/7/stacks":
			_, _ = w.Write([]byte(`[{"id":8,"title":"Backlog","boardId":7,"order":1}]`))
		case "/index.php/apps/deck/api/v1.0/boards/7":
			_, _ = w.Write([]byte(`{"id":7,"title":"Project","color":"ff0000","archived":false,"labels":[{"id":9,"title":"Bug","color":"31CC7C","boardId":7}]}`))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	tests := []struct {
		name string
		args []string
		id   float64
	}{
		{name: "board", args: []string{"board", "find", "--title", "Project", "--json"}, id: 7},
		{name: "list", args: []string{"list", "find", "--board", "7", "--title", "Backlog", "--json"}, id: 8},
		{name: "label", args: []string{"label", "find", "--board", "7", "--title", "Bug", "--json"}, id: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			var payload map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
			}
			if payload["id"] != tt.id {
				t.Fatalf("id = %#v, want %v", payload["id"], tt.id)
			}
		})
	}
}

func TestRunFindReportsValidationErrors(t *testing.T) {
	runInvalidCommandDoesNotCallAPI(t, []string{"board", "find"}, "board find requires --title")
	runInvalidCommandDoesNotCallAPI(t, []string{"list", "find", "--title", "Backlog"}, "list find requires --board --title")
	runInvalidCommandDoesNotCallAPI(t, []string{"label", "find", "--board", "7"}, "label find requires --board --title")
}

func TestMainFindNotFoundClassifiesAsValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"title":"Project","color":"ff0000","archived":false}]`))
	}))
	defer server.Close()

	setNextcloudEnv(t, server.URL)

	var stdout, stderr bytes.Buffer
	if code := Main([]string{"board", "find", "--title", "Missing"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Main() exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, `error: validation: board title "Missing" not found`) {
		t.Fatalf("stderr = %q", got)
	}
}

func TestRunFindHelpPathsWithoutCredentials(t *testing.T) {
	clearNextcloudEnv(t)

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"help", "board", "find"}, want: "deck board find --title TEXT"},
		{args: []string{"help", "list", "find"}, want: "deck list find --board ID --title TEXT"},
		{args: []string{"help", "label", "find"}, want: "deck label find --board ID --title TEXT"},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("Run(%v) error = %v; stderr=%s", tt.args, err, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), tt.want)
			}
		})
	}
}

func setNextcloudEnv(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("NEXTCLOUD_BASE_URL", baseURL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
}
