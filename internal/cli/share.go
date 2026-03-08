package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runShare(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck share list|create|update|delete")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("share list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		rules, err := rt.client.ListShares(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, rules)
	case "create":
		fs := newFlagSet("share create", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		participantType := fs.Int("type", 0, "participant type")
		participant := fs.String("participant", "", "uid or group")
		edit := fs.Bool("edit", true, "edit permission")
		share := fs.Bool("share", false, "share permission")
		manage := fs.Bool("manage", false, "manage permission")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		rules, err := rt.client.CreateShare(rt.ctx, *boardID, deck.CreateACLRuleRequest{Type: *participantType, Participant: *participant, PermissionEdit: *edit, PermissionShare: *share, PermissionManage: *manage})
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, rules)
	case "update":
		fs := newFlagSet("share update", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		shareID := fs.Int64("share-id", 0, "share id")
		edit := fs.Bool("edit", true, "edit permission")
		sharePerm := fs.Bool("share", false, "share permission")
		manage := fs.Bool("manage", false, "manage permission")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := rt.client.UpdateShare(rt.ctx, *boardID, *shareID, deck.UpdateACLRuleRequest{PermissionEdit: *edit, PermissionShare: *sharePerm, PermissionManage: *manage}); err != nil {
			return err
		}
		return printLine(rt.stdout, "updated share %d", *shareID)
	case "delete":
		fs := newFlagSet("share delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		shareID := fs.Int64("share-id", 0, "share id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := rt.client.DeleteShare(rt.ctx, *boardID, *shareID); err != nil {
			return err
		}
		return printLine(rt.stdout, "deleted share %d", *shareID)
	default:
		return fmt.Errorf("unknown share command %q", args[0])
	}
}
