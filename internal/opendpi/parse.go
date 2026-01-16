// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// Parser decodes an OpenDPI spec from an io.Reader.
type Parser struct {
	parse func(io.Reader) (*Spec, error)
}

var (
	// JSON parses OpenDPI specs from JSON.
	JSON = Parser{parseJSON}
	// YAML parses OpenDPI specs from YAML.
	YAML = Parser{parseYAML}
)

// Parse decodes an OpenDPI spec from r and resolves all $ref references.
func (p Parser) Parse(r io.Reader) (*Spec, error) {
	return p.parse(r)
}

type rawSpec struct {
	OpenDPI     string                   `yaml:"opendpi" json:"opendpi"`
	Info        rawInfo                  `yaml:"info" json:"info"`
	Tags        []rawTag                 `yaml:"tags,omitempty" json:"tags,omitempty"`
	Connections map[string]rawConnection `yaml:"connections" json:"connections"`
	Ports       map[string]rawPorts      `yaml:"ports" json:"ports"`
	Components  *rawComponents           `yaml:"components,omitempty" json:"components,omitempty"`
}

type rawInfo struct {
	Title       string `yaml:"title" json:"title"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type rawTag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type rawConnection struct {
	Protocol    string         `yaml:"protocol" json:"protocol"`
	Host        string         `yaml:"host" json:"host"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
}

type rawPorts struct {
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	Connections []rawPortConnection `yaml:"connections" json:"connections"`
	Schema      *jsonschema.Schema  `yaml:"schema" json:"schema"`
}

type rawPortConnection struct {
	Connection connectionRef `yaml:"connection" json:"connection"`
	Location   string        `yaml:"location" json:"location"`
}

type connectionRef struct {
	Ref string `yaml:"$ref" json:"$ref"`
}

type rawComponents struct {
	Schemas map[string]*jsonschema.Schema `yaml:"schemas,omitempty" json:"schemas,omitempty"`
}

func parseJSON(r io.Reader) (*Spec, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var raw rawSpec
	if err = json.Unmarshal(data, &raw); err != nil { //nolint:gocritic
		return nil, err
	}

	return resolveRefs(&raw)
}

func parseYAML(r io.Reader) (*Spec, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var intermediate map[string]any
	if err = yaml.Unmarshal(data, &intermediate); err != nil { //nolint:gocritic
		return nil, err
	}

	jsonData, err := json.Marshal(intermediate)
	if err != nil {
		return nil, err
	}

	var raw rawSpec
	if err = json.Unmarshal(jsonData, &raw); err != nil { //nolint:gocritic
		return nil, err
	}

	return resolveRefs(&raw)
}

func resolveRefs(raw *rawSpec) (*Spec, error) {
	connections := make(map[string]Connection, len(raw.Connections))
	for name, rc := range raw.Connections {
		connections[name] = Connection(rc)
	}

	tags := make([]Tag, len(raw.Tags))
	for i, rt := range raw.Tags {
		tags[i] = Tag(rt)
	}

	ports := make(map[string]Port, len(raw.Ports))
	for name, rp := range raw.Ports {
		portConns := make([]PortConnection, len(rp.Connections))
		for i, rpc := range rp.Connections {
			ref := strings.TrimPrefix(rpc.Connection.Ref, "#/connections/")
			conn, ok := connections[ref]
			if !ok {
				return nil, fmt.Errorf("port %q: connection %q not found", name, rpc.Connection.Ref)
			}
			portConns[i] = PortConnection{
				Connection: &conn,
				Location:   rpc.Location,
			}
		}

		var schema *jsonschema.Schema
		if rp.Schema != nil {
			if rp.Schema.Ref != "" {
				if raw.Components == nil {
					return nil, fmt.Errorf("port %q: schema %q not found (no components)", name, rp.Schema.Ref)
				}
				ref := strings.TrimPrefix(rp.Schema.Ref, "#/components/schemas/")
				resolved, ok := raw.Components.Schemas[ref]
				if !ok {
					return nil, fmt.Errorf("port %q: schema %q not found", name, rp.Schema.Ref)
				}
				schema = resolved
			} else {
				schema = rp.Schema
			}

			if raw.Components != nil && len(raw.Components.Schemas) > 0 {
				if schema.Defs == nil {
					schema.Defs = make(map[string]*jsonschema.Schema)
				}
				for schemaName, s := range raw.Components.Schemas {
					if _, ok := schema.Defs[schemaName]; ok {
						return nil, fmt.Errorf("port %q: schema definition conflict: %q already exists in $defs", name, schemaName)
					}
					schema.Defs[schemaName] = s
				}
			}
		}

		ports[name] = Port{
			Description: rp.Description,
			Connections: portConns,
			Schema:      schema,
		}
	}

	return &Spec{
		OpenDPI:     raw.OpenDPI,
		Info:        Info(raw.Info),
		Tags:        tags,
		Connections: connections,
		Ports:       ports,
		rawSpec:     raw,
	}, nil
}
