package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Print(usage)
		return nil
	}
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v" || args[0] == "version") {
		fmt.Printf("daco %s\n", version)
		return nil
	}
	fmt.Print(usage)
	return fmt.Errorf("not implemented: Catalog commands are coming soon")
}

const version = "0.0.0-dev"

const usage = `daco - Daco Catalog CLI

Usage:
  daco --help
  daco --version

Commands will call the remote Catalog API with WorkOS Connect M2M credentials.
`
