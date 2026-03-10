package deck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestBackupRestoreRichKanbanBoard(t *testing.T) {
	if os.Getenv("NEXTCLOUD_FULL_BACKUP_SCENARIO") != "1" {
		t.Skip("set NEXTCLOUD_FULL_BACKUP_SCENARIO=1 to run the 200-card backup/restore scenario")
	}
	if os.Getenv("NEXTCLOUD_BASE_URL") == "" || os.Getenv("NEXTCLOUD_USERNAME") == "" || (os.Getenv("NEXTCLOUD_PASSWORD") == "" && os.Getenv("NEXTCLOUD_APP_PASSWORD") == "") {
		t.Skip("integration env not set")
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	client := NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	prefix := fmt.Sprintf("backup-rich-%d", time.Now().UnixNano())
	fixture, err := createRichKanbanFixture(ctx, client, t.TempDir(), prefix)
	if err != nil {
		t.Fatalf("createRichKanbanFixture() error = %v", err)
	}
	defer func() { _ = client.DeleteBoard(context.Background(), fixture.Board.ID) }()

	exportPath := filepath.Join(t.TempDir(), prefix+".json")
	if err := client.ExportBoard(ctx, fixture.Board.ID, exportPath); err != nil {
		t.Fatalf("ExportBoard() error = %v", err)
	}

	restored, err := client.ImportBoardFromFile(ctx, exportPath)
	if err != nil {
		t.Fatalf("ImportBoardFromFile() error = %v", err)
	}
	defer func() { _ = client.DeleteBoard(context.Background(), restored.ID) }()

	restoredStacks, err := client.GetStacks(ctx, restored.ID)
	if err != nil {
		t.Fatalf("GetStacks(restored) error = %v", err)
	}
	if len(restoredStacks) != 6 {
		t.Fatalf("expected 6 restored stacks, got %d", len(restoredStacks))
	}

	stackMap, cardMap, err := loadBoardCardsByStack(ctx, client, restored.ID, restoredStacks)
	if err != nil {
		t.Fatalf("loadBoardCardsByStack() error = %v", err)
	}

	if totalCards(cardMap) != 200 {
		t.Fatalf("expected 200 restored cards, got %d", totalCards(cardMap))
	}

	expectStackCardCount(t, cardMap, prefix+"-Inbox", 20)
	expectStackCardCount(t, cardMap, prefix+"-Backlog", 35)
	expectStackCardCount(t, cardMap, prefix+"-Ready", 30)
	expectStackCardCount(t, cardMap, prefix+"-Doing", 45)
	expectStackCardCount(t, cardMap, prefix+"-Blocked", 20)
	expectStackCardCount(t, cardMap, prefix+"-Done", 50)

	assertCardInStack(t, cardMap, prefix+"-Design onboarding checklist", prefix+"-Doing")
	assertCardInStack(t, cardMap, prefix+"-Ship API rate limit dashboard", prefix+"-Done")
	assertCardInStack(t, cardMap, prefix+"-Investigate flaky Deck session sync", prefix+"-Blocked")

	doingCard := findCardByTitle(cardMap[prefix+"-Doing"], prefix+"-Design onboarding checklist")
	if doingCard == nil {
		t.Fatal("expected representative doing card")
	}
	if !strings.Contains(doingCard.Description, "- [x] Write acceptance criteria") || !strings.Contains(doingCard.Description, "- [ ] Validate copy with Toni") {
		t.Fatalf("expected preserved mixed checklist state, got description=%q", doingCard.Description)
	}

	doneCard := findCardByTitle(cardMap[prefix+"-Done"], prefix+"-Ship API rate limit dashboard")
	if doneCard == nil {
		t.Fatal("expected representative done card")
	}
	if doneCard.Duedate == nil || *doneCard.Duedate == "" {
		t.Fatal("expected due date preserved on restored done card")
	}

	doingStack := stackMap[prefix+"-Doing"]
	doneStack := stackMap[prefix+"-Done"]
	if doingStack.ID == 0 || doneStack.ID == 0 {
		t.Fatal("expected restored doing/done stacks")
	}
}

type richFixture struct {
	Board  Board
	Stacks map[string]Stack
	Labels map[string]Label
}

func createRichKanbanFixture(ctx context.Context, client *Client, tempDir, prefix string) (richFixture, error) {
	board, err := client.CreateBoard(ctx, BoardCreateRequest{Title: prefix + "-board", Color: "0f766e"})
	if err != nil {
		return richFixture{}, err
	}

	stackNames := []string{"Inbox", "Backlog", "Ready", "Doing", "Blocked", "Done"}
	stacks := make(map[string]Stack, len(stackNames))
	for i, name := range stackNames {
		stack, err := client.CreateStack(ctx, board.ID, CreateStackRequest{Title: prefix + "-" + name, Order: int64((i + 1) * 10)})
		if err != nil {
			return richFixture{}, err
		}
		stacks[name] = stack
	}

	labels := map[string]Label{}
	for name, color := range map[string]string{"backend": "317CCC", "frontend": "F1DB50", "ops": "31CC7C", "bug": "FF7A66"} {
		label, err := client.CreateLabel(ctx, board.ID, CreateLabelRequest{Title: prefix + "-" + name, Color: color})
		if err != nil {
			return richFixture{}, err
		}
		labels[name] = label
	}

	seedCards := []struct {
		stack       string
		title       string
		description string
		due         *string
		labels      []Label
		moveTo      string
		markDone    bool
	}{
		{stack: "Backlog", title: prefix + "-Design onboarding checklist", description: "Prepare rollout\n- [x] Write acceptance criteria\n- [ ] Validate copy with Toni", due: futureTime(72 * time.Hour), labels: []Label{labels["frontend"]}, moveTo: "Doing"},
		{stack: "Ready", title: prefix + "-Ship API rate limit dashboard", description: "Ready for release\n- [x] Wire backend metrics\n- [x] Review alerts\n- [x] Publish dashboard", due: futureTime(24 * time.Hour), labels: []Label{labels["backend"], labels["ops"]}, moveTo: "Done", markDone: true},
		{stack: "Doing", title: prefix + "-Investigate flaky Deck session sync", description: "Observed on staging\n- [x] Reproduce timeout\n- [ ] Collect traces\n- [ ] Verify server logs", labels: []Label{labels["bug"], labels["backend"]}, moveTo: "Blocked"},
		{stack: "Inbox", title: prefix + "-Clarify backup retention policy", description: "Need stakeholder input\n- [ ] Confirm retention window\n- [ ] Confirm restore owner", labels: []Label{labels["ops"]}},
	}

	for _, seed := range seedCards {
		stack := stacks[seed.stack]
		card, err := client.CreateCard(ctx, board.ID, stack.ID, CreateCardRequest{Title: seed.title, Type: "plain", Order: 999, Description: &seed.description, Duedate: seed.due})
		if err != nil {
			return richFixture{}, err
		}
		for _, label := range seed.labels {
			_ = client.AssignLabel(ctx, board.ID, stack.ID, card.ID, label.ID)
		}
		if seed.markDone {
			_, _ = client.MarkCardDone(ctx, card.ID)
		}
		if seed.moveTo != "" {
			_ = client.ReorderCard(ctx, board.ID, stack.ID, card.ID, ReorderCardRequest{Order: 1, StackID: stacks[seed.moveTo].ID})
		}
	}

	attachmentPath := filepath.Join(tempDir, prefix+"-note.txt")
	if err := os.WriteFile(attachmentPath, []byte("restore scenario attachment"), 0o600); err == nil {
		card, err := client.CreateCard(ctx, board.ID, stacks["Doing"].ID, CreateCardRequest{Title: prefix + "-Attach architecture sketch", Type: "plain", Order: 999})
		if err == nil {
			_, _ = client.UploadAttachment(ctx, board.ID, stacks["Doing"].ID, card.ID, attachmentPath)
			_, _ = client.CreateComment(ctx, card.ID, "Attachment uploaded for restore scenario")
		}
	}

	targetCounts := map[string]int{"Inbox": 20, "Backlog": 35, "Ready": 30, "Doing": 45, "Blocked": 20, "Done": 50}
	for stackName, count := range targetCounts {
		created := 0
		for _, seed := range seedCards {
			if seed.stack == stackName && seed.moveTo == "" {
				created++
			}
			if seed.moveTo == stackName {
				created++
			}
		}
		if stackName == "Doing" {
			created++
		}
		remaining := count - created
		if remaining <= 0 {
			continue
		}
		if err := bulkCreateScenarioCards(ctx, client, board.ID, stacks[stackName], prefix, stackName, remaining); err != nil {
			return richFixture{}, err
		}
	}

	return richFixture{Board: board, Stacks: stacks, Labels: labels}, nil
}

func bulkCreateScenarioCards(ctx context.Context, client *Client, boardID int64, stack Stack, prefix, stackName string, count int) error {
	jobs := make(chan int)
	errCh := make(chan error, count)
	var wg sync.WaitGroup
	workers := 8
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				description := scenarioDescription(stackName, n)
				_, err := client.CreateCard(ctx, boardID, stack.ID, CreateCardRequest{Title: fmt.Sprintf("%s-%s task %03d", prefix, strings.ToLower(stackName), n), Type: "plain", Order: 999, Description: &description, Duedate: scenarioDueDate(stackName, n)})
				errCh <- err
			}
		}()
	}
	for i := 0; i < count; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func scenarioDescription(stackName string, n int) string {
	base := fmt.Sprintf("Workflow item for %s\n", stackName)
	switch stackName {
	case "Doing":
		return base + fmt.Sprintf("- [x] Investigate dependency %03d\n- [ ] Update implementation\n- [ ] Request review", n)
	case "Done":
		return base + fmt.Sprintf("- [x] Finish implementation %03d\n- [x] QA sign-off\n- [x] Merge to main", n)
	case "Blocked":
		return base + fmt.Sprintf("- [x] Capture issue %03d\n- [ ] Wait for vendor\n- [ ] Resume once dependency lands", n)
	default:
		return base + fmt.Sprintf("- [ ] Triage item %03d\n- [ ] Estimate effort", n)
	}
}

func scenarioDueDate(stackName string, n int) *string {
	switch stackName {
	case "Doing", "Done", "Blocked":
		return futureTime(time.Duration(24+n%5) * time.Hour)
	default:
		return nil
	}
}

func futureTime(delta time.Duration) *string {
	v := time.Now().Add(delta).UTC().Format(time.RFC3339)
	return &v
}

func loadBoardCardsByStack(ctx context.Context, client *Client, boardID int64, stacks []Stack) (map[string]Stack, map[string][]Card, error) {
	stackMap := make(map[string]Stack, len(stacks))
	cardMap := make(map[string][]Card, len(stacks))
	for _, stack := range stacks {
		fullStack, err := client.GetStack(ctx, boardID, stack.ID)
		if err != nil {
			return nil, nil, err
		}
		stackMap[fullStack.Title] = fullStack
		cardMap[fullStack.Title] = fullStack.Cards
	}
	return stackMap, cardMap, nil
}

func totalCards(cardMap map[string][]Card) int {
	total := 0
	for _, cards := range cardMap {
		total += len(cards)
	}
	return total
}

func expectStackCardCount(t *testing.T, cardMap map[string][]Card, stackTitle string, want int) {
	t.Helper()
	got := len(cardMap[stackTitle])
	if got != want {
		t.Fatalf("stack %q card count = %d, want %d", stackTitle, got, want)
	}
}

func assertCardInStack(t *testing.T, cardMap map[string][]Card, title, stackTitle string) {
	t.Helper()
	if findCardByTitle(cardMap[stackTitle], title) == nil {
		t.Fatalf("card %q not found in stack %q", title, stackTitle)
	}
}

func findCardByTitle(cards []Card, title string) *Card {
	for i := range cards {
		if cards[i].Title == title {
			return &cards[i]
		}
	}
	return nil
}
