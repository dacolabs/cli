// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package internal contains the main application logic for the CLI.
package internal

import (
	"context"

	"github.com/dacolabs/cli/internal/commands"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/cli/internal/translate/pyspark"
	"github.com/dacolabs/cli/internal/translate/python"
	"github.com/dacolabs/cli/internal/translate/scala"
)

func registerTranslators() translate.Register {
	translators := make(translate.Register)
	translators["pyspark"] = &pyspark.Translator{}
	translators["python"] = &python.Translator{}
	translators["scala"] = &scala.Translator{}
	translators["spark-scala"] = &pyspark.Translator{}
	return translators
}

// Run is the main application logic, extracted for testability.
// It accepts OS dependencies as parameters (context, env lookup).
func Run(ctx context.Context, getenv func(string) string) error {
	translators := registerTranslators()
	rootCmd := commands.NewRootCmd(translators)
	return rootCmd.ExecuteContext(ctx)
}
