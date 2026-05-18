package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runShare(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck share list|permissions|create|update|delete|leave|transfer-owner")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("share list", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "share list requires --board"); err != nil {
			return err
		}
		rules, err := rt.client.ListShares(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return rt.printValue(rules, nil)
	case "permissions":
		fs := newFlagSet("share permissions", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "share permissions requires --board"); err != nil {
			return err
		}
		permissions, err := rt.client.GetBoardPermissions(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return rt.printValue(permissions, nil)
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
		if err := require(*boardID != 0 && *participant != "", "share create requires --board --participant"); err != nil {
			return err
		}
		rules, err := rt.client.CreateShare(rt.ctx, *boardID, deck.CreateACLRuleRequest{Type: *participantType, Participant: *participant, PermissionEdit: *edit, PermissionShare: *share, PermissionManage: *manage})
		if err != nil {
			return err
		}
		return rt.printValue(rules, nil)
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
		if err := require(*boardID != 0 && *shareID != 0, "share update requires --board --share-id"); err != nil {
			return err
		}
		if err := rt.client.UpdateShare(rt.ctx, *boardID, *shareID, deck.UpdateACLRuleRequest{PermissionEdit: *edit, PermissionShare: *sharePerm, PermissionManage: *manage}); err != nil {
			return err
		}
		return rt.printStatus("updated", map[string]any{"boardId": *boardID, "shareId": *shareID}, "updated share %d", *shareID)
	case "leave":
		fs := newFlagSet("share leave", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "share leave requires --board"); err != nil {
			return err
		}
		if err := rt.client.LeaveBoard(rt.ctx, *boardID); err != nil {
			return err
		}
		return rt.printStatus("left", map[string]any{"boardId": *boardID}, "left board %d", *boardID)
	case "transfer-owner":
		fs := newFlagSet("share transfer-owner", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		newOwner := fs.String("new-owner", "", "new owner user id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *newOwner != "", "share transfer-owner requires --board --new-owner"); err != nil {
			return err
		}
		result, err := rt.client.TransferBoardOwner(rt.ctx, *boardID, *newOwner)
		if err != nil {
			return err
		}
		return rt.printValue(result, nil)
	case "delete":
		fs := newFlagSet("share delete", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		shareID := fs.Int64("share-id", 0, "share id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *shareID != 0, "share delete requires --board --share-id"); err != nil {
			return err
		}
		if err := rt.client.DeleteShare(rt.ctx, *boardID, *shareID); err != nil {
			return err
		}
		return rt.printStatus("deleted", map[string]any{"boardId": *boardID, "shareId": *shareID}, "deleted share %d", *shareID)
	default:
		return fmt.Errorf("unknown share command %q", args[0])
	}
}
