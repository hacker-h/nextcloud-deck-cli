package cli

import "fmt"

func runSearch(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck search cards")
	}
	switch args[0] {
	case "cards":
		fs := newFlagSet("search cards", rt.stderr)
		term := fs.String("term", "", "search term")
		limit := fs.Int("limit", 20, "result limit")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*term != "", "search cards requires --term"); err != nil {
			return err
		}
		cards, err := rt.client.SearchCards(rt.ctx, *term, *limit)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, cards)
	default:
		return fmt.Errorf("unknown search command %q", args[0])
	}
}
