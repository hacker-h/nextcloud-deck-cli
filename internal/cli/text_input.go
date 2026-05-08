package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var commandStdin io.Reader = os.Stdin

type textInputOptions struct {
	directName string
	direct     string
	fileNames  []string
	fileValues []*string
	stdinNames []string
	stdinFlags []*bool
}

func addTextInputFlags(fs *flag.FlagSet, directName, fileName, stdinName, usage string, bodyAliases bool) *textInputOptions {
	opt := &textInputOptions{directName: directName}
	fs.StringVar(&opt.direct, directName, "", usage)
	opt.addFileFlag(fs, fileName, usage+" file")
	opt.addStdinFlag(fs, stdinName, "read "+usage+" from stdin")
	if bodyAliases {
		opt.addFileFlag(fs, "body-file", "text body file")
		opt.addStdinFlag(fs, "body-stdin", "read text body from stdin")
	}
	return opt
}

func (opt *textInputOptions) addFileFlag(fs *flag.FlagSet, name, usage string) {
	value := new(string)
	opt.fileNames = append(opt.fileNames, name)
	opt.fileValues = append(opt.fileValues, value)
	fs.StringVar(value, name, "", usage)
}

func (opt *textInputOptions) addStdinFlag(fs *flag.FlagSet, name, usage string) {
	value := new(bool)
	opt.stdinNames = append(opt.stdinNames, name)
	opt.stdinFlags = append(opt.stdinFlags, value)
	fs.BoolVar(value, name, false, usage)
}

func (opt *textInputOptions) resolve(fs *flag.FlagSet) (string, bool, error) {
	seen := visitedFlags(fs)
	var selected []string
	if seen[opt.directName] {
		selected = append(selected, "--"+opt.directName)
	}
	for i, name := range opt.fileNames {
		if seen[name] {
			if *opt.fileValues[i] == "" {
				return "", false, fmt.Errorf("--%s requires a path", name)
			}
			selected = append(selected, "--"+name)
		}
	}
	for i, name := range opt.stdinNames {
		if *opt.stdinFlags[i] {
			selected = append(selected, "--"+name)
		}
	}
	if len(selected) == 0 {
		return "", false, nil
	}
	if len(selected) > 1 {
		return "", false, fmt.Errorf("choose only one text source: %s", strings.Join(selected, ", "))
	}
	if selected[0] == "--"+opt.directName {
		return opt.direct, true, nil
	}
	for i, name := range opt.fileNames {
		if selected[0] == "--"+name {
			data, err := os.ReadFile(*opt.fileValues[i])
			if err != nil {
				return "", false, fmt.Errorf("read --%s: %w", name, err)
			}
			return string(data), true, nil
		}
	}
	data, err := io.ReadAll(commandStdin)
	if err != nil {
		return "", false, fmt.Errorf("read stdin: %w", err)
	}
	return string(data), true, nil
}

func visitedFlags(fs *flag.FlagSet) map[string]bool {
	seen := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})
	return seen
}
