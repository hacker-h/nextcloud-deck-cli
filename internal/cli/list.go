package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runList(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, listUsage())
	}
	args = normalizeListArgs(args)
	switch args[0] {
	case "list", "archived":
		fs := newFlagSet("list list", rt.stderr)
		boardSelector := fs.String("board", "", "board id or title")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardSelector != "", fmt.Sprintf("list %s requires --board", args[0])); err != nil {
			return err
		}
		boardID, err := resolveBoardSelector(rt, *boardSelector)
		if err != nil {
			return err
		}
		var (
			stacks any
		)
		if args[0] == "archived" {
			stacks, err = rt.client.GetArchivedStacks(rt.ctx, boardID)
		} else {
			stacks, err = rt.client.GetStacks(rt.ctx, boardID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(stacks, nil)
	case "find":
		fs := newFlagSet("list find", rt.stderr)
		boardSelector := fs.String("board", "", "board id or title")
		title := fs.String("title", "", "list title")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardSelector != "" && *title != "", "list find requires --board --title"); err != nil {
			return err
		}
		boardID, err := resolveBoardSelector(rt, *boardSelector)
		if err != nil {
			return err
		}
		stack, err := rt.client.FindStackByTitle(rt.ctx, boardID, *title)
		if err != nil {
			return err
		}
		return rt.printValue(stack, nil)
	case "get":
		fs := newFlagSet("list get", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		listID := fs.Int64("list", 0, "list id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *listID != 0, "list get requires --board --list"); err != nil {
			return err
		}
		stack, err := rt.client.GetStack(rt.ctx, *boardID, *listID)
		if err != nil {
			return err
		}
		return rt.printValue(stack, nil)
	case "create":
		fs := newFlagSet("list create", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		title := fs.String("title", "", "list title")
		order := fs.Int64("order", 999, "list order")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *title != "", "list create requires --board --title"); err != nil {
			return err
		}
		stack, err := rt.client.CreateStack(rt.ctx, *boardID, deck.CreateStackRequest{Title: *title, Order: *order})
		if err != nil {
			return err
		}
		return rt.printValue(stack, nil)
	case "rename", "reorder":
		fs := newFlagSet("list update", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		listID := fs.Int64("list", 0, "list id")
		title := fs.String("title", "", "list title")
		order := fs.Int64("order", -1, "list order")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *listID != 0, fmt.Sprintf("list %s requires --board --list", args[0])); err != nil {
			return err
		}
		stack, err := rt.client.GetStack(rt.ctx, *boardID, *listID)
		if err != nil {
			return err
		}
		if args[0] == "rename" {
			if err := require(*title != "", "list rename requires --title"); err != nil {
				return err
			}
			stack.Title = *title
		}
		if args[0] == "reorder" {
			if err := require(*order >= 0, "list reorder requires --order"); err != nil {
				return err
			}
			stack.Order = *order
		}
		updated, err := rt.client.UpdateStack(rt.ctx, *boardID, *listID, deck.UpdateStackRequest{Title: stack.Title, Order: stack.Order})
		if err != nil {
			return err
		}
		return rt.printValue(updated, nil)
	case "done", "undone":
		fs := newFlagSet("list done", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		listID := fs.Int64("list", 0, "list id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *listID != 0, fmt.Sprintf("list %s requires --board --list", args[0])); err != nil {
			return err
		}
		if err := rt.client.SetStackDone(rt.ctx, *boardID, *listID, args[0] == "done"); err != nil {
			return err
		}
		return rt.printStatus(args[0], map[string]any{"boardId": *boardID, "listId": *listID}, "marked list %d %s", *listID, args[0])
	case "delete":
		fs := newFlagSet("list delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		listID := fs.Int64("list", 0, "list id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *listID != 0, "list delete requires --board --list"); err != nil {
			return err
		}
		if err := rt.client.DeleteStack(rt.ctx, *boardID, *listID); err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"boardId": *boardID, "listId": *listID}, "deleted list %d", *listID)
	default:
		return validationf("unknown list command %q\nExamples:\n  deck list --board <id-or-title>\n  deck list board <id-or-title>\n  deck list find --board <id-or-title> --title <list-title>", args[0])
	}
}

func listUsage() string {
	return strings.TrimSpace(`deck list [list] --board <id-or-title>
deck list board <id-or-title>
deck list find --board <id-or-title> --title TEXT
deck list archived|get|create|rename|reorder|done|undone|delete
Aliases: deck stack ..., deck stacks ...`)
}

func normalizeListArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	if strings.HasPrefix(args[0], "-") {
		return append([]string{"list"}, args...)
	}
	if args[0] == "board" && len(args) >= 2 {
		normalized := []string{"list", "--board", args[1]}
		return append(normalized, args[2:]...)
	}
	return args
}

func resolveBoardSelector(rt *runtime, raw string) (int64, error) {
	selector := strings.TrimSpace(raw)
	if selector == "" {
		return 0, validationError("board selector must be an id or title")
	}
	if id, err := strconv.ParseInt(selector, 10, 64); err == nil {
		if id <= 0 {
			return 0, validationf("board %q is not valid; use a positive id or board title", raw)
		}
		return id, nil
	}

	boards, err := rt.client.GetBoards(rt.ctx, false)
	if err != nil {
		return 0, err
	}
	if board, ok, err := resolveBoardMatch(boards, selector, func(title string) bool { return title == selector }); ok || err != nil {
		return board.ID, err
	}
	lowerSelector := strings.ToLower(selector)
	if board, ok, err := resolveBoardMatch(boards, selector, func(title string) bool { return strings.ToLower(title) == lowerSelector }); ok || err != nil {
		return board.ID, err
	}
	if board, ok, err := resolveBoardMatch(boards, selector, func(title string) bool { return strings.Contains(strings.ToLower(title), lowerSelector) }); ok || err != nil {
		return board.ID, err
	}
	return 0, validationf("board %q not found; use a board id, exact title, or a unique case-insensitive title substring", selector)
}

func resolveBoardMatch(boards []deck.Board, selector string, matches func(string) bool) (deck.Board, bool, error) {
	var match deck.Board
	matched := make([]deck.Board, 0, 2)
	for _, board := range boards {
		if matches(board.Title) {
			match = board
			matched = append(matched, board)
		}
	}
	if len(matched) == 0 {
		return deck.Board{}, false, nil
	}
	if len(matched) > 1 {
		return deck.Board{}, false, validationf("board %q matched %d boards: %s; use a numeric board id or a more specific title", selector, len(matched), formatBoardMatches(matched))
	}
	return match, true, nil
}

func formatBoardMatches(boards []deck.Board) string {
	parts := make([]string, 0, len(boards))
	for _, board := range boards {
		parts = append(parts, fmt.Sprintf("%d %q", board.ID, board.Title))
	}
	return strings.Join(parts, ", ")
}
