// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"path"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
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

	var schemaDefs map[string]*jsonschema.Schema
	if raw.Components != nil {
		schemaDefs = make(map[string]*jsonschema.Schema, len(raw.Components.Schemas))
		for name, rs := range raw.Components.Schemas {
			if rs.Ref != "" && !strings.HasPrefix(rs.Ref, "#/") {
				schema, err := loadSchemaFile(fsys, rs.Ref)
				if err != nil {
					return nil, fmt.Errorf("component schema %q: failed to load external schema %q: %w", name, rs.Ref, err)
				}
				basePath := path.Dir(rs.Ref)
				if err := loadExternalSchemas(fsys, schema, basePath); err != nil {
					return nil, fmt.Errorf("component schema %q: %w", name, err)
				}
				schemaDefs[name] = schema
			} else {
				if err := loadExternalSchemas(fsys, rs, ""); err != nil {
					return nil, fmt.Errorf("component schema %q: %w", name, err)
				}
				schemaDefs[name] = rs
			}
		}
	} else {
		schemaDefs = make(map[string]*jsonschema.Schema)
	}

	for name, rp := range raw.Ports {
		if rp.Schema == nil || rp.Schema.Ref == "" || strings.HasPrefix(rp.Schema.Ref, "#/") {
			continue
		}
		schema, err := loadSchemaFile(fsys, rp.Schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("port %q: failed to load external schema %q: %w", name, rp.Schema.Ref, err)
		}
		basePath := path.Dir(rp.Schema.Ref)
		if err := loadExternalSchemas(fsys, schema, basePath); err != nil {
			return nil, fmt.Errorf("port %q: %w", name, err)
		}
		schemaDefs[rp.Schema.Ref] = schema
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
					schema = resolved
				} else {
					// External file ref - already loaded in schemaDefs with full path
					resolved, ok := schemaDefs[rp.Schema.Ref]
					if !ok {
						return nil, fmt.Errorf("port %q: external schema %q not found", name, rp.Schema.Ref)
					}
					schema = resolved
				}
			} else {
				schema = rp.Schema
				// Load external file refs inside inline schema
				if err := loadExternalSchemas(fsys, schema, ""); err != nil {
					return nil, fmt.Errorf("port %q: %w", name, err)
				}
			}

			// Only add component schemas that are actually referenced
			if raw.Components != nil && len(schemaDefs) > 0 {
				collectComponentRefs := func(schema *jsonschema.Schema, schemaDefs map[string]*jsonschema.Schema) map[string]struct{} {
					refs := make(map[string]struct{})
					resolver := func(ref string) *jsonschema.Schema {
						if schemaName, ok := strings.CutPrefix(ref, "#/components/schemas/"); ok {
							return schemaDefs[schemaName]
						}
						return nil
					}
					for s := range iterNestedRefs(schema, resolver) {
						if schemaName, ok := strings.CutPrefix(s.Ref, "#/components/schemas/"); ok {
							refs[schemaName] = struct{}{}
						}
					}
					return refs
				}
				refs := collectComponentRefs(schema, schemaDefs)
				if len(refs) > 0 {
					if schema.Defs == nil {
						schema.Defs = make(map[string]*jsonschema.Schema)
					}
					for ref := range refs {
						s, ok := schemaDefs[ref]
						if !ok {
							return nil, fmt.Errorf("port %q: referenced schema %q not found in components", name, ref)
						}
						if _, ok := schema.Defs[ref]; ok {
							return nil, fmt.Errorf("port %q: schema definition conflict: %q already exists in $defs", name, ref)
						}
						schema.Defs[ref] = s
					}
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

	return &raw, nil
}

func loadSchemaFile(fsys fs.FS, filePath string) (*jsonschema.Schema, error) {
	f, err := fsys.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var schema jsonschema.Schema

	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		var intermediate any
		if err := yaml.Unmarshal(data, &intermediate); err != nil {
			return nil, err
		}
		jsonData, err := json.Marshal(intermediate)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(jsonData, &schema); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, err
		}
	}

	return &schema, nil
}

func iterNestedRefs(schema *jsonschema.Schema, resolver func(ref string) *jsonschema.Schema) iter.Seq[*jsonschema.Schema] {
	return func(yield func(*jsonschema.Schema) bool) {
		visited := make(map[*jsonschema.Schema]struct{})
		refsWithVisited(schema, resolver, yield, visited)
	}
}

func refsWithVisited(schema *jsonschema.Schema, resolver func(ref string) *jsonschema.Schema, yield func(*jsonschema.Schema) bool, visited map[*jsonschema.Schema]struct{}) bool {
	if schema == nil {
		return true
	}
	if _, ok := visited[schema]; ok {
		return true
	}
	visited[schema] = struct{}{}

	if !yield(schema) {
		return false
	}

	// Follow $ref if resolver is provided
	if schema.Ref != "" && resolver != nil {
		if resolved := resolver(schema.Ref); resolved != nil {
			if !refsWithVisited(resolved, resolver, yield, visited) {
				return false
			}
		}
	}

	// Objects
	for _, s := range schema.Properties {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.PatternProperties {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !refsWithVisited(schema.AdditionalProperties, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.PropertyNames, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.UnevaluatedProperties, resolver, yield, visited) {
		return false
	}

	// Arrays
	if !refsWithVisited(schema.Items, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.ItemsArray {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.PrefixItems {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !refsWithVisited(schema.AdditionalItems, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.Contains, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.UnevaluatedItems, resolver, yield, visited) {
		return false
	}

	// Logic
	for _, s := range schema.AllOf {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.AnyOf {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.OneOf {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	if !refsWithVisited(schema.Not, resolver, yield, visited) {
		return false
	}

	// Conditional
	if !refsWithVisited(schema.If, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.Then, resolver, yield, visited) {
		return false
	}
	if !refsWithVisited(schema.Else, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.DependentSchemas {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.DependencySchemas {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}

	// Other
	if !refsWithVisited(schema.ContentSchema, resolver, yield, visited) {
		return false
	}
	for _, s := range schema.Defs {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}
	for _, s := range schema.Definitions {
		if !refsWithVisited(s, resolver, yield, visited) {
			return false
		}
	}

	return true
}

func loadExternalSchemas(fsys fs.FS, schema *jsonschema.Schema, basePath string) error {
	for s := range iterNestedRefs(schema, nil) {
		if s.Ref == "" || strings.HasPrefix(s.Ref, "#/") {
			continue
		}
		refPath := path.Join(basePath, s.Ref)
		loaded, err := loadSchemaFile(fsys, refPath)
		if err != nil {
			return fmt.Errorf("failed to load external schema %q: %w", s.Ref, err)
		}
		newBase := path.Dir(refPath)
		if err := loadExternalSchemas(fsys, loaded, newBase); err != nil {
			return err
		}
		*s = *loaded
	}
	return nil
}
