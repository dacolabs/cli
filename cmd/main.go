// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package main is the entry point for the daco CLI.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dacolabs/cli/cmd/internal"
)

func main() {
	if err := internal.Run(context.Background(), os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
