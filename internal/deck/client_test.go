package deck

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestCreateCardRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/1/stacks/2/cards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("OCS-APIRequest"); got != "true" {
			t.Fatalf("OCS-APIRequest = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Basic "+base64.StdEncoding.EncodeToString([]byte("antonia:pw")) {
			t.Fatalf("Authorization = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["title"] != "Test card" {
			t.Fatalf("title = %#v", payload["title"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":9,"title":"Test card","stackId":2,"order":999,"archived":false}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	card, err := client.CreateCard(context.Background(), 1, 2, CreateCardRequest{Title: "Test card", Order: 999})
	if err != nil {
		t.Fatalf("CreateCard() error = %v", err)
	}
	if card.ID != 9 {
		t.Fatalf("card.ID = %d", card.ID)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"message":"title must be provided"}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	_, err := client.GetBoards(context.Background(), false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "title must be provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}
