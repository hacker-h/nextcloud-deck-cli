package deck

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestIntegrationGetBoards(t *testing.T) {
	if os.Getenv("NEXTCLOUD_BASE_URL") == "" || os.Getenv("NEXTCLOUD_USERNAME") == "" || (os.Getenv("NEXTCLOUD_PASSWORD") == "" && os.Getenv("NEXTCLOUD_APP_PASSWORD") == "") {
		t.Skip("integration env not set")
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewClient(cfg)
	boards, err := client.GetBoards(ctx, false)
	if err != nil {
		t.Fatalf("GetBoards() error = %v", err)
	}
	if boards == nil {
		t.Fatal("GetBoards() returned nil slice")
	}
}
