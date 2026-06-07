// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package registry

import (
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

func Default() translate.Register {
	return translate.Register{
		"avro":               &avro.Translator{},
		"databricks-pyspark": &databrickspyspark.Translator{},
		"databricks-scala":   &databricksscala.Translator{},
		"databricks-sql":     &databrickssql.Translator{},
		"dqx-yaml":           &dqxyaml.Translator{},
		"gotypes":            &gotypes.Translator{},
		"markdown":           &markdown.Translator{},
		"protobuf":           &protobuf.Translator{},
		"pydantic":           &pydantic.Translator{},
		"pyspark":            &pyspark.Translator{},
		"python":             &python.Translator{},
		"scala":              &scala.Translator{},
		"spark-scala":        &sparkscala.Translator{},
		"spark-sql":          &sparksql.Translator{},
	}
}