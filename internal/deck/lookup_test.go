package deck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestFindBoardByTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":1,"title":"Project","color":"ff0000","archived":false},
			{"id":2,"title":"project","color":"00ff00","archived":false},
			{"id":3,"title":"Duplicate","color":"0000ff","archived":false},
			{"id":4,"title":"Duplicate","color":"000000","archived":false}
		]`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	board, err := client.FindBoardByTitle(context.Background(), "Project")
	if err != nil {
		t.Fatalf("FindBoardByTitle() error = %v", err)
	}
	if board.ID != 1 {
		t.Fatalf("board.ID = %d, want 1", board.ID)
	}

	_, err = client.FindBoardByTitle(context.Background(), "Missing")
	assertLookupError(t, err, `board title "Missing" not found`)
	_, err = client.FindBoardByTitle(context.Background(), "PROJECT")
	assertLookupError(t, err, `board title "PROJECT" not found`)
	_, err = client.FindBoardByTitle(context.Background(), "Duplicate")
	assertLookupError(t, err, `board title "Duplicate" matched 2 boards; use id`)
}

func TestFindStackByTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/7/stacks" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":10,"title":"Backlog","boardId":7,"order":1},
			{"id":11,"title":"Doing","boardId":7,"order":2},
			{"id":12,"title":"Done","boardId":7,"order":3},
			{"id":13,"title":"Done","boardId":7,"order":4}
		]`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	stack, err := client.FindStackByTitle(context.Background(), 7, "Backlog")
	if err != nil {
		t.Fatalf("FindStackByTitle() error = %v", err)
	}
	if stack.ID != 10 {
		t.Fatalf("stack.ID = %d, want 10", stack.ID)
	}

	_, err = client.FindStackByTitle(context.Background(), 7, "backlog")
	assertLookupError(t, err, `stack title "backlog" not found on board 7`)
	_, err = client.FindStackByTitle(context.Background(), 7, "Done")
	assertLookupError(t, err, `stack title "Done" matched 2 stacks on board 7; use id`)
}

func TestFindLabelByTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/7" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":7,
			"title":"Project",
			"color":"ff0000",
			"archived":false,
			"labels":[
				{"id":21,"title":"Bug","color":"ff0000","boardId":7},
				{"id":22,"title":"Feature","color":"00ff00","boardId":7},
				{"id":23,"title":"Duplicate","color":"0000ff","boardId":7},
				{"id":24,"title":"Duplicate","color":"000000","boardId":7}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	label, err := client.FindLabelByTitle(context.Background(), 7, "Bug")
	if err != nil {
		t.Fatalf("FindLabelByTitle() error = %v", err)
	}
	if label.ID != 21 {
		t.Fatalf("label.ID = %d, want 21", label.ID)
	}

	_, err = client.FindLabelByTitle(context.Background(), 7, "bug")
	assertLookupError(t, err, `label title "bug" not found on board 7`)
	_, err = client.FindLabelByTitle(context.Background(), 7, "Duplicate")
	assertLookupError(t, err, `label title "Duplicate" matched 2 labels on board 7; use id`)
}

func assertLookupError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want %q", want)
	}
	var lookupErr LookupError
	if !errors.As(err, &lookupErr) {
		t.Fatalf("error type = %T, want LookupError", err)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}
