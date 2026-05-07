package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"help"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(help) error = %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help output")
	}
}
