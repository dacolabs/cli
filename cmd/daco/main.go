// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dacolabs/daco/cmd/daco/app"
	"github.com/dacolabs/daco/internal/cli/engine"
)

func main() {
	if err := app.Run(context.Background(), os.Getenv); err != nil {
		writeJSONErrors(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func writeJSONErrors(err error) {
	var list engine.Errors
	if !errors.As(err, &list) {
		list = engine.Errors{err.Error()}
	}
	data, jerr := json.MarshalIndent(list, "", "  ")
	if jerr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}
	fmt.Fprintln(os.Stderr, string(data))
}