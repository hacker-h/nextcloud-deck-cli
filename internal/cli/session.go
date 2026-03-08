package cli

import "fmt"

func runSession(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck session create|sync|close")
	}
	switch args[0] {
	case "create":
		fs := newFlagSet("session create", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0, "session create requires --board"); err != nil {
			return err
		}
		session, err := rt.client.CreateSession(rt.ctx, *boardID)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, session)
	case "sync", "close":
		fs := newFlagSet("session sync", rt.stderr)
		boardID := fs.Int64("board", 0, "board id")
		token := fs.String("token", "", "session token")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := require(*boardID != 0 && *token != "", fmt.Sprintf("session %s requires --board --token", args[0])); err != nil {
			return err
		}
		if args[0] == "sync" {
			if err := rt.client.SyncSession(rt.ctx, *boardID, *token); err != nil {
				return err
			}
			return printLine(rt.stdout, "synced session for board %d", *boardID)
		}
		if err := rt.client.CloseSession(rt.ctx, *boardID, *token); err != nil {
			return err
		}
		return printLine(rt.stdout, "closed session for board %d", *boardID)
	default:
		return fmt.Errorf("unknown session command %q", args[0])
	}
}
