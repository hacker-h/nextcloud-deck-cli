package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func TestCLIIntegrationDeckFlow(t *testing.T) {
	if os.Getenv("NEXTCLOUD_BASE_URL") == "" || os.Getenv("NEXTCLOUD_USERNAME") == "" || (os.Getenv("NEXTCLOUD_PASSWORD") == "" && os.Getenv("NEXTCLOUD_APP_PASSWORD") == "") {
		t.Skip("integration env not set")
	}

	prefix := fmt.Sprintf("opencode-%d", time.Now().UnixNano())
	boardTitle := prefix + "-board"
	listA := prefix + "-list-a"
	listB := prefix + "-list-b"
	cardTitle := prefix + "-card"
	due := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	board := runJSON[deck.Board](t, "board", "create", "--title", boardTitle, "--color", "ff6600")
	boardID := board.ID
	defer func() { _ = runMaybe(t, "board", "delete", "--board", fmt.Sprint(boardID)) }()

	boards := runJSON[[]deck.Board](t, "board", "list", "--details", "--json")
	assertBoardPresent(t, boards, boardID)

	board = runJSON[deck.Board](t, "board", "update", "--board", fmt.Sprint(boardID), "--title", boardTitle+"-updated", "--color", "00aa88")
	if board.Title != boardTitle+"-updated" {
		t.Fatalf("board title not updated: %#v", board)
	}
	board = runJSON[deck.Board](t, "board", "archive", "--board", fmt.Sprint(boardID))
	if !board.Archived {
		t.Fatalf("board not archived: %#v", board)
	}
	board = runJSON[deck.Board](t, "board", "unarchive", "--board", fmt.Sprint(boardID))
	if board.Archived {
		t.Fatalf("board still archived: %#v", board)
	}

	stack1 := runJSON[deck.Stack](t, "list", "create", "--board", fmt.Sprint(boardID), "--title", listA, "--order", "10")
	stack2 := runJSON[deck.Stack](t, "list", "create", "--board", fmt.Sprint(boardID), "--title", listB, "--order", "20")
	stack1 = runJSON[deck.Stack](t, "list", "rename", "--board", fmt.Sprint(boardID), "--list", fmt.Sprint(stack1.ID), "--title", listA+"-renamed")
	stack2 = runJSON[deck.Stack](t, "list", "reorder", "--board", fmt.Sprint(boardID), "--list", fmt.Sprint(stack2.ID), "--order", "5")
	stacks := runJSON[[]deck.Stack](t, "list", "list", "--board", fmt.Sprint(boardID))
	if len(stacks) < 2 {
		t.Fatalf("expected at least 2 stacks, got %d", len(stacks))
	}
	_ = runJSON[[]deck.Stack](t, "list", "archived", "--board", fmt.Sprint(boardID))
	_ = runJSON[deck.Stack](t, "list", "get", "--board", fmt.Sprint(boardID), "--list", fmt.Sprint(stack1.ID))

	card := runJSON[deck.Card](t, "card", "create", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--title", cardTitle, "--description", "initial", "--due", due)
	cardID := card.ID
	card = runJSON[deck.Card](t, "card", "rename", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID), "--title", cardTitle+"-renamed")
	card = runJSON[deck.Card](t, "card", "describe", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID), "--description", "hello")
	if !strings.Contains(card.Description, "hello") {
		t.Fatalf("card description not updated: %#v", card)
	}
	_ = runJSON[map[string]any](t, "card", "due", "get", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID))
	card = runJSON[deck.Card](t, "card", "due", "set", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID), "--value", time.Now().Add(48*time.Hour).UTC().Format(time.RFC3339))
	if card.Duedate == nil {
		t.Fatal("expected due date set")
	}
	card = runJSON[deck.Card](t, "card", "due", "clear", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID))
	if card.Duedate != nil {
		t.Fatalf("expected due date cleared: %#v", card.Duedate)
	}
	_ = runJSON[[]deck.Card](t, "card", "list", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID))
	_ = runJSON[deck.Card](t, "card", "get", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID))
	runOK(t, "card", "reorder", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID), "--card", fmt.Sprint(cardID), "--order", "3")
	runOK(t, "card", "move", "--board", fmt.Sprint(boardID), "--from-stack", fmt.Sprint(stack1.ID), "--to-stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--order", "1")
	stack1Cards := runJSON[[]deck.Card](t, "card", "list", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack1.ID))
	stack2Cards := runJSON[[]deck.Card](t, "card", "list", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID))
	if containsCard(stack1Cards, cardID) || !containsCard(stack2Cards, cardID) {
		t.Fatalf("card move verification failed; stack1=%#v stack2=%#v", stack1Cards, stack2Cards)
	}
	card = runJSON[deck.Card](t, "card", "get", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	card = runJSON[deck.Card](t, "card", "archive", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	if !card.Archived {
		t.Fatalf("card not archived: %#v", card)
	}
	card = runJSON[deck.Card](t, "card", "unarchive", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	if card.Archived {
		t.Fatalf("card still archived: %#v", card)
	}

	todos := runJSON[[]markdownTodo](t, "todo", "add", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--text", "first")
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	todos = runJSON[[]markdownTodo](t, "todo", "check", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--index", "1")
	if !todos[0].Checked {
		t.Fatalf("todo not checked: %#v", todos[0])
	}
	todos = runJSON[[]markdownTodo](t, "todo", "uncheck", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--index", "1")
	if todos[0].Checked {
		t.Fatalf("todo not unchecked: %#v", todos[0])
	}
	_ = runJSON[[]markdownTodo](t, "todo", "list", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))

	label := runJSON[deck.Label](t, "label", "create", "--board", fmt.Sprint(boardID), "--title", prefix+"-label", "--color", "31CC7C")
	_ = runJSON[[]deck.Label](t, "label", "list", "--board", fmt.Sprint(boardID))
	label = runJSON[deck.Label](t, "label", "get", "--board", fmt.Sprint(boardID), "--label", fmt.Sprint(label.ID))
	label = runJSON[deck.Label](t, "label", "update", "--board", fmt.Sprint(boardID), "--label", fmt.Sprint(label.ID), "--title", prefix+"-label-updated")
	if label.Title != prefix+"-label-updated" {
		t.Fatalf("label not updated: %#v", label)
	}
	runOK(t, "card", "assign-label", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--label", fmt.Sprint(label.ID))
	card = runJSON[deck.Card](t, "card", "get", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	if len(card.Labels) == 0 {
		t.Fatal("expected label assignment")
	}
	runOK(t, "card", "remove-label", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--label", fmt.Sprint(label.ID))
	userID := os.Getenv("NEXTCLOUD_USERNAME")
	assignment := runJSON[deck.Assignment](t, "card", "assign-user", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--user", userID)
	if assignment.Participant.UID == "" {
		t.Fatalf("expected assignment participant, got %#v", assignment)
	}
	runOK(t, "card", "unassign-user", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--user", userID)
	runOK(t, "label", "delete", "--board", fmt.Sprint(boardID), "--label", fmt.Sprint(label.ID))

	comment := runJSON[deck.Comment](t, "comment", "create", "--card", fmt.Sprint(cardID), "--message", "first comment")
	commentID := comment.ID
	comments := runJSON[[]deck.Comment](t, "comment", "list", "--card", fmt.Sprint(cardID))
	if len(comments) == 0 {
		t.Fatal("expected listed comments")
	}
	comment = runJSON[deck.Comment](t, "comment", "update", "--card", fmt.Sprint(cardID), "--comment", fmt.Sprint(commentID), "--message", "edited comment")
	if comment.Message != "edited comment" {
		t.Fatalf("expected updated comment, got %#v", comment)
	}
	runOK(t, "comment", "delete", "--card", fmt.Sprint(cardID), "--comment", fmt.Sprint(commentID))

	uploadPath := filepath.Join(t.TempDir(), "attachment.txt")
	if err := os.WriteFile(uploadPath, []byte("attachment-body"), 0o600); err != nil {
		t.Fatal(err)
	}
	attachment := runJSON[deck.Attachment](t, "attachment", "upload", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--file", uploadPath)
	attachments := runJSON[[]deck.Attachment](t, "attachment", "list", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	if len(attachments) == 0 {
		t.Fatal("expected attachments")
	}
	downloadPath := filepath.Join(t.TempDir(), "downloaded.txt")
	runOK(t, "attachment", "download", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--attachment", fmt.Sprint(attachment.ID), "--out", downloadPath)
	data, err := os.ReadFile(downloadPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("attachment")) {
		t.Fatalf("unexpected downloaded content: %q", string(data))
	}
	runOK(t, "attachment", "delete", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--attachment", fmt.Sprint(attachment.ID))
	_ = runMaybe(t, "attachment", "restore", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID), "--attachment", fmt.Sprint(attachment.ID))

	configBefore := runJSON[map[string]any](t, "config", "get")
	if current, ok := configBefore["cardIdBadge"].(bool); ok {
		_ = runMaybe(t, "config", "set", "--key", "cardIdBadge", "--value", fmt.Sprint(!current))
		_ = runMaybe(t, "config", "set", "--key", "cardIdBadge", "--value", fmt.Sprint(current))
	}
	_ = runJSON[[]deck.ACLRule](t, "share", "list", "--board", fmt.Sprint(boardID))
	if shareRules, err := runMaybeJSON[[]deck.ACLRule](t, "share", "create", "--board", fmt.Sprint(boardID), "--type", "0", "--participant", userID, "--edit", "true", "--share", "false", "--manage", "false"); err == nil && len(shareRules) > 0 {
		_ = runMaybe(t, "share", "update", "--board", fmt.Sprint(boardID), "--share-id", fmt.Sprint(shareRules[0].ID), "--edit", "true", "--share", "false", "--manage", "false")
		_ = runMaybe(t, "share", "delete", "--board", fmt.Sprint(boardID), "--share-id", fmt.Sprint(shareRules[0].ID))
	} else if err != nil {
		t.Logf("share create blocked on this server: %v", err)
	}

	runOK(t, "board", "delete", "--board", fmt.Sprint(boardID))
	if err := runMaybe(t, "board", "restore", "--board", fmt.Sprint(boardID)); err == nil {
		board = runJSON[deck.Board](t, "board", "get", "--board", fmt.Sprint(boardID))
		if board.ID != boardID {
			t.Fatalf("restored wrong board: %#v", board)
		}
	} else {
		t.Logf("board restore not permitted on this server: %v", err)
	}
	_ = runMaybe(t, "card", "delete", "--board", fmt.Sprint(boardID), "--stack", fmt.Sprint(stack2.ID), "--card", fmt.Sprint(cardID))
	_ = runMaybe(t, "list", "delete", "--board", fmt.Sprint(boardID), "--list", fmt.Sprint(stack1.ID))
	_ = runMaybe(t, "list", "delete", "--board", fmt.Sprint(boardID), "--list", fmt.Sprint(stack2.ID))
	_ = runMaybe(t, "board", "delete", "--board", fmt.Sprint(boardID))
	boardID = 0
}

func runJSON[T any](t *testing.T, args ...string) T {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if err := Run(args, &stdout, &stderr); err != nil {
		t.Fatalf("Run(%v) error = %v\nstderr=%s\nstdout=%s", args, err, stderr.String(), stdout.String())
	}
	var value T
	if err := json.Unmarshal(stdout.Bytes(), &value); err != nil {
		t.Fatalf("json.Unmarshal(%v) error = %v\nstdout=%s", args, err, stdout.String())
	}
	return value
}

func runOK(t *testing.T, args ...string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if err := Run(args, &stdout, &stderr); err != nil {
		t.Fatalf("Run(%v) error = %v\nstderr=%s\nstdout=%s", args, err, stderr.String(), stdout.String())
	}
}

func runMaybe(t *testing.T, args ...string) error {
	t.Helper()
	var stdout, stderr bytes.Buffer
	return Run(args, &stdout, &stderr)
}

func runMaybeJSON[T any](t *testing.T, args ...string) (T, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	var zero T
	if err := Run(args, &stdout, &stderr); err != nil {
		return zero, err
	}
	var value T
	if err := json.Unmarshal(stdout.Bytes(), &value); err != nil {
		return zero, err
	}
	return value, nil
}

func assertBoardPresent(t *testing.T, boards []deck.Board, boardID int64) {
	t.Helper()
	for _, board := range boards {
		if board.ID == boardID {
			return
		}
	}
	t.Fatalf("board %d not found in list", boardID)
}

func containsCard(cards []deck.Card, id int64) bool {
	for _, card := range cards {
		if card.ID == id {
			return true
		}
	}
	return false
}
