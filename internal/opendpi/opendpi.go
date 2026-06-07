// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

const Version = "1.0.0"

type Schema struct {
	jsonschema.Schema
}

func (s Schema) MarshalYAML() (any, error) {
	j, err := json.Marshal(&s.Schema)
	if err != nil {
		return nil, err
	}
	var raw any
	if err := json.Unmarshal(j, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *Schema) UnmarshalYAML(node *yaml.Node) error {
	var raw any
	if err := node.Decode(&raw); err != nil {
		return err
	}
	j, err := json.Marshal(toJSONCompatible(raw))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(j, &s.Schema); err != nil {
		return err
	}
	populatePropertyOrder(&s.Schema, node)
	return nil
}

type Document struct {
	OpenDPI     string                `yaml:"opendpi"`
	Info        Info                  `yaml:"info"`
	Connections map[string]Connection `yaml:"connections"`
	Ports       map[string]Port       `yaml:"ports"`
	Components  map[string]any        `yaml:"components,omitempty"`
}

type Info struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type Connection struct {
	Ref         string         `yaml:"$ref,omitempty"`
	Type        string         `yaml:"type,omitempty"`
	Host        string         `yaml:"host,omitempty"`
	Description string         `yaml:"description,omitempty"`
	Variables   map[string]any `yaml:"variables,omitempty"`
}

func (c Connection) IsRef() bool { return c.Ref != "" }

func (c Connection) Resolve(baseDir string) (*Connection, error) {
	if c.IsRef() {
		path := c.Ref
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		return LoadConnection(path)
	}
	if c.Type == "" && c.Host == "" {
		return nil, errors.New("connection has neither $ref nor inline definition")
	}
	cc := c
	return &cc, nil
}

type Port struct {
	Description string           `yaml:"description,omitempty"`
	Connections []PortConnection `yaml:"connections,omitempty"`
	Schema      *Schema          `yaml:"schema,omitempty"`
}

type PortConnection struct {
	Connection string `yaml:"connection"`
	Location   string `yaml:"location"`
}

func Scaffold(path, title string) error {
	doc := Document{
		OpenDPI: Version,
		Info: Info{
			Title:   title,
			Version: "1.0.0",
		},
		Connections: map[string]Connection{},
		Ports:       map[string]Port{},
	}
	return writeYAML(path, &doc)
}

func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func Save(path string, doc *Document) error {
	return writeYAML(path, doc)
}

func ScaffoldConnection(path string, c Connection) error {
	if c.Variables == nil && !c.IsRef() {
		c.Variables = map[string]any{}
	}
	return writeYAML(path, &c)
}

func LoadConnection(path string) (*Connection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Connection
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func SaveConnection(path string, c *Connection) error {
	return writeYAML(path, c)
}

func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	var raw any
	if err := node.Decode(&raw); err != nil {
		return nil, err
	}
	j, err := json.Marshal(toJSONCompatible(raw))
	if err != nil {
		return nil, err
	}
	var s Schema
	if err := json.Unmarshal(j, &s.Schema); err != nil {
		return nil, err
	}
	populatePropertyOrder(&s.Schema, &node)
	return &s, nil
}

func SaveSchema(path string, s *Schema) error {
	j, err := json.Marshal(&s.Schema)
	if err != nil {
		return err
	}
	var raw any
	if err := json.Unmarshal(j, &raw); err != nil {
		return err
	}
	return writeYAML(path, raw)
}

func SchemaTypeString(s *Schema) string {
	if s == nil {
		return ""
	}
	if s.Type != "" {
		return s.Type
	}
	if len(s.Types) > 0 {
		out := s.Types[0]
		for _, t := range s.Types[1:] {
			out += "|" + t
		}
		return out
	}
	return ""
}

// DefEntry pairs a $defs name with its schema.
type DefEntry struct {
	Name   string
	Schema *jsonschema.Schema
}

// TraverseDefs returns s.Defs in topological order — a def comes after every
// other def it references via "$ref: #/$defs/<other>". Independent defs are
// ordered alphabetically for stable output. Cycles are broken by visiting in
// alphabetical order.
func TraverseDefs(s *jsonschema.Schema) []DefEntry {
	if s == nil || len(s.Defs) == 0 {
		return nil
	}
	names := make([]string, 0, len(s.Defs))
	for name := range s.Defs {
		names = append(names, name)
	}
	sort.Strings(names)

	deps := make(map[string][]string, len(names))
	for _, name := range names {
		seen := map[string]bool{}
		collectDefRefs(s.Defs[name], seen)
		delete(seen, name)
		sub := make([]string, 0, len(seen))
		for d := range seen {
			if _, ok := s.Defs[d]; ok {
				sub = append(sub, d)
			}
		}
		sort.Strings(sub)
		deps[name] = sub
	}

	visited := map[string]bool{}
	var result []DefEntry
	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		for _, d := range deps[name] {
			visit(d)
		}
		result = append(result, DefEntry{Name: name, Schema: s.Defs[name]})
	}
	for _, name := range names {
		visit(name)
	}
	return result
}

func collectDefRefs(s *jsonschema.Schema, out map[string]bool) {
	if s == nil {
		return
	}
	if name, ok := strings.CutPrefix(s.Ref, "#/$defs/"); ok {
		out[name] = true
	}
	for _, sub := range s.Properties {
		collectDefRefs(sub, out)
	}
	for _, sub := range s.Defs {
		collectDefRefs(sub, out)
	}
	if s.Items != nil {
		collectDefRefs(s.Items, out)
	}
	if s.AdditionalProperties != nil {
		collectDefRefs(s.AdditionalProperties, out)
	}
	for _, sub := range s.AllOf {
		collectDefRefs(sub, out)
	}
	for _, sub := range s.AnyOf {
		collectDefRefs(sub, out)
	}
	for _, sub := range s.OneOf {
		collectDefRefs(sub, out)
	}
}

// populatePropertyOrder walks a yaml.Node tree alongside the unmarshaled
// jsonschema.Schema and populates PropertyOrder so the translate package can
// emit fields in source order.
func populatePropertyOrder(s *jsonschema.Schema, node *yaml.Node) {
	if s == nil || node == nil {
		return
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return
	}

	if propsNode := findChild(node, "properties"); propsNode != nil && propsNode.Kind == yaml.MappingNode {
		order := make([]string, 0, len(propsNode.Content)/2)
		for i := 0; i < len(propsNode.Content); i += 2 {
			order = append(order, propsNode.Content[i].Value)
		}
		s.PropertyOrder = order
		for i := 0; i < len(propsNode.Content); i += 2 {
			propName := propsNode.Content[i].Value
			propNode := propsNode.Content[i+1]
			if sub, ok := s.Properties[propName]; ok {
				populatePropertyOrder(sub, propNode)
			}
		}
	}

	if defsNode := findChild(node, "$defs"); defsNode != nil && defsNode.Kind == yaml.MappingNode {
		for i := 0; i < len(defsNode.Content); i += 2 {
			defName := defsNode.Content[i].Value
			defNode := defsNode.Content[i+1]
			if sub, ok := s.Defs[defName]; ok {
				populatePropertyOrder(sub, defNode)
			}
		}
	}

	if itemsNode := findChild(node, "items"); itemsNode != nil && s.Items != nil {
		populatePropertyOrder(s.Items, itemsNode)
	}
	if apNode := findChild(node, "additionalProperties"); apNode != nil && s.AdditionalProperties != nil {
		populatePropertyOrder(s.AdditionalProperties, apNode)
	}
	populateArrayOrder(s.AllOf, findChild(node, "allOf"))
	populateArrayOrder(s.AnyOf, findChild(node, "anyOf"))
	populateArrayOrder(s.OneOf, findChild(node, "oneOf"))
}

func populateArrayOrder(schemas []*jsonschema.Schema, node *yaml.Node) {
	if node == nil || node.Kind != yaml.SequenceNode {
		return
	}
	for i, sub := range schemas {
		if i >= len(node.Content) {
			break
		}
		populatePropertyOrder(sub, node.Content[i])
	}
}

func findChild(node *yaml.Node, key string) *yaml.Node {
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

func writeYAML(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func toJSONCompatible(v any) any {
	switch t := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = toJSONCompatible(val)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = toJSONCompatible(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = toJSONCompatible(item)
		}
		return out
	default:
		return v
	}
}