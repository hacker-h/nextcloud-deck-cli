package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func runConfig(rt *runtime, args []string) error {
	if len(args) == 0 {
		return printLine(rt.stdout, "deck config get|set")
	}
	switch args[0] {
	case "get":
		config, err := rt.client.GetConfig(rt.ctx)
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, config)
	case "set":
		fs := newFlagSet("config set", rt.stderr)
		key := fs.String("key", "", "config key")
		value := fs.String("value", "", "config value")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := rt.client.SetConfig(rt.ctx, *key, coerceValue(*value))
		if err != nil {
			return err
		}
		return printJSON(rt.stdout, result)
	default:
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

func coerceValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	switch strings.ToLower(trimmed) {
	case "true":
		return true
	case "false":
		return false
	}
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	return trimmed
}
