package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

type runtime struct {
	client *deck.Client
	stdout io.Writer
	stderr io.Writer
	ctx    context.Context
	cancel context.CancelFunc
	output outputFormat
}

type outputFormat string

const (
	outputJSON outputFormat = "json"
	outputText outputFormat = "text"
)

func newRuntimeWithTimeout(stdout, stderr io.Writer, output outputFormat, timeoutOverride time.Duration, profile string) (*runtime, error) {
	if profile == "" {
		profile = os.Getenv("DECK_PROFILE")
	}
	cfg, err := config.LoadProfile(profile)
	if err != nil {
		return nil, err
	}
	if timeoutOverride > 0 {
		cfg.Timeout = timeoutOverride
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	return &runtime{
		client: deck.NewClient(cfg),
		stdout: stdout,
		stderr: stderr,
		ctx:    ctx,
		cancel: cancel,
		output: output,
	}, nil
}

func newFlagSet(name string, _ io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func printJSON(out io.Writer, value any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func (rt *runtime) printValue(value any, text func() error) error {
	if rt.output == outputText {
		if text != nil {
			return text()
		}
		return printTextValue(rt.stdout, value)
	}
	return printJSON(rt.stdout, value)
}

func (rt *runtime) printStatus(status string, fields map[string]any, textFormat string, args ...any) error {
	if rt.output == outputText {
		return printLine(rt.stdout, textFormat, args...)
	}
	payload := map[string]any{"status": status}
	maps.Copy(payload, fields)
	return printJSON(rt.stdout, payload)
}

func require(value bool, message string) error {
	if !value {
		return validationError(message)
	}
	return nil
}

func printLine(out io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(out, format+"\n", args...)
	return err
}

func printTextValue(out io.Writer, value any) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil
	}
	v = indirectValue(v)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := printTextRow(out, v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		return printTextMap(out, v)
	case reflect.Struct:
		return printTextRow(out, v)
	default:
		_, err := fmt.Fprintln(out, v.Interface())
		return err
	}
}

func printTextRow(out io.Writer, value reflect.Value) error {
	v := indirectValue(value)
	if !v.IsValid() {
		_, err := fmt.Fprintln(out)
		return err
	}
	switch v.Kind() {
	case reflect.Map:
		return printTextMap(out, v)
	case reflect.Struct:
		values := preferredTextFields(v, []string{"id", "title", "message", "token", "uid", "displayname", "type", "data", "order", "status"})
		if len(values) == 0 {
			values = exportedScalarFields(v)
		}
		_, err := fmt.Fprintln(out, strings.Join(values, "\t"))
		return err
	default:
		_, err := fmt.Fprintln(out, formatTextScalar(v))
		return err
	}
}

func printTextMap(out io.Writer, v reflect.Value) error {
	v = indirectValue(v)
	if !v.IsValid() || v.Kind() != reflect.Map {
		return nil
	}
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})
	for _, key := range keys {
		value := indirectValue(v.MapIndex(key))
		if _, err := fmt.Fprintf(out, "%v\t%s\n", key.Interface(), formatTextScalar(value)); err != nil {
			return err
		}
	}
	return nil
}

func preferredTextFields(v reflect.Value, names []string) []string {
	var values []string
	for _, name := range names {
		if field, ok := fieldByJSONName(v, name); ok {
			value := formatTextScalar(indirectValue(field))
			if value != "" {
				values = append(values, value)
			}
		}
	}
	return values
}

func exportedScalarFields(v reflect.Value) []string {
	var values []string
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := t.Field(i)
		if fieldInfo.PkgPath != "" {
			continue
		}
		field := indirectValue(v.Field(i))
		if !isTextScalar(field) {
			continue
		}
		value := formatTextScalar(field)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func fieldByJSONName(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := t.Field(i)
		if fieldInfo.PkgPath != "" {
			continue
		}
		jsonName := strings.Split(fieldInfo.Tag.Get("json"), ",")[0]
		if jsonName == "" {
			jsonName = fieldInfo.Name
		}
		if jsonName == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func indirectValue(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func isTextScalar(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

func formatTextScalar(v reflect.Value) string {
	v = indirectValue(v)
	if !v.IsValid() {
		return ""
	}
	if isTextScalar(v) {
		return fmt.Sprint(v.Interface())
	}
	data, err := json.Marshal(v.Interface())
	if err != nil {
		return fmt.Sprint(v.Interface())
	}
	return string(data)
}
