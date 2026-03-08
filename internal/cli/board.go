package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runBoard(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck board list|get|create|update|archive|unarchive|delete|restore")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("board list", rt.stderr)
		details := fs.Bool("details", false, "include details")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		boards, err := rt.client.GetBoards(rt.ctx, *details)
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(rt.stdout, boards)
		}
		for _, board := range boards {
			if err := printLine(rt.stdout, "%d\t%s", board.ID, board.Title); err != nil {
				return err
			}
		}
		return nil
	case "get":
		fs := newFlagSet("board get", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "board get requires --board"); err != nil {
			return err
		}
		board, err := rt.client.GetBoard(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(rt.stdout, board)
		}
		return printJSON(rt.stdout, boardSummary(board))
	case "create":
		fs := newFlagSet("board create", rt.stderr)
		title := fs.String("title", "", "board title")
		color := fs.String("color", "ff0000", "board color")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*title != "", "board create requires --title"); err != nil {
			return err
		}
		board, err := rt.client.CreateBoard(rt.ctx, deck.BoardCreateRequest{Title: *title, Color: *color})
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(rt.stdout, board)
		}
		return printJSON(rt.stdout, boardSummary(board))
	case "update", "archive", "unarchive":
		fs := newFlagSet("board update", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		title := fs.String("title", "", "board title")
		color := fs.String("color", "", "board color")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, fmt.Sprintf("board %s requires --board", args[0])); err != nil {
			return err
		}
		board, err := rt.client.GetBoard(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		if args[0] == "update" {
			if *title != "" {
				board.Title = *title
			}
			if *color != "" {
				board.Color = *color
			}
		}
		if args[0] == "archive" {
			board.Archived = true
		}
		if args[0] == "unarchive" {
			board.Archived = false
		}
		updated, err := rt.client.UpdateBoard(rt.ctx, *boardID, deck.BoardUpdateRequest{Title: board.Title, Color: board.Color, Archived: board.Archived})
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(rt.stdout, updated)
		}
		return printJSON(rt.stdout, boardSummary(updated))
	case "delete":
		fs := newFlagSet("board delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "board delete requires --board"); err != nil {
			return err
		}
		if err := rt.client.DeleteBoard(rt.ctx, *boardID); err != nil {
			return err
		}
		return printLine(rt.stdout, "deleted board %d", *boardID)
	case "restore":
		fs := newFlagSet("board restore", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		jsonOut := fs.Bool("json", false, "json output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "board restore requires --board"); err != nil {
			return err
		}
		board, err := rt.client.RestoreBoard(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(rt.stdout, board)
		}
		return printJSON(rt.stdout, boardSummary(board))
	default:
		return fmt.Errorf("unknown board command %q", args[0])
	}
}
