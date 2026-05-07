package cli

import "fmt"

func runOverview(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck overview upcoming")
	}
	switch args[0] {
	case "upcoming":
		cards, err := rt.client.UpcomingCards(rt.ctx)
		if err != nil {
			return err
		}
		return rt.printValue(cards, nil)
	default:
		return fmt.Errorf("unknown overview command %q", args[0])
	}
}
