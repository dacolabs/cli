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
			ref := strings.TrimPrefix(rpc.Connection, "#/connections/")
			conn, ok := connections[ref]
			if !ok {
				return nil, fmt.Errorf("port %q: connection %q not found", name, rpc.Connection)
			}
			portConns[i] = PortConnection{
				Connection: &conn,
				Location:   rpc.Location,
			}
		}

		var portSchema *jsonschema.Schema
		var schemaRef string
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
					schemaRef = rp.Schema.Ref
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
			SchemaRef:   schemaRef,
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
	Type        string         `yaml:"type" json:"type"`
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
	Connection string `yaml:"connection" json:"connection"`
	Location   string `yaml:"location" json:"location"`
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

	setInlineSchemaOrderJSON(data, &raw)

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

	setInlineSchemaOrderYAML(data, &raw)

	return &raw, nil
}

// setInlineSchemaOrderJSON extracts property order from raw JSON spec bytes
// and sets PropertyOrder on inline port schemas and component schemas.
func setInlineSchemaOrderJSON(data []byte, raw *rawSpec) {
	var docMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &docMap); err != nil {
		return
	}

	// Extract port schemas
	if portsRaw, ok := docMap["ports"]; ok {
		var ports map[string]json.RawMessage
		if err := json.Unmarshal(portsRaw, &ports); err == nil {
			for portName, portRaw := range ports {
				var portObj map[string]json.RawMessage
				if err := json.Unmarshal(portRaw, &portObj); err == nil {
					if schemaRaw, ok := portObj["schema"]; ok {
						if rp, ok := raw.Ports[portName]; ok && rp.Schema != nil {
							keyOrder, err := jschema.ExtractKeyOrderFromJSON(schemaRaw)
							if err == nil {
								jschema.SetPropertyOrder(rp.Schema, keyOrder)
							}
						}
					}
				}
			}
		}
	}

	// Extract component schemas
	if componentsRaw, ok := docMap["components"]; ok {
		var components map[string]json.RawMessage
		if err := json.Unmarshal(componentsRaw, &components); err == nil {
			if schemasRaw, ok := components["schemas"]; ok {
				var schemas map[string]json.RawMessage
				if err := json.Unmarshal(schemasRaw, &schemas); err == nil {
					for schemaName, schemaRaw := range schemas {
						if raw.Components != nil {
							if s, ok := raw.Components.Schemas[schemaName]; ok {
								keyOrder, err := jschema.ExtractKeyOrderFromJSON(schemaRaw)
								if err == nil {
									jschema.SetPropertyOrder(s, keyOrder)
								}
							}
						}
					}
				}
			}
		}
	}
}

// setInlineSchemaOrderYAML extracts property order from raw YAML spec bytes
// and sets PropertyOrder on inline port schemas and component schemas.
func setInlineSchemaOrderYAML(data []byte, raw *rawSpec) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(doc.Content); i += 2 {
		keyNode := doc.Content[i]
		valueNode := doc.Content[i+1]

		switch keyNode.Value {
		case "ports":
			if valueNode.Kind != yaml.MappingNode {
				continue
			}
			for j := 0; j < len(valueNode.Content); j += 2 {
				portName := valueNode.Content[j].Value
				portNode := valueNode.Content[j+1]
				if portNode.Kind != yaml.MappingNode {
					continue
				}
				schemaNode := findYAMLMappingKey(portNode, "schema")
				if schemaNode != nil {
					rp, ok := raw.Ports[portName]
					if ok && rp.Schema != nil {
						keyOrder := make(map[string][]string)
						jschema.ExtractYAMLNodeKeyOrder(schemaNode, "", keyOrder)
						jschema.SetPropertyOrder(rp.Schema, keyOrder)
					}
				}
			}
		case "components":
			if valueNode.Kind != yaml.MappingNode {
				continue
			}
			schemasNode := findYAMLMappingKey(valueNode, "schemas")
			if schemasNode == nil || schemasNode.Kind != yaml.MappingNode {
				continue
			}
			for j := 0; j < len(schemasNode.Content); j += 2 {
				schemaName := schemasNode.Content[j].Value
				schemaValNode := schemasNode.Content[j+1]
				if raw.Components != nil {
					if s, ok := raw.Components.Schemas[schemaName]; ok {
						keyOrder := make(map[string][]string)
						jschema.ExtractYAMLNodeKeyOrder(schemaValNode, "", keyOrder)
						jschema.SetPropertyOrder(s, keyOrder)
					}
				}
			}
		}
	}
}

// findYAMLMappingKey finds the value node for a given key in a YAML mapping node.
func findYAMLMappingKey(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
