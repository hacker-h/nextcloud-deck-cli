package cli

import "fmt"

func runUser(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck user search|get")
	}
	switch args[0] {
	case "search":
		fs := newFlagSet("user search", rt.stderr)
		term := fs.String("term", "", "search term")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		results, err := rt.client.SearchSharees(rt.ctx, *term)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, results)
	case "get":
		fs := newFlagSet("user get", rt.stderr)
		userID := fs.String("user", "", "user id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		user, err := rt.client.GetUser(rt.ctx, *userID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, user)
	default:
		return fmt.Errorf("unknown user command %q", args[0])
	}
}
