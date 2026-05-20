package app

import (
	"context"
	"fmt"
	"os"

	"github.com/dacolabs/daco/internal/cli"
	"github.com/dacolabs/daco/internal/translate"
	"github.com/dacolabs/daco/internal/translate/avro"
	"github.com/dacolabs/daco/internal/translate/databrickspyspark"
	"github.com/dacolabs/daco/internal/translate/databricksscala"
	"github.com/dacolabs/daco/internal/translate/databrickssql"
	"github.com/dacolabs/daco/internal/translate/dqxyaml"
	"github.com/dacolabs/daco/internal/translate/gotypes"
	"github.com/dacolabs/daco/internal/translate/markdown"
	"github.com/dacolabs/daco/internal/translate/protobuf"
	"github.com/dacolabs/daco/internal/translate/pydantic"
	"github.com/dacolabs/daco/internal/translate/pyspark"
	"github.com/dacolabs/daco/internal/translate/python"
	"github.com/dacolabs/daco/internal/translate/scala"
	"github.com/dacolabs/daco/internal/translate/sparkscala"
	"github.com/dacolabs/daco/internal/translate/sparksql"
)


func registerTranslators() translate.Register {
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
	translators["markdown"] = &markdown.Translator{}
	return translators
}

// Run is the main application logic, extracted for testability.
// It accepts OS dependencies as parameters (context, env lookup).
func Run(ctx context.Context, getenv func(string) string) error {
	translators := registerTranslators()
	
	// Build root command and execute
	rootCmd := cli.NewRootCmd(translators)
	rootCmd.SetArgs(os.Args[1:])
	rootCmd.SetIn(os.Stdin)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}

	return nil
}
