package cli

import (
	"errors"
	"flag"
	"io"
	"strings"
	"time"
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
		usage:              "deck board list|get|find|create|update|archive|unarchive|clone|export|import|import-server|delete|restore|import-systems|import-schema",
		requiresSubcommand: true,
		unknownLabel:       "board",
		subcommands: map[string]commandHelp{
			"list": {}, "get": {}, "create": {}, "update": {}, "archive": {}, "unarchive": {}, "clone": {}, "export": {}, "import": {}, "import-server": {}, "delete": {}, "restore": {}, "import-systems": {}, "import-schema": {},
			"find": {usage: "deck board find --title TEXT"},
		},
	},
	"list": {
		usage:              "deck list list|get|find|archived|create|rename|reorder|done|undone|delete",
		requiresSubcommand: true,
		unknownLabel:       "list",
		subcommands: map[string]commandHelp{
			"list": {}, "get": {}, "archived": {}, "create": {}, "rename": {}, "reorder": {}, "done": {}, "undone": {}, "delete": {},
			"find": {usage: "deck list find --board ID --title TEXT"},
		},
	},
	"card": {
		usage:              "deck card list|get|deleted|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|update|due|assign-user|unassign-user|assign-label|remove-label|assign-dependent|remove-dependent",
		requiresSubcommand: true,
		unknownLabel:       "card",
		subcommands: map[string]commandHelp{
			"list": {}, "get": {}, "deleted": {}, "create": {}, "clone": {}, "delete": {}, "move": {}, "reorder": {}, "archive": {}, "unarchive": {}, "done": {}, "undone": {}, "rename": {}, "describe": {}, "update": {}, "assign-user": {}, "unassign-user": {}, "assign-label": {}, "remove-label": {}, "assign-dependent": {}, "remove-dependent": {},
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
		usage:              "deck activity list|card",
		requiresSubcommand: true,
		unknownLabel:       "activity",
		subcommands:        knownSubcommands("list", "card"),
	},
	"todo": {
		usage:              "deck todo list|add|check|uncheck",
		requiresSubcommand: true,
		unknownLabel:       "todo",
		subcommands:        knownSubcommands("list", "add", "check", "uncheck"),
	},
	"label": {
		usage:              "deck label list|get|find|create|update|delete",
		requiresSubcommand: true,
		unknownLabel:       "label",
		subcommands: map[string]commandHelp{
			"list": {}, "get": {}, "create": {}, "update": {}, "delete": {},
			"find": {usage: "deck label find --board ID --title TEXT"},
		},
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
		usage:              "deck share list|permissions|create|update|delete|leave|transfer-owner",
		requiresSubcommand: true,
		unknownLabel:       "share",
		subcommands:        knownSubcommands("list", "permissions", "create", "update", "delete", "leave", "transfer-owner"),
	},
	"config": {
		usage:              "deck config get|set",
		requiresSubcommand: true,
		unknownLabel:       "config",
		subcommands:        knownSubcommands("get", "set"),
	},
	"auth": {
		usage:              "deck auth setup|profiles",
		requiresSubcommand: true,
		unknownLabel:       "auth",
		subcommands: map[string]commandHelp{
			"setup":    {usage: "deck auth setup [--profile NAME]"},
			"profiles": {usage: "deck auth profiles"},
		},
	},
}

func Run(args []string, stdout, stderr io.Writer) error {
	var err error
	var output outputFormat
	var timeoutOverride time.Duration
	var profile string
	args, output, timeoutOverride, profile, err = parseGlobalArgs(args)
	if err != nil {
		return err
	}
	if handled, err := handleBootstrap(args, stdout); handled {
		return err
	}
	if isAuthSetupCommand(args) {
		return runAuthSetup(args[2:], stdout, profile)
	}
	if isAuthProfilesCommand(args) {
		return runAuthProfiles(args[2:], stdout, output)
	}
	rt, err := newRuntimeWithTimeout(stdout, stderr, output, timeoutOverride, profile)
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
	case "auth":
		return runAuth(rt, args[1:])
	default:
		return validationf("unknown command %q", args[0])
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
		return true, validationf("unknown command %q", args[0])
	}
	return handleCommandBootstrap(stdout, []string{args[0]}, command, args[1:])
}

func handleCommandBootstrap(stdout io.Writer, path []string, command commandHelp, args []string) (bool, error) {
	if len(args) == 0 {
		if command.requiresSubcommand {
			if command.missingMessage != "" {
				return true, validationError(command.missingMessage)
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
		return true, validationf("unknown %s command %q", command.unknownLabel, args[0])
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
		return validationf("unknown command %q", path[0])
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

Output:
  --json, --text, -o json|text, --output json|text
                         Output format. Defaults to text.

Timeout:
  --timeout DURATION      Request timeout. Defaults to 90s or DECK_TIMEOUT.

Profiles:
  --profile NAME          Saved auth profile. Overrides DECK_PROFILE. "default" uses the flat config.
  deck auth profiles      Lists profiles without printing app passwords.

Commands:
  board      list|get|find|create|update|archive|unarchive|clone|export|import|import-server|delete|restore|import-systems|import-schema
  list       list|get|find|archived|create|rename|reorder|done|undone|delete
  card       list|get|deleted|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|update|due|assign-user|unassign-user|assign-label|remove-label|assign-dependent|remove-dependent
  search     cards
  overview   upcoming
  session    create|sync|close
  capabilities
  user       search|get
  activity   list|card
  todo       list|add|check|uncheck
  label      list|get|find|create|update|delete
  comment    list|create|update|delete
  attachment list|upload|download|delete|restore
  share      list|permissions|create|update|delete|leave|transfer-owner
  config     get|set
  auth       setup|profiles
`)+"\n")
}

func parseGlobalArgs(args []string) ([]string, outputFormat, time.Duration, string, error) {
	output := outputText
	var timeoutOverride time.Duration
	var profile string
	cleaned := make([]string, 0, len(args))
	commandSeen := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			output = outputJSON
		case arg == "--text" && !(commandSeen && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-")):
			output = outputText
		case arg == "-o" || arg == "--output":
			if i+1 >= len(args) {
				return nil, "", 0, "", validationf("%s requires a value", arg)
			}
			parsed, err := parseOutputFormat(args[i+1])
			if err != nil {
				return nil, "", 0, "", err
			}
			output = parsed
			i++
		case strings.HasPrefix(arg, "-o="):
			parsed, err := parseOutputFormat(strings.TrimPrefix(arg, "-o="))
			if err != nil {
				return nil, "", 0, "", err
			}
			output = parsed
		case strings.HasPrefix(arg, "--output="):
			parsed, err := parseOutputFormat(strings.TrimPrefix(arg, "--output="))
			if err != nil {
				return nil, "", 0, "", err
			}
			output = parsed
		case arg == "--timeout":
			if i+1 >= len(args) {
				return nil, "", 0, "", validationf("%s requires a value", arg)
			}
			parsed, err := parseTimeoutFlag(args[i+1])
			if err != nil {
				return nil, "", 0, "", err
			}
			timeoutOverride = parsed
			i++
		case strings.HasPrefix(arg, "--timeout="):
			parsed, err := parseTimeoutFlag(strings.TrimPrefix(arg, "--timeout="))
			if err != nil {
				return nil, "", 0, "", err
			}
			timeoutOverride = parsed
		case arg == "--profile":
			if i+1 >= len(args) {
				return nil, "", 0, "", validationf("%s requires a value", arg)
			}
			parsed := strings.TrimSpace(args[i+1])
			if parsed == "" {
				return nil, "", 0, "", validationError("--profile requires a non-empty value")
			}
			profile = parsed
			i++
		case strings.HasPrefix(arg, "--profile="):
			parsed := strings.TrimSpace(strings.TrimPrefix(arg, "--profile="))
			if parsed == "" {
				return nil, "", 0, "", validationError("--profile requires a non-empty value")
			}
			profile = parsed
		default:
			cleaned = append(cleaned, arg)
			if !strings.HasPrefix(arg, "-") {
				commandSeen = true
			}
		}
	}
	return cleaned, output, timeoutOverride, profile, nil
}

func parseOutputFormat(raw string) (outputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "json":
		return outputJSON, nil
	case "text", "table":
		return outputText, nil
	default:
		return "", validationf("unsupported output format %q; supported formats: json, text", raw)
	}
}

func parseTimeoutFlag(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, validationf("invalid timeout %q: %v", raw, err)
	}
	if timeout <= 0 {
		return 0, validationf("invalid timeout %q: must be greater than 0", raw)
	}
	return timeout, nil
}
