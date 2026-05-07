package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if err := Run([]string{"board", "create", "--title", "Test Board"}, &stdout, &stderr); err != nil {
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
