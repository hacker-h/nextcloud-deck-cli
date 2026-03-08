package cli

import "fmt"

func runAttachment(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck attachment list|upload|download|delete|restore")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("attachment list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		attachments, err := rt.client.ListAttachments(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, attachments)
	case "upload":
		fs := newFlagSet("attachment upload", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		filePath := fs.String("file", "", "file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		attachment, err := rt.client.UploadAttachment(rt.ctx, *boardID, *stackID, *cardID, *filePath)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, attachment)
	case "download":
		fs := newFlagSet("attachment download", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := rt.client.DownloadAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID, *out); err != nil {
			return err
		}
		return printLine(rt.stdout, "downloaded attachment %d", *attachmentID)
	case "delete":
		fs := newFlagSet("attachment delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := rt.client.DeleteAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID); err != nil {
			return err
		}
		return printLine(rt.stdout, "deleted attachment %d", *attachmentID)
	case "restore":
		fs := newFlagSet("attachment restore", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		attachment, err := rt.client.RestoreAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, attachment)
	default:
		return fmt.Errorf("unknown attachment command %q", args[0])
	}
}
