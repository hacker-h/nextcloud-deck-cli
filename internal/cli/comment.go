package cli

import "fmt"

func runComment(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck comment list|create|update|delete")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("comment list", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		comments, err := rt.client.ListComments(rt.ctx, *cardID)
		if err != nil {
			return err
		}
		return rt.printValue(comments, nil)
	case "create":
		fs := newFlagSet("comment create", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		message := fs.String("message", "", "comment message")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		comment, err := rt.client.CreateComment(rt.ctx, *cardID, *message)
		if err != nil {
			return err
		}
		return rt.printValue(comment, nil)
	case "update":
		fs := newFlagSet("comment update", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		commentID := fs.Int64("comment", 0, "comment id")
		message := fs.String("message", "", "comment message")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		comment, err := rt.client.UpdateComment(rt.ctx, *cardID, *commentID, *message)
		if err != nil {
			return err
		}
		return rt.printValue(comment, nil)
	case "delete":
		fs := newFlagSet("comment delete", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		commentID := fs.Int64("comment", 0, "comment id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := rt.client.DeleteComment(rt.ctx, *cardID, *commentID); err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"cardId": *cardID, "commentId": *commentID}, "deleted comment %d", *commentID)
	default:
		return fmt.Errorf("unknown comment command %q", args[0])
	}
}
