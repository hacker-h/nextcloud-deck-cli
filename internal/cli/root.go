package cli

import (
	"fmt"
	"io"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) error {
	var err error
	args, output, err := parseOutputArgs(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage(stdout)
		return nil
	}
	rt, err := newRuntime(stdout, stderr, output)
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

func printUsage(out io.Writer) {
	_, _ = io.WriteString(out, strings.TrimSpace(`deck <command>

Output:
  --json, --text, -o json|text, --output json|text
                         Output format. Defaults to text.

Commands:
  board      list|get|create|update|archive|unarchive|clone|export|import|delete|restore|import-systems|import-schema
  list       list|get|archived|create|rename|reorder|delete
  card       list|get|create|clone|delete|move|reorder|archive|unarchive|done|undone|rename|describe|due|assign-user|unassign-user|assign-label|remove-label
  search     cards
  overview   upcoming
  session    create|sync|close
  capabilities get
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

func parseOutputArgs(args []string) ([]string, outputFormat, error) {
	output := outputText
	cleaned := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			output = outputJSON
		case arg == "--text":
			output = outputText
		case arg == "-o" || arg == "--output":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("%s requires a value", arg)
			}
			parsed, err := parseOutputFormat(args[i+1])
			if err != nil {
				return nil, "", err
			}
			output = parsed
			i++
		case strings.HasPrefix(arg, "-o="):
			parsed, err := parseOutputFormat(strings.TrimPrefix(arg, "-o="))
			if err != nil {
				return nil, "", err
			}
			output = parsed
		case strings.HasPrefix(arg, "--output="):
			parsed, err := parseOutputFormat(strings.TrimPrefix(arg, "--output="))
			if err != nil {
				return nil, "", err
			}
			output = parsed
		default:
			cleaned = append(cleaned, arg)
		}
	}
	return cleaned, output, nil
}

func parseOutputFormat(raw string) (outputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "json":
		return outputJSON, nil
	case "text", "table":
		return outputText, nil
	default:
		return "", fmt.Errorf("unsupported output format %q; supported formats: json, text", raw)
	}
}
