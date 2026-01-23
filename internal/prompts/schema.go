// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/jschema"
)

// RunSchemaForm prompts the user to define or select a schema.
// If name is nil, the schema name prompt is skipped (used for nested schemas like properties/items).
func RunSchemaForm(name *string, schema *jschema.Schema, schemas map[string]*jschema.Schema) error {
	if name != nil && *name == "" {
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Schema name").
					Placeholder("e.g., User").
					Value(name).
					Validate(identifierValidator(schemas)),
			),
		).Run(); err != nil {
			return err
		}
	}

	if len(schemas) > 0 {
		var useExisting bool
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Use an existing schema?").
					Affirmative("Yes").
					Negative("No, create new").
					Value(&useExisting),
			),
		).Run(); err != nil {
			return err
		}

		if useExisting {
			options := make([]huh.Option[string], 0, len(schemas))
			for name := range schemas {
				options = append(options, huh.NewOption(name, "#/$defs/"+name))
			}
			return huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select existing schema").
						Options(options...).
						Filtering(true).
						Value(&schema.Ref).
						Height(8),
				),
			).Run()
		}
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Schema type").
				Options(
					huh.NewOption("object", "object"),
					huh.NewOption("array", "array"),
					huh.NewOption("string", "string"),
					huh.NewOption("number", "number"),
					huh.NewOption("integer", "integer"),
					huh.NewOption("boolean", "boolean"),
				).
				Value(&schema.Type),
			huh.NewInput().
				Title("Description (optional)").
				Placeholder("e.g., User profile information").
				Value(&schema.Description),
		),
	).Run(); err != nil {
		return err
	}

	switch schema.Type {
	case "string":
		if err := promptStringAttributes(schema); err != nil {
			return err
		}

	case "number", "integer":
		if err := promptNumberAttributes(schema); err != nil {
			return err
		}

	case "object":
		schema.Properties = make(map[string]*jschema.Schema)
		for {
			var propName string
			var isRequired bool
			var addAnother bool

			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Property name").
						Placeholder("e.g., id").
						Value(&propName).
						Validate(identifierValidator(schema.Properties)),
					huh.NewConfirm().
						Title("Is this property required?").
						Affirmative("Yes").
						Negative("No").
						Value(&isRequired),
				),
			).Run(); err != nil {
				return err
			}

			// Recursive call for property schema (nil name - nested schema)
			propSchema := &jschema.Schema{}
			if err := RunSchemaForm(nil, propSchema, schemas); err != nil {
				return err
			}

			// If property is a complex type (object or array of objects), store in schemas and use ref
			switch {
			case propSchema.Type == "object":
				schemas[propName] = propSchema
				schema.Properties[propName] = &jschema.Schema{Ref: "#/$defs/" + propName}
			case propSchema.Type == "array" && propSchema.Items != nil && propSchema.Items.Type == "object":
				// Array of objects - store the array schema in $defs
				schemas[propName] = propSchema
				schema.Properties[propName] = &jschema.Schema{Ref: "#/$defs/" + propName}
			default:
				schema.Properties[propName] = propSchema
			}

			if isRequired {
				schema.Required = append(schema.Required, propName)
			}

			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Add another property?").
						Affirmative("Yes").
						Negative("No").
						Value(&addAnother),
				),
			).Run(); err != nil {
				return err
			}

			if !addAnother {
				break
			}
		}

	case "array":
		itemSchema := &jschema.Schema{}
		if err := RunSchemaForm(nil, itemSchema, schemas); err != nil {
			return err
		}
		schema.Items = itemSchema
	}

	return nil
}

func promptStringAttributes(schema *jschema.Schema) error {
	var format string
	var enumValues string

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Format (optional)").
				Options(
					huh.NewOption("None", ""),
					huh.NewOption("date-time", "date-time"),
					huh.NewOption("date", "date"),
					huh.NewOption("time", "time"),
					huh.NewOption("email", "email"),
					huh.NewOption("uri", "uri"),
					huh.NewOption("uuid", "uuid"),
				).
				Value(&format),
			huh.NewInput().
				Title("Allowed values (comma-separated, optional)").
				Placeholder("e.g., active,inactive,pending").
				Value(&enumValues),
		),
	).Run(); err != nil {
		return err
	}

	if format != "" {
		schema.Format = format
	}
	if enumValues != "" {
		parts := strings.Split(enumValues, ",")
		schema.Enum = make([]any, len(parts))
		for i, p := range parts {
			schema.Enum[i] = strings.TrimSpace(p)
		}
	}

	return nil
}

func promptNumberAttributes(schema *jschema.Schema) error {
	var minStr, maxStr string

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Minimum value (optional)").
				Placeholder("e.g., 0").
				Value(&minStr).
				Validate(optionalNumberValidator),
			huh.NewInput().
				Title("Maximum value (optional)").
				Placeholder("e.g., 100").
				Value(&maxStr).
				Validate(optionalNumberValidator),
		),
	).Run(); err != nil {
		return err
	}

	if minStr != "" {
		if val, err := strconv.ParseFloat(minStr, 64); err == nil {
			schema.Minimum = &val
		}
	}
	if maxStr != "" {
		if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
			schema.Maximum = &val
		}
	}

	return nil
}

func identifierValidator[T any](existing map[string]T) func(string) error {
	return func(s string) error {
		if s == "" {
			return errors.New("name is required")
		}
		for i, r := range s {
			if i == 0 && !unicode.IsLetter(r) && r != '_' {
				return errors.New("must start with letter or underscore")
			}
			if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return errors.New("must contain only letters, numbers, underscores")
			}
		}
		if _, exists := existing[s]; exists {
			return fmt.Errorf("%q already exists", s)
		}
		return nil
	}
}

func requiredValidator(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func optionalNumberValidator(s string) error {
	if s == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return errors.New("must be a valid number")
	}
	return nil
}

func copySchemas(src map[string]*jschema.Schema) map[string]*jschema.Schema {
	if src == nil {
		return make(map[string]*jschema.Schema)
	}
	dst := make(map[string]*jschema.Schema, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
