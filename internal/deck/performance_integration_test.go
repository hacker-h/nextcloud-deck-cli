package deck

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestPerformanceLargeBoard(t *testing.T) {
	if os.Getenv("NEXTCLOUD_BASE_URL") == "" || os.Getenv("NEXTCLOUD_USERNAME") == "" || (os.Getenv("NEXTCLOUD_PASSWORD") == "" && os.Getenv("NEXTCLOUD_APP_PASSWORD") == "") {
		t.Skip("integration env not set")
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	client := NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	prefix := fmt.Sprintf("perf-%d", time.Now().UnixNano())
	board, err := client.CreateBoard(ctx, BoardCreateRequest{Title: prefix + "-board", Color: "224466"})
	if err != nil {
		t.Fatalf("CreateBoard() error = %v", err)
	}
	defer func() { _ = client.DeleteBoard(context.Background(), board.ID) }()

	stackA, err := client.CreateStack(ctx, board.ID, CreateStackRequest{Title: prefix + "-a", Order: 10})
	if err != nil {
		t.Fatalf("CreateStack(A) error = %v", err)
	}
	stackB, err := client.CreateStack(ctx, board.ID, CreateStackRequest{Title: prefix + "-b", Order: 20})
	if err != nil {
		t.Fatalf("CreateStack(B) error = %v", err)
	}

	seqCards, seqDuration := createCardsSequentially(t, ctx, client, board.ID, stackA.ID, prefix+"-seq", 100)
	t.Logf("create 100 cards sequential: total=%s avg=%s", seqDuration, seqDuration/100)

	parallelCards, parDuration := createCardsInParallel(t, ctx, client, board.ID, stackA.ID, prefix+"-par", 100, 8)
	t.Logf("create 100 cards parallel(8): total=%s avg=%s speedup=%.2fx", parDuration, parDuration/100, float64(seqDuration)/float64(parDuration))

	stackList100Start := time.Now()
	stackAfter100, err := client.GetStack(ctx, board.ID, stackA.ID)
	if err != nil {
		t.Fatalf("GetStack(100+) error = %v", err)
	}
	stackList100Duration := time.Since(stackList100Start)
	t.Logf("fetch stack with %d cards: %s", len(stackAfter100.Cards), stackList100Duration)

	boardDetailsStart := time.Now()
	boardDetails, err := client.GetBoard(ctx, board.ID)
	if err != nil {
		t.Fatalf("GetBoard() error = %v", err)
	}
	boardDetailsDuration := time.Since(boardDetailsStart)
	t.Logf("fetch board details with %d stacks: %s", len(boardDetails.Stacks), boardDetailsDuration)

	moveIDs := append([]int64{}, seqCards[:25]...)
	moveIDs = append(moveIDs, parallelCards[:25]...)
	moveDuration := moveCardsSequentially(t, ctx, client, board.ID, stackA.ID, stackB.ID, moveIDs)
	t.Logf("move %d cards sequentially: total=%s avg=%s", len(moveIDs), moveDuration, moveDuration/time.Duration(len(moveIDs)))

	getStack200Start := time.Now()
	stackAfterMoves, err := client.GetStack(ctx, board.ID, stackA.ID)
	if err != nil {
		t.Fatalf("GetStack(after moves) error = %v", err)
	}
	stackBState, err := client.GetStack(ctx, board.ID, stackB.ID)
	if err != nil {
		t.Fatalf("GetStack(target) error = %v", err)
	}
	getStack200Duration := time.Since(getStack200Start)
	t.Logf("fetch both stacks after moves: %s source_cards=%d target_cards=%d", getStack200Duration, len(stackAfterMoves.Cards), len(stackBState.Cards))

	if len(stackAfter100.Cards) < 200 {
		t.Fatalf("expected at least 200 cards after creation, got %d", len(stackAfter100.Cards))
	}
	if len(stackBState.Cards) < len(moveIDs) {
		t.Fatalf("expected at least %d moved cards in target stack, got %d", len(moveIDs), len(stackBState.Cards))
	}

	measureSearchStart := time.Now()
	results, err := client.SearchCards(ctx, prefix, 20)
	if err != nil {
		t.Fatalf("SearchCards() error = %v", err)
	}
	t.Logf("search cards on large board: %s results=%d", time.Since(measureSearchStart), len(results))
}

func createCardsSequentially(t *testing.T, ctx context.Context, client *Client, boardID, stackID int64, prefix string, count int) ([]int64, time.Duration) {
	t.Helper()
	ids := make([]int64, 0, count)
	start := time.Now()
	for i := 0; i < count; i++ {
		card, err := client.CreateCard(ctx, boardID, stackID, CreateCardRequest{Title: fmt.Sprintf("%s-%03d", prefix, i), Type: "plain", Order: 999})
		if err != nil {
			t.Fatalf("CreateCard(sequential %d) error = %v", i, err)
		}
		ids = append(ids, card.ID)
	}
	return ids, time.Since(start)
}

func createCardsInParallel(t *testing.T, ctx context.Context, client *Client, boardID, stackID int64, prefix string, count, workers int) ([]int64, time.Duration) {
	t.Helper()
	type result struct {
		id  int64
		err error
	}
	jobs := make(chan int)
	results := make(chan result, count)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				card, err := client.CreateCard(ctx, boardID, stackID, CreateCardRequest{Title: fmt.Sprintf("%s-%03d", prefix, i), Type: "plain", Order: 999})
				if err != nil {
					results <- result{err: err}
					continue
				}
				results <- result{id: card.ID}
			}
		}()
	}
	start := time.Now()
	for i := 0; i < count; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(results)
	ids := make([]int64, 0, count)
	for result := range results {
		if result.err != nil {
			t.Fatalf("CreateCard(parallel) error = %v", result.err)
		}
		ids = append(ids, result.id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, time.Since(start)
}

func moveCardsSequentially(t *testing.T, ctx context.Context, client *Client, boardID, fromStackID, toStackID int64, cardIDs []int64) time.Duration {
	t.Helper()
	start := time.Now()
	for i, cardID := range cardIDs {
		if err := client.ReorderCard(ctx, boardID, fromStackID, cardID, ReorderCardRequest{Order: int64(i + 1), StackID: toStackID}); err != nil {
			t.Fatalf("ReorderCard(move %d) error = %v", i, err)
		}
	}
	return time.Since(start)
}
