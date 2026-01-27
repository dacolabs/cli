// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

// PortAddResult holds the result of RunAddPortForm.
type PortAddResult struct {
	Name        string
	Description string
	Schema      *jsonschema.Schema
	SchemaPath  string
	Connections []opendpi.PortConnection
	NewConns    map[string]opendpi.Connection
}

// RunAddPortForm runs the interactive form for adding a port.
// Returns PortAddResult containing port name, description, schema, schemaPath (if file selected),
// connections, and newConns (connections created during this form that need to be added to spec).
func RunAddPortForm(
	existingPorts map[string]opendpi.Port,
	existingConns map[string]opendpi.Connection,
	existingSchemas map[string]*jsonschema.Schema,
) (result PortAddResult, _ error) {
	// Step 1: Port name and description
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Port name").
				Placeholder("e.g., user_events").
				Value(&result.Name).
				Validate(identifierValidator(existingPorts)),
			huh.NewInput().
				Title("Port description (optional)").
				Placeholder("e.g., Stream of user activity events").
				Value(&result.Description),
		),
	).Run(); err != nil {
		return result, err
	}

	// Step 2: Schema (optional)
	var hasSchema bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Define a schema?").
				Value(&hasSchema),
		),
	).Run(); err != nil {
		return result, err
	}

	if hasSchema {
		var schemaSource string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Schema source").
					Options(
						huh.NewOption("Create interactively", "interactive"),
						huh.NewOption("Load from file", "file"),
					).
					Value(&schemaSource),
			),
		).Run(); err != nil {
			return result, err
		}

		switch schemaSource {
		case "file":
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Schema file path").
						Placeholder("e.g., ./schemas/event.json").
						Value(&result.SchemaPath).
						Validate(requiredValidator("schema file path")),
				),
			).Run(); err != nil {
				return result, err
			}
		case "interactive":
			result.Schema = &jsonschema.Schema{}
			schemas := copySchemas(existingSchemas)
			if err := RunSchemaForm(nil, result.Schema, schemas); err != nil {
				return result, err
			}
			// Attach collected $defs (new schemas created during form)
			for k, v := range schemas {
				if _, existed := existingSchemas[k]; !existed {
					if result.Schema.Defs == nil {
						result.Schema.Defs = make(map[string]*jsonschema.Schema)
					}
					result.Schema.Defs[k] = v
				}
			}
		}
	}

	// Step 3: Connections (optional)
	var hasConnection bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add a connection?").
				Value(&hasConnection),
		),
	).Run(); err != nil {
		return result, err
	}

	if hasConnection {
		var err error
		result.Connections, result.NewConns, err = RunPortConnectionsForm(existingConns)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

// RunPortConnectionsForm handles adding connections to a port, allowing both
// creating new connections and using existing ones.
func RunPortConnectionsForm(existingConns map[string]opendpi.Connection) ([]opendpi.PortConnection, map[string]opendpi.Connection, error) {
	var conns []opendpi.PortConnection
	newConns := make(map[string]opendpi.Connection)

	// Make a working copy of existing connections
	workingConns := make(map[string]opendpi.Connection)
	for k, v := range existingConns {
		workingConns[k] = v
	}

	for {
		var connSource string
		options := []huh.Option[string]{
			huh.NewOption("Create new connection", "new"),
		}
		if len(workingConns) > 0 {
			options = append(options, huh.NewOption("Use existing connection", "existing"))
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Connection source").
					Options(options...).
					Value(&connSource),
			),
		).Run(); err != nil {
			return nil, nil, err
		}

		var connResult ConnectionAddResult
		var err error

		switch connSource {
		case "new":
			connResult, err = RunAddNewConnectionForm(workingConns)
			if err != nil {
				return nil, nil, err
			}
			// Add to working connections so it can be reused
			workingConns[connResult.Name] = connResult.Connection
			// Track as new connection to be added to spec
			newConns[connResult.Name] = connResult.Connection

		case "existing":
			connResult, err = RunSelectConnectionForm(workingConns)
			if err != nil {
				return nil, nil, err
			}
		}

		// Prompt for location
		var location string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Location").
					Placeholder("e.g., events_topic, users_table, /data/events").
					Value(&location).
					Validate(requiredValidator("location")),
			),
		).Run(); err != nil {
			return nil, nil, err
		}

		connCopy := connResult.Connection
		conns = append(conns, opendpi.PortConnection{
			Connection: &connCopy,
			Location:   location,
		})

		var addAnother bool
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add another connection?").
					Affirmative("Yes").
					Negative("No").
					Value(&addAnother),
			),
		).Run(); err != nil {
			return nil, nil, err
		}

		if !addAnother {
			break
		}
	}

	return conns, newConns, nil
}

// RunTranslateFormatSelect returns a select field for choosing translation output format.
func RunTranslateFormatSelect(value *string, formats []string) *huh.Select[string] {
	options := make([]huh.Option[string], len(formats))
	for i, f := range formats {
		options[i] = huh.NewOption(f, f)
	}
	return huh.NewSelect[string]().
		Title("Output format").
		Options(options...).
		Value(value)
}
