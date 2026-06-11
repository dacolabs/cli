// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"fmt"

	"github.com/dacolabs/daco/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "T.DateType()"
		case "date-time":
			return "T.TimestampType()"
		case "uuid":
			return "T.StringType()"
		}
	}

	switch schemaType {
	case "string":
		return "T.StringType()"
	case "integer":
		return "T.LongType()"
	case "number":
		return "T.DoubleType()"
	case "boolean":
		return "T.BooleanType()"
	default:
		return "T.StringType()"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("T.ArrayType(%s)", elemType)
}

func (r *resolver) MapType(keyType, valueType string) string {
	return fmt.Sprintf("T.MapType(%s, %s)", keyType, valueType)
}

func (r *resolver) RefType(defName string) string {
	return "_" + translate.ToSnakeCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return "_" + translate.ToSnakeCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return portName + "_schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	switch f.Type {
	case "T.DoubleType()":
		if kind, shape := translate.NarrowNumber(f.Constraints); kind == translate.NumberDecimal {
			f.Type = fmt.Sprintf("T.DecimalType(%d, %d)", shape.Precision, shape.Scale)
		}
	case "T.LongType()":
		f.Type = sparkIntType(translate.NarrowInteger(f.Constraints))
	}

	if meta := translate.ConstraintsPyDict(f.Constraints); meta != "" {
		f.Tag = ", metadata=" + meta
	}
}

func sparkIntType(k translate.IntKind) string {
	switch k {
	case translate.Int8:
		return "T.ByteType()"
	case translate.Int16:
		return "T.ShortType()"
	case translate.Int32:
		return "T.IntegerType()"
	default:
		return "T.LongType()"
	}
}
