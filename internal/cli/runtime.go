package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

const timeout = 90 * time.Second

type runtime struct {
	client *deck.Client
	stdout io.Writer
	stderr io.Writer
	ctx    context.Context
	cancel context.CancelFunc
}

func newRuntime(stdout, stderr io.Writer) (*runtime, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &runtime{
		client: deck.NewClient(cfg),
		stdout: stdout,
		stderr: stderr,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func printJSON(out io.Writer, value any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func require(value bool, message string) error {
	if !value {
		return errors.New(message)
	}
	return nil
}

func printLine(out io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(out, format+"\n", args...)
	return err
}
