package cli

import (
	"fmt"
	"io"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage(stdout)
		return nil
	}
	rt, err := newRuntime(stdout, stderr)
	if err != nil {
		return err
	}
	defer rt.cancel()

	switch args[0] {
	case "board":
		return runBoard(rt, args[1:])
	case "list":
		return runList(rt, args[1:])
	case "card":
		return runCard(rt, args[1:])
	case "todo":
		return runTodo(rt, args[1:])
	case "label":
		return runLabel(rt, args[1:])
	case "comment":
		return runComment(rt, args[1:])
	case "attachment":
		return runAttachment(rt, args[1:])
	case "share":
		return runShare(rt, args[1:])
	case "config":
		return runConfig(rt, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(out io.Writer) {
	_, _ = io.WriteString(out, strings.TrimSpace(`deck <command>

Commands:
  board      list|get|create|update|archive|unarchive|delete|restore
  list       list|get|archived|create|rename|reorder|delete
  card       list|get|create|delete|move|reorder|archive|unarchive|rename|describe|due|assign-user|unassign-user|assign-label|remove-label
  todo       list|add|check|uncheck
  label      list|get|create|update|delete
  comment    list|create|update|delete
  attachment list|upload|download|delete|restore
  share      list|create|update|delete
  config     get|set
`)+"\n")
}
