// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Command gendocs generates LLM-friendly markdown documentation for the daco CLI.
//
// Usage:
//
//	go run ./cmd/gendocs [output-dir]
//
// Default output directory is ./docs/cli.
package main

import (
	"fmt"
	"os"

	"github.com/dacolabs/cli/internal/commands"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/cli/internal/translate/avro"
	"github.com/dacolabs/cli/internal/translate/databrickspyspark"
	"github.com/dacolabs/cli/internal/translate/databricksscala"
	"github.com/dacolabs/cli/internal/translate/databrickssql"
	"github.com/dacolabs/cli/internal/translate/dqxyaml"
	"github.com/dacolabs/cli/internal/translate/gotypes"
	"github.com/dacolabs/cli/internal/translate/protobuf"
	"github.com/dacolabs/cli/internal/translate/pydantic"
	"github.com/dacolabs/cli/internal/translate/pyspark"
	"github.com/dacolabs/cli/internal/translate/python"
	"github.com/dacolabs/cli/internal/translate/scala"
	"github.com/dacolabs/cli/internal/translate/sparkscala"
	"github.com/dacolabs/cli/internal/translate/sparksql"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "./docs/cli"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	translators := make(translate.Register)
	translators["pyspark"] = &pyspark.Translator{}
	translators["gotypes"] = &gotypes.Translator{}
	translators["avro"] = &avro.Translator{}
	translators["databricks-pyspark"] = &databrickspyspark.Translator{}
	translators["pydantic"] = &pydantic.Translator{}
	translators["python"] = &python.Translator{}
	translators["databricks-sql"] = &databrickssql.Translator{}
	translators["scala"] = &scala.Translator{}
	translators["spark-scala"] = &sparkscala.Translator{}
	translators["databricks-scala"] = &databricksscala.Translator{}
	translators["protobuf"] = &protobuf.Translator{}
	translators["spark-sql"] = &sparksql.Translator{}
	translators["dqx-yaml"] = &dqxyaml.Translator{}

	rootCmd := commands.NewRootCmd(translators)
	rootCmd.DisableAutoGenTag = true

	if err := os.MkdirAll(dir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output dir: %v\n", err)
		os.Exit(1)
	}

	if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating docs: %v\n", err)
		os.Exit(1)
	}

	// Rename daco.md to index.md
	oldPath := dir + "/daco.md"
	newPath := dir + "/index.md"
	if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error renaming %s to %s: %v\n", oldPath, newPath, err)
		os.Exit(1)
	}

	fmt.Printf("Documentation generated in %s\n", dir)
}
