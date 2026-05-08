package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runList(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck list list|get|find|archived|create|rename|reorder|delete")
	}
	switch args[0] {
	case "list", "archived":
		fs := newFlagSet("list list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, fmt.Sprintf("list %s requires --board", args[0])); err != nil {
			return err
		}
		var (
			stacks any
			err    error
		)
		if args[0] == "archived" {
			stacks, err = rt.client.GetArchivedStacks(rt.ctx, *boardID)
		} else {
			stacks, err = rt.client.GetStacks(rt.ctx, *boardID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(stacks, nil)
	case "find":
		fs := newFlagSet("list find", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		title := fs.String("title", "", "list title")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *title != "", "list find requires --board --title"); err != nil {
			return err
		}
		stack, err := rt.client.FindStackByTitle(rt.ctx, *boardID, *title)
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
		return fmt.Errorf("unknown list command %q", args[0])
	}
}
