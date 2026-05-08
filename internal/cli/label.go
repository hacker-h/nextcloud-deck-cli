package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runLabel(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck label list|get|find|create|update|delete")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("label list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "label list requires --board"); err != nil {
			return err
		}
		labels, err := rt.client.ListLabels(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return rt.printValue(labels, nil)
	case "find":
		fs := newFlagSet("label find", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		title := fs.String("title", "", "label title")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *title != "", "label find requires --board --title"); err != nil {
			return err
		}
		label, err := rt.client.FindLabelByTitle(rt.ctx, *boardID, *title)
		if err != nil {
			return err
		}
		return rt.printValue(label, nil)
	case "get":
		fs := newFlagSet("label get", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		labelID := fs.Int64("label", 0, "label id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *labelID != 0, "label get requires --board --label"); err != nil {
			return err
		}
		label, err := rt.client.GetLabel(rt.ctx, *boardID, *labelID)
		if err != nil {
			return err
		}
		return rt.printValue(label, nil)
	case "create":
		fs := newFlagSet("label create", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		title := fs.String("title", "", "label title")
		color := fs.String("color", "31CC7C", "label color")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *title != "", "label create requires --board --title"); err != nil {
			return err
		}
		label, err := rt.client.CreateLabel(rt.ctx, *boardID, deck.CreateLabelRequest{Title: *title, Color: *color})
		if err != nil {
			return err
		}
		return rt.printValue(label, nil)
	case "update":
		fs := newFlagSet("label update", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		labelID := fs.Int64("label", 0, "label id")
		title := fs.String("title", "", "label title")
		color := fs.String("color", "", "label color")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *labelID != 0, "label update requires --board --label"); err != nil {
			return err
		}
		current, err := rt.client.GetLabel(rt.ctx, *boardID, *labelID)
		if err != nil {
			return err
		}
		if *title != "" {
			current.Title = *title
		}
		if *color != "" {
			current.Color = *color
		}
		label, err := rt.client.UpdateLabel(rt.ctx, *boardID, *labelID, deck.UpdateLabelRequest{Title: current.Title, Color: current.Color})
		if err != nil {
			return err
		}
		return rt.printValue(label, nil)
	case "delete":
		fs := newFlagSet("label delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		labelID := fs.Int64("label", 0, "label id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *labelID != 0, "label delete requires --board --label"); err != nil {
			return err
		}
		if err := rt.client.DeleteLabel(rt.ctx, *boardID, *labelID); err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"boardId": *boardID, "labelId": *labelID}, "deleted label %d", *labelID)
	default:
		return fmt.Errorf("unknown label command %q", args[0])
	}
}
