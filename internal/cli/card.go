package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runCard(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck card list|get|deleted|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|update|due|assign-user|unassign-user|assign-label|remove-label|assign-dependent|remove-dependent")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("card list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0, "card list requires --board --stack"); err != nil {
			return err
		}
		cards, err := rt.client.ListCards(rt.ctx, *boardID, *stackID)
		if err != nil {
			return err
		}
		return rt.printValue(cards, nil)
	case "get":
		fs := newFlagSet("card get", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, "card get requires --board --stack --card"); err != nil {
			return err
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	case "deleted":
		fs := newFlagSet("card deleted", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "card deleted requires --board"); err != nil {
			return err
		}
		cards, err := rt.client.ListDeletedCards(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return rt.printValue(cards, nil)
	case "create":
		fs := newFlagSet("card create", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		title := fs.String("title", "", "card title")
		cardType := fs.String("type", "plain", "card type")
		color := fs.String("color", "", "card color")
		descriptionInput := addTextInputFlags(fs, "description", "description-file", "description-stdin", "card description", true)
		due := fs.String("due", "", "ISO-8601 due date")
		start := fs.String("start", "", "ISO-8601 start date")
		order := fs.Int64("order", 999, "card order")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *title != "", "card create requires --board --stack --title"); err != nil {
			return err
		}
		description, hasDescription, err := descriptionInput.resolve(fs)
		if err != nil {
			return err
		}
		req := deck.CreateCardRequest{Title: *title, Type: *cardType, Color: *color, Order: *order}
		if hasDescription {
			req.Description = stringPtr(description)
		}
		if *due != "" {
			req.Duedate = stringPtr(*due)
		}
		if *start != "" {
			req.Startdate = stringPtr(*start)
		}
		card, err := rt.client.CreateCard(rt.ctx, *boardID, *stackID, req)
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	case "clone":
		fs := newFlagSet("card clone", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		targetStackID := fs.Int64("to-stack", 0, "target stack id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*cardID != 0 && *targetStackID != 0, "card clone requires --card --to-stack"); err != nil {
			return err
		}
		card, err := rt.client.CloneCard(rt.ctx, *cardID, *targetStackID)
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	case "delete":
		fs := newFlagSet("card delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, "card delete requires --board --stack --card"); err != nil {
			return err
		}
		if err := rt.client.DeleteCard(rt.ctx, *boardID, *stackID, *cardID); err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID}, "deleted card %d", *cardID)
	case "move", "reorder":
		fs := newFlagSet("card reorder", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		fromStackID := fs.Int64("from-stack", 0, "source stack id")
		toStackID := fs.Int64("to-stack", 0, "target stack id")
		cardID := fs.Int64("card", 0, "card id")
		order := fs.Int64("order", -1, "target order")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if args[0] == "move" {
			if err := require(*boardID != 0 && *fromStackID != 0 && *toStackID != 0 && *cardID != 0, "card move requires --board --from-stack --to-stack --card"); err != nil {
				return err
			}
			targetOrder := *order
			if targetOrder < 0 {
				targetOrder = 999
			}
			if err := rt.client.ReorderCard(rt.ctx, *boardID, *fromStackID, *cardID, deck.ReorderCardRequest{Order: targetOrder, StackID: *toStackID}); err != nil {
				return err
			}
			return rt.printStatus("moved", map[string]any{"boardId": *boardID, "fromStackId": *fromStackID, "toStackId": *toStackID, "cardId": *cardID, "order": targetOrder}, "moved card %d", *cardID)
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *order >= 0, "card reorder requires --board --stack --card --order"); err != nil {
			return err
		}
		if err := rt.client.ReorderCard(rt.ctx, *boardID, *stackID, *cardID, deck.ReorderCardRequest{Order: *order, StackID: *stackID}); err != nil {
			return err
		}
		return rt.printStatus("reordered", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "order": *order}, "reordered card %d", *cardID)
	case "archive", "unarchive":
		fs := newFlagSet("card archive", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, fmt.Sprintf("card %s requires --board --stack --card", args[0])); err != nil {
			return err
		}
		var (
			card deck.Card
			err  error
		)
		if args[0] == "archive" {
			card, err = rt.client.ArchiveCard(rt.ctx, *boardID, *stackID, *cardID)
		} else {
			card, err = rt.client.UnarchiveCard(rt.ctx, *boardID, *stackID, *cardID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	case "done", "undone":
		fs := newFlagSet("card done", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*cardID != 0, fmt.Sprintf("card %s requires --card", args[0])); err != nil {
			return err
		}
		var (
			card deck.Card
			err  error
		)
		if args[0] == "done" {
			card, err = rt.client.MarkCardDone(rt.ctx, *cardID)
		} else {
			card, err = rt.client.MarkCardUndone(rt.ctx, *cardID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	case "rename", "describe", "update":
		fs := newFlagSet("card update", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		title := fs.String("title", "", "card title")
		cardType := fs.String("type", "", "card type")
		color := fs.String("color", "", "card color")
		due := fs.String("due", "", "ISO-8601 due date")
		start := fs.String("start", "", "ISO-8601 start date")
		descriptionInput := addTextInputFlags(fs, "description", "description-file", "description-stdin", "card description", true)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, fmt.Sprintf("card %s requires --board --stack --card", args[0])); err != nil {
			return err
		}
		description, _, err := descriptionInput.resolve(fs)
		if err != nil {
			return err
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		if args[0] == "rename" {
			if err := require(*title != "", "card rename requires --title"); err != nil {
				return err
			}
			card.Title = *title
		}
		if args[0] == "describe" {
			card.Description = description
		}
		if args[0] == "update" {
			if *title != "" {
				card.Title = *title
			}
			if *cardType != "" {
				card.Type = *cardType
			}
			if *color != "" {
				card.Color = *color
			}
			if *due != "" {
				card.Duedate = stringPtr(*due)
			}
			if *start != "" {
				card.Startdate = stringPtr(*start)
			}
			if description != "" {
				card.Description = description
			}
		}
		updated, err := rt.client.UpdateCard(rt.ctx, *boardID, *stackID, *cardID, baseCardUpdate(card, card.Title, &card.Description, card.Duedate))
		if err != nil {
			return err
		}
		return rt.printValue(updated, nil)
	case "due":
		return runCardDue(rt, args[1:])
	case "assign-user", "unassign-user":
		fs := newFlagSet("card user", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		userID := fs.String("user", "", "user id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *userID != "", fmt.Sprintf("card %s requires --board --stack --card --user", args[0])); err != nil {
			return err
		}
		if args[0] == "assign-user" {
			assignment, err := rt.client.AssignUser(rt.ctx, *boardID, *stackID, *cardID, *userID)
			if err != nil {
				return err
			}
			return rt.printValue(assignment, nil)
		}
		if err := rt.client.UnassignUser(rt.ctx, *boardID, *stackID, *cardID, *userID); err != nil {
			return err
		}
		return rt.printStatus("unassigned", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "user": *userID}, "unassigned user %s", *userID)
	case "assign-label", "remove-label":
		fs := newFlagSet("card label", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		labelID := fs.Int64("label", 0, "label id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *labelID != 0, fmt.Sprintf("card %s requires --board --stack --card --label", args[0])); err != nil {
			return err
		}
		if args[0] == "assign-label" {
			if err := rt.client.AssignLabel(rt.ctx, *boardID, *stackID, *cardID, *labelID); err != nil {
				return err
			}
			return rt.printStatus("assigned", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "labelId": *labelID}, "assigned label %d", *labelID)
		}
		if err := rt.client.RemoveLabel(rt.ctx, *boardID, *stackID, *cardID, *labelID); err != nil {
			return err
		}
		return rt.printStatus("removed", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "labelId": *labelID}, "removed label %d", *labelID)
	case "assign-dependent", "remove-dependent":
		fs := newFlagSet("card dependent", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		dependentID := fs.Int64("dependent-card", 0, "dependent card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *dependentID != 0, fmt.Sprintf("card %s requires --board --stack --card --dependent-card", args[0])); err != nil {
			return err
		}
		var (
			card deck.Card
			err  error
		)
		if args[0] == "assign-dependent" {
			card, err = rt.client.AssignDependentCard(rt.ctx, *boardID, *stackID, *cardID, *dependentID)
		} else {
			card, err = rt.client.RemoveDependentCard(rt.ctx, *boardID, *stackID, *cardID, *dependentID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(card, nil)
	default:
		return fmt.Errorf("unknown card command %q", args[0])
	}
}

func runCardDue(rt *runtime, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("card due requires get, set, or clear")
	}
	switch args[0] {
	case "get":
		fs := newFlagSet("card due get", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, "card due get requires --board --stack --card"); err != nil {
			return err
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		return rt.printValue(map[string]any{"id": card.ID, "dueDate": card.Duedate}, nil)
	case "set", "clear":
		fs := newFlagSet("card due set", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		value := fs.String("value", "", "ISO-8601 due date")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, fmt.Sprintf("card due %s requires --board --stack --card", args[0])); err != nil {
			return err
		}
		if args[0] == "set" {
			if err := require(*value != "", "card due set requires --value"); err != nil {
				return err
			}
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		var due *string
		if args[0] == "set" {
			due = stringPtr(*value)
		}
		updated, err := rt.client.UpdateCard(rt.ctx, *boardID, *stackID, *cardID, baseCardUpdate(card, card.Title, &card.Description, due))
		if err != nil {
			return err
		}
		return rt.printValue(updated, nil)
	default:
		return fmt.Errorf("unknown card due command %q", args[0])
	}
}
