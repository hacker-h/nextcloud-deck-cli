package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

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
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0, "attachment list requires --board --stack --card"); err != nil {
			return err
		}
		attachments, err := rt.client.ListAttachments(rt.ctx, *boardID, *stackID, *cardID)
		if err != nil {
			return err
		}
		return rt.printValue(attachments, nil)
	case "upload":
		fs := newFlagSet("attachment upload", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		filePath := fs.String("file", "", "file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *filePath != "", "attachment upload requires --board --stack --card --file"); err != nil {
			return err
		}
		attachment, err := rt.client.UploadAttachment(rt.ctx, *boardID, *stackID, *cardID, *filePath)
		if err != nil {
			return err
		}
		return rt.printValue(attachment, nil)
	case "download":
		fs := newFlagSet("attachment download", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		attachmentType := fs.String("type", "", "typed attachment type")
		out := fs.String("out", "", "output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *attachmentID != 0 && *out != "", "attachment download requires --board --stack --card --attachment --out"); err != nil {
			return err
		}
		var err error
		if *attachmentType != "" {
			err = rt.client.DownloadAttachmentRef(rt.ctx, *cardID, fmt.Sprintf("%s:%d", *attachmentType, *attachmentID), *out)
		} else {
			err = rt.client.DownloadAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID, *out)
		}
		if err != nil {
			return err
		}
		return rt.printStatus("downloaded", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "attachmentId": *attachmentID, "path": *out}, "downloaded attachment %d", *attachmentID)
	case "delete":
		fs := newFlagSet("attachment delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		attachmentType := fs.String("type", "", "typed attachment type")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *attachmentID != 0, "attachment delete requires --board --stack --card --attachment"); err != nil {
			return err
		}
		var err error
		if *attachmentType != "" {
			err = rt.client.DeleteAttachmentRef(rt.ctx, *cardID, fmt.Sprintf("%s:%d", *attachmentType, *attachmentID))
		} else {
			err = rt.client.DeleteAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID)
		}
		if err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"boardId": *boardID, "stackId": *stackID, "cardId": *cardID, "attachmentId": *attachmentID}, "deleted attachment %d", *attachmentID)
	case "restore":
		fs := newFlagSet("attachment restore", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		stackID := fs.Int64("stack", 0, "stack id")
		cardID := fs.Int64("card", 0, "card id")
		attachmentID := fs.Int64("attachment", 0, "attachment id")
		attachmentType := fs.String("type", "", "typed attachment type")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *stackID != 0 && *cardID != 0 && *attachmentID != 0, "attachment restore requires --board --stack --card --attachment"); err != nil {
			return err
		}
		var attachment deck.Attachment
		var err error
		if *attachmentType != "" {
			err = rt.client.RestoreAttachmentRef(rt.ctx, *cardID, fmt.Sprintf("%s:%d", *attachmentType, *attachmentID), &attachment)
		} else {
			attachment, err = rt.client.RestoreAttachment(rt.ctx, *boardID, *stackID, *cardID, *attachmentID)
		}
		if err != nil {
			return err
		}
		return rt.printValue(attachment, nil)
	default:
		return fmt.Errorf("unknown attachment command %q", args[0])
	}
}
