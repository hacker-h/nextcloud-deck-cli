package cli

import (
	"fmt"
	"strconv"
)

func runTodo(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck todo list|add|check|uncheck")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("todo list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, extractTodos(card.Description))
	case "add":
		fs := newFlagSet("todo add", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		text := fs.String("text", "", "todo text")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*text != "", "todo add requires --text"); err != nil {
			return err
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		card.Description = addTodo(card.Description, *text)
		updated, err := rt.client.UpdateCard(rt.ctx, *boardID, *stackID, *cardID, baseCardUpdate(card, card.Title, &card.Description, card.Duedate))
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, extractTodos(updated.Description))
	case "check", "uncheck":
		fs := newFlagSet("todo check", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		index := fs.String("index", "", "todo index")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		parsed, err := strconv.Atoi(*index)
		if err != nil {
			return fmt.Errorf("invalid --index: %w", err)
		}
		card, err := rt.client.GetCard(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		card.Description, err = setTodoState(card.Description, parsed, args[0] == "check")
		if err != nil {
			return err
		}
		updated, err := rt.client.UpdateCard(rt.ctx, *boardID, *stackID, *cardID, baseCardUpdate(card, card.Title, &card.Description, card.Duedate))
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, extractTodos(updated.Description))
	default:
		return fmt.Errorf("unknown todo command %q", args[0])
	}
}
