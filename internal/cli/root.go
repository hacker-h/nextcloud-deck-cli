package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type commandHelp struct {
	usage              string
	requiresSubcommand bool
	unknownLabel       string
	subcommands        map[string]commandHelp
	missingMessage     string
}

var helpCommands = map[string]commandHelp{
	"board": {
		usage:              "deck board list|get|create|update|archive|unarchive|clone|export|import|delete|restore|import-systems|import-schema",
		requiresSubcommand: true,
		unknownLabel:       "board",
		subcommands:        knownSubcommands("list", "get", "create", "update", "archive", "unarchive", "clone", "export", "import", "delete", "restore", "import-systems", "import-schema"),
	},
	"list": {
		usage:              "deck list list|get|archived|create|rename|reorder|delete",
		requiresSubcommand: true,
		unknownLabel:       "list",
		subcommands:        knownSubcommands("list", "get", "archived", "create", "rename", "reorder", "delete"),
	},
	"card": {
		usage:              "deck card list|get|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|due|assign-user|unassign-user|assign-label|remove-label",
		requiresSubcommand: true,
		unknownLabel:       "card",
		subcommands: map[string]commandHelp{
			"list": {}, "get": {}, "create": {}, "clone": {}, "delete": {}, "move": {}, "reorder": {}, "archive": {}, "unarchive": {}, "done": {}, "undone": {}, "rename": {}, "describe": {}, "assign-user": {}, "unassign-user": {}, "assign-label": {}, "remove-label": {},
			"due": {
				usage:              "deck card due get|set|clear",
				requiresSubcommand: true,
				unknownLabel:       "card due",
				missingMessage:     "card due requires get, set, or clear",
				subcommands:        knownSubcommands("get", "set", "clear"),
			},
		},
	},
	"search": {
		usage:              "deck search cards",
		requiresSubcommand: true,
		unknownLabel:       "search",
		subcommands:        knownSubcommands("cards"),
	},
	"overview": {
		usage:              "deck overview upcoming",
		requiresSubcommand: true,
		unknownLabel:       "overview",
		subcommands:        knownSubcommands("upcoming"),
	},
	"session": {
		usage:              "deck session create|sync|close",
		requiresSubcommand: true,
		unknownLabel:       "session",
		subcommands:        knownSubcommands("create", "sync", "close"),
	},
	"capabilities": {usage: "deck capabilities", unknownLabel: "capabilities", subcommands: knownSubcommands()},
	"user": {
		usage:              "deck user search|get",
		requiresSubcommand: true,
		unknownLabel:       "user",
		subcommands:        knownSubcommands("search", "get"),
	},
	"activity": {
		usage:              "deck activity card",
		requiresSubcommand: true,
		unknownLabel:       "activity",
		subcommands:        knownSubcommands("card"),
	},
	"todo": {
		usage:              "deck todo list|add|check|uncheck",
		requiresSubcommand: true,
		unknownLabel:       "todo",
		subcommands:        knownSubcommands("list", "add", "check", "uncheck"),
	},
	"label": {
		usage:              "deck label list|get|create|update|delete",
		requiresSubcommand: true,
		unknownLabel:       "label",
		subcommands:        knownSubcommands("list", "get", "create", "update", "delete"),
	},
	"comment": {
		usage:              "deck comment list|create|update|delete",
		requiresSubcommand: true,
		unknownLabel:       "comment",
		subcommands:        knownSubcommands("list", "create", "update", "delete"),
	},
	"attachment": {
		usage:              "deck attachment list|upload|download|delete|restore",
		requiresSubcommand: true,
		unknownLabel:       "attachment",
		subcommands:        knownSubcommands("list", "upload", "download", "delete", "restore"),
	},
	"share": {
		usage:              "deck share list|create|update|delete",
		requiresSubcommand: true,
		unknownLabel:       "share",
		subcommands:        knownSubcommands("list", "create", "update", "delete"),
	},
	"config": {
		usage:              "deck config get|set",
		requiresSubcommand: true,
		unknownLabel:       "config",
		subcommands:        knownSubcommands("get", "set"),
	},
}

func Run(args []string, stdout, stderr io.Writer) error {
	if handled, err := handleBootstrap(args, stdout); handled {
		return err
	}
	rt, err := newRuntime(stdout, stderr)
	if err != nil {
		return err
	}
	defer rt.cancel()
	if err := dispatch(rt, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return nil
}

func dispatch(rt *runtime, args []string) error {
	switch args[0] {
	case "board":
		return runBoard(rt, args[1:])
	case "list":
		return runList(rt, args[1:])
	case "card":
		return runCard(rt, args[1:])
	case "search":
		return runSearch(rt, args[1:])
	case "overview":
		return runOverview(rt, args[1:])
	case "session":
		return runSession(rt, args[1:])
	case "capabilities":
		return runCapabilities(rt, args[1:])
	case "user":
		return runUser(rt, args[1:])
	case "activity":
		return runActivity(rt, args[1:])
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

func handleBootstrap(args []string, stdout io.Writer) (bool, error) {
	if len(args) == 0 || isHelpArg(args[0]) {
		printUsage(stdout)
		return true, nil
	}
	if args[0] == "help" {
		return true, printHelpPath(stdout, args[1:])
	}
	command, ok := helpCommands[args[0]]
	if !ok {
		return true, fmt.Errorf("unknown command %q", args[0])
	}
	return handleCommandBootstrap(stdout, []string{args[0]}, command, args[1:])
}

func handleCommandBootstrap(stdout io.Writer, path []string, command commandHelp, args []string) (bool, error) {
	if len(args) == 0 {
		if command.requiresSubcommand {
			if command.missingMessage != "" {
				return true, errors.New(command.missingMessage)
			}
			return true, printLine(stdout, command.usage)
		}
		return false, nil
	}
	if isHelpArg(args[0]) || args[0] == "help" {
		return true, printLine(stdout, command.usage)
	}
	subcommand, ok := command.subcommands[args[0]]
	if command.subcommands != nil && !ok {
		return true, fmt.Errorf("unknown %s command %q", command.unknownLabel, args[0])
	}
	if len(args) > 1 && (isHelpArg(args[1]) || args[1] == "help") {
		usage := subcommand.usage
		if usage == "" {
			usage = "deck " + strings.Join(append(path, args[0]), " ")
		}
		return true, printLine(stdout, usage)
	}
	if subcommand.requiresSubcommand {
		return handleCommandBootstrap(stdout, append(path, args[0]), subcommand, args[1:])
	}
	return false, nil
}

func printHelpPath(stdout io.Writer, path []string) error {
	if len(path) == 0 {
		printUsage(stdout)
		return nil
	}
	command, ok := helpCommands[path[0]]
	if !ok {
		return fmt.Errorf("unknown command %q", path[0])
	}
	if len(path) == 1 {
		return printLine(stdout, command.usage)
	}
	_, err := handleCommandBootstrap(stdout, []string{path[0]}, command, append(path[1:], "help"))
	return err
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func knownSubcommands(names ...string) map[string]commandHelp {
	subcommands := make(map[string]commandHelp, len(names))
	for _, name := range names {
		subcommands[name] = commandHelp{}
	}
	return subcommands
}

func printUsage(out io.Writer) {
	_, _ = io.WriteString(out, strings.TrimSpace(`deck <command>

Commands:
  board      list|get|create|update|archive|unarchive|clone|export|import|delete|restore|import-systems|import-schema
  list       list|get|archived|create|rename|reorder|delete
  card       list|get|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|due|assign-user|unassign-user|assign-label|remove-label
  search     cards
  overview   upcoming
  session    create|sync|close
  capabilities
  user       search|get
  activity   card
  todo       list|add|check|uncheck
  label      list|get|create|update|delete
  comment    list|create|update|delete
  attachment list|upload|download|delete|restore
  share      list|create|update|delete
  config     get|set
`)+"\n")
}
