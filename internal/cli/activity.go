package cli

import (
	"fmt"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func runActivity(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck activity list|card")
	}
	switch args[0] {
	case "list":
		fs := newFlagSet("activity list", rt.stderr)
		objectType := fs.String("object-type", "deck_card", "activity object type")
		objectID := fs.Int64("object-id", 0, "activity object id")
		limit := fs.Int("limit", 50, "result limit")
		since := fs.Int64("since", -1, "activity since cursor")
		sort := fs.String("sort", "asc", "sort order")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		items, err := rt.client.GetActivity(rt.ctx, deck.ActivityQuery{ObjectType: *objectType, ObjectID: *objectID, Limit: *limit, Since: *since, Sort: *sort})
		if err != nil {
			return err
		}
		return rt.printValue(items, nil)
	case "card":
		fs := newFlagSet("activity card", rt.stderr)
		cardID := fs.Int64("card", 0, "card id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*cardID != 0, "activity card requires --card"); err != nil {
			return err
		}
		items, err := rt.client.GetCardActivity(rt.ctx, *cardID)
		if err != nil {
			return err
		}
		return rt.printValue(items, nil)
	default:
		return fmt.Errorf("unknown activity command %q", args[0])
	}
}
