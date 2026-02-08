package opendpi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// Writer encodes an OpenDPI spec to a file.
type Writer struct {
	write     func(path string, v any) error
	extension string
}

var (
	// JSONWriter writes OpenDPI specs as JSON.
	JSONWriter = Writer{writeJSON, ".json"}
	// YAMLWriter writes OpenDPI specs as YAML.
	YAMLWriter = Writer{writeYaml, ".yaml"}
)

// Write encodes the spec to the given specDir as opendpi.<ext>.
func (wr Writer) Write(spec *Spec, specDir string) error {
	raw, err := toRaw(spec)
	if err != nil {
		return err
	}
	specPath := filepath.Join(specDir, "opendpi"+wr.extension)
	return wr.write(specPath, raw)
}

func toRaw(spec *Spec) (*rawSpec, error) {
	connections := make(map[string]rawConnection, len(spec.Connections))
	for name, c := range spec.Connections {
		connections[name] = rawConnection(c)
	}

	tags := make([]rawTag, len(spec.Tags))
	for i, t := range spec.Tags {
		tags[i] = rawTag(t)
	}

	// Build reverse lookup: *Connection → name
	connNames := make(map[*Connection]string)
	for name := range spec.Connections {
		c := spec.Connections[name]
		connNames[&c] = name
	}

	ports := make(map[string]rawPort, len(spec.Ports))
	for name, p := range spec.Ports {
		rp := rawPort{
			Description: p.Description,
		}

		for _, pc := range p.Connections {
			// Find connection name by matching pointer or by iterating
			connName, err := findConnectionName(spec.Connections, pc.Connection)
			if err != nil {
				return nil, fmt.Errorf("port %q: %w", name, err)
			}
			rp.Connections = append(rp.Connections, rawPortConnection{
				Connection: "#/connections/" + connName,
				Location:   pc.Location,
			})
		}

		if p.SchemaRef != "" {
			rp.Schema = &jsonschema.Schema{Ref: p.SchemaRef}
		} else if p.Schema != nil {
			rp.Schema = p.Schema
		}

		ports[name] = rp
	}

	raw := &rawSpec{
		OpenDPI:     spec.OpenDPI,
		Info:        rawInfo(spec.Info),
		Connections: connections,
		Ports:       ports,
	}
	if len(tags) > 0 {
		raw.Tags = tags
	}
	return raw, nil
}

func variablesEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !reflect.DeepEqual(v, bv) {
			return false
		}
	}
	return true
}

func findConnectionName(connections map[string]Connection, target *Connection) (string, error) {
	for name, c := range connections {
		if c.Type == target.Type &&
			c.Host == target.Host &&
			c.Description == target.Description &&
			variablesEqual(c.Variables, target.Variables) {
			return name, nil
		}
	}
	return "", fmt.Errorf("connection not found in spec (type=%q, host=%q, description=%q)",
		target.Type, target.Host, target.Description)
}

// WriteSchemaFile writes an empty object schema to the given path.
func WriteSchemaFile(dir, portName string, writeFunc func(string, any) error) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create schemas directory: %w", err)
	}
	schema := &jsonschema.Schema{
		Type:       "object",
		Properties: make(map[string]*jsonschema.Schema),
	}
	return writeFunc(filepath.Join(dir, portName+".schema.yaml"), schema)
}

// WriteEmptySchemaYAML writes an empty object schema as YAML.
func WriteEmptySchemaYAML(dir, portName string) error {
	return WriteSchemaFile(dir, portName, writeYaml)
}

// WriteEmptySchemaJSON writes an empty object schema as JSON.
func WriteEmptySchemaJSON(dir, portName string) error {
	return WriteSchemaFile(dir, portName, writeJSON)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeYaml(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(v)
}
