package main

import (
	"os"

	"github.com/hacker-h/nextcloud-deck-api/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
