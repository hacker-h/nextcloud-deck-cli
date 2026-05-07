package cli

import "fmt"

func runActivity(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck activity card")
	}
	switch args[0] {
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
