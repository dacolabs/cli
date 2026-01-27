// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// Parser decodes an OpenDPI spec from an io.Reader.
type Parser struct {
	parse func(io.Reader) (*rawSpec, error)
}

var (
	// JSON parses OpenDPI specs from JSON.
	JSON = Parser{parseJSON}
	// YAML parses OpenDPI specs from YAML.
	YAML = Parser{parseYAML}
)

// Parse decodes an OpenDPI spec from r and resolves all $ref references.
// External file references are not supported; use ParseFS for that.
func (p Parser) Parse(r io.Reader, fsys fs.FS) (*Spec, error) {
	if fsys == nil {
		return nil, errors.New("fsys is required to resolve external schema references")
	}
	raw, err := p.parse(r)
	if err != nil {
		return nil, err
	}

	connections := make(map[string]Connection, len(raw.Connections))
	for name, rc := range raw.Connections {
		connections[name] = Connection(rc)
	}

	tags := make([]Tag, len(raw.Tags))
	for i, rt := range raw.Tags {
		tags[i] = Tag(rt)
	}

	loader := jschema.NewLoader(fsys)

	var schemaDefs map[string]*jsonschema.Schema
	if raw.Components != nil {
		schemaDefs = make(map[string]*jsonschema.Schema, len(raw.Components.Schemas))
		for name, rs := range raw.Components.Schemas {
			if jschema.IsFileRef(rs.Ref) {
				s, err := loader.LoadFile(rs.Ref)
				if err != nil {
					return nil, fmt.Errorf("component schema %q: failed to load external schema %q: %w", name, rs.Ref, err)
				}
				basePath := path.Dir(rs.Ref)
				if err := loader.ResolveRefs(s, basePath); err != nil {
					return nil, fmt.Errorf("component schema %q: %w", name, err)
				}
				schemaDefs[name] = s
			} else {
				if err := loader.ResolveRefs(rs, ""); err != nil {
					return nil, fmt.Errorf("component schema %q: %w", name, err)
				}
				schemaDefs[name] = rs
			}
		}
	} else {
		schemaDefs = make(map[string]*jsonschema.Schema)
	}

	for name, rp := range raw.Ports {
		if rp.Schema == nil || !jschema.IsFileRef(rp.Schema.Ref) {
			continue
		}
		s, err := loader.LoadFile(rp.Schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("port %q: failed to load external schema %q: %w", name, rp.Schema.Ref, err)
		}
		basePath := path.Dir(rp.Schema.Ref)
		if err := loader.ResolveRefs(s, basePath); err != nil {
			return nil, fmt.Errorf("port %q: %w", name, err)
		}
		schemaDefs[rp.Schema.Ref] = s
	}

	// Unified schemas map, it will contain all resolved schemas keyed by name
	schemas := make(map[string]*jsonschema.Schema)

	// Add component schemas to unified map (already resolved)
	for name, s := range schemaDefs {
		// Skip file path keys (they contain '/' or '.')
		if strings.Contains(name, "/") || strings.Contains(name, ".") {
			continue
		}
		schemas[name] = s
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

		var portSchema *jsonschema.Schema
		if rp.Schema != nil {
			if rp.Schema.Ref != "" {
				if strings.HasPrefix(rp.Schema.Ref, "#/") {
					// Component ref
					if raw.Components == nil {
						return nil, fmt.Errorf("port %q: schema %q not found (no components)", name, rp.Schema.Ref)
					}
					ref := strings.TrimPrefix(rp.Schema.Ref, "#/components/schemas/")
					resolved, ok := schemaDefs[ref]
					if !ok {
						return nil, fmt.Errorf("port %q: schema %q not found", name, rp.Schema.Ref)
					}
					portSchema = resolved
				} else {
					// External file ref - already loaded in schemaDefs with full path
					resolved, ok := schemaDefs[rp.Schema.Ref]
					if !ok {
						return nil, fmt.Errorf("port %q: external schema %q not found", name, rp.Schema.Ref)
					}
					portSchema = resolved
					// Add to unified schemas map using port name
					if existing, exists := schemas[name]; exists && existing != portSchema {
						return nil, fmt.Errorf("duplicate schema name %q", name)
					}
					schemas[name] = portSchema
				}
			} else {
				portSchema = rp.Schema
				// Load external file refs inside inline schema
				if err := loader.ResolveRefs(portSchema, ""); err != nil {
					return nil, fmt.Errorf("port %q: %w", name, err)
				}
				// Add inline schema to unified map using port name
				if existing, exists := schemas[name]; exists && existing != portSchema {
					return nil, fmt.Errorf("duplicate schema name %q", name)
				}
				schemas[name] = portSchema
			}

			// Only add component schemas that are actually referenced
			if raw.Components != nil && len(schemaDefs) > 0 {
				collectComponentRefs := func(s *jsonschema.Schema, schemaDefs map[string]*jsonschema.Schema) map[string]struct{} {
					refs := make(map[string]struct{})
					resolver := func(ref string) *jsonschema.Schema {
						if schemaName, ok := strings.CutPrefix(ref, "#/components/schemas/"); ok {
							return schemaDefs[schemaName]
						}
						return nil
					}
					for item := range jschema.Traverse(s, resolver) {
						if schemaName, ok := strings.CutPrefix(item.Ref, "#/components/schemas/"); ok {
							refs[schemaName] = struct{}{}
						}
					}
					return refs
				}
				refs := collectComponentRefs(portSchema, schemaDefs)
				if len(refs) > 0 {
					if portSchema.Defs == nil {
						portSchema.Defs = make(map[string]*jsonschema.Schema)
					}
					for ref := range refs {
						s, ok := schemaDefs[ref]
						if !ok {
							return nil, fmt.Errorf("port %q: referenced schema %q not found in components", name, ref)
						}
						if _, ok := portSchema.Defs[ref]; ok {
							return nil, fmt.Errorf("port %q: schema definition conflict: %q already exists in $defs", name, ref)
						}
						portSchema.Defs[ref] = s
					}
					// Rewrite refs from #/components/schemas/ and #/definitions/ to #/$defs/
					jschema.RewriteRefs(portSchema)
				}
			}

			// Validate $defs don't conflict with existing schema names
			for defName := range portSchema.Defs {
				if existing, exists := schemas[defName]; exists && existing != portSchema.Defs[defName] {
					return nil, fmt.Errorf("port %q: schema definition %q conflicts with existing schema", name, defName)
				}
			}
		}

		ports[name] = Port{
			Description: rp.Description,
			Connections: portConns,
			Schema:      portSchema,
		}
	}

	return &Spec{
		OpenDPI:     raw.OpenDPI,
		Info:        Info(raw.Info),
		Tags:        tags,
		Connections: connections,
		Ports:       ports,
		Schemas:     schemas,
	}, nil
}

type rawSpec struct {
	OpenDPI     string                   `yaml:"opendpi" json:"opendpi"`
	Info        rawInfo                  `yaml:"info" json:"info"`
	Tags        []rawTag                 `yaml:"tags,omitempty" json:"tags,omitempty"`
	Connections map[string]rawConnection `yaml:"connections" json:"connections"`
	Ports       map[string]rawPort       `yaml:"ports" json:"ports"`
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

type rawPort struct {
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

func parseJSON(r io.Reader) (*rawSpec, error) {
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

	return &raw, nil
}

func parseYAML(r io.Reader) (*rawSpec, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var raw rawSpec
	if err = yaml.Unmarshal(data, &raw); err != nil { //nolint:gocritic
		return nil, err
	}

	return &raw, nil
}
