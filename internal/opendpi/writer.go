package opendpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dacolabs/cli/internal/config"
	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// Writer encodes an OpenDPI spec to a file.
type Writer struct {
	write func(path string, v any) error
}

var (
	// JSONWriter writes OpenDPI specs as JSON.
	JSONWriter = Writer{writeJSON}
	// YAMLWriter writes OpenDPI specs as YAML.
	YAMLWriter = Writer{writeYaml}
)

// Write encodes the spec to cfg.Path/opendpi.yaml.
func (wr Writer) Write(spec *Spec, cfg *config.Config) error {
	if spec == nil {
		return errors.New("nil spec")
	}
	if cfg == nil {
		return errors.New("nil config")
	}
	if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
		return fmt.Errorf("spec directory does not exist: %s", cfg.Path)
	}

	// Convert Spec to rawSpec (ignoring schemas and components for now)
	connections := make(map[string]rawConnection, len(spec.Connections))
	for name, c := range spec.Connections {
		connections[name] = rawConnection(c)
	}

	tags := make([]rawTag, len(spec.Tags))
	for i, t := range spec.Tags {
		tags[i] = rawTag(t)
	}

	// For modular/components modes, collect schema names and check for duplicates
	var schemaNames map[string]string
	if cfg.Schema.Organization == config.SchemaModular || cfg.Schema.Organization == config.SchemaComponents {
		schemaNames = make(map[string]string) // name -> port that defined it
		for name, p := range spec.Ports {
			if p.Schema == nil {
				continue
			}
			if existingPort, ok := schemaNames[name]; ok {
				return fmt.Errorf("duplicate schema name %q: defined by ports %q and %q", name, existingPort, name)
			}
			schemaNames[name] = name
			for defName := range p.Schema.Defs {
				if existingPort, ok := schemaNames[defName]; ok {
					return fmt.Errorf("duplicate schema name %q: defined in port %q and port %q", defName, existingPort, name)
				}
				schemaNames[defName] = name
			}
		}
	}

	ports := make(map[string]rawPort, len(spec.Ports))
	var components *rawComponents

	switch cfg.Schema.Organization {
	case config.SchemaModular:
		schemasDir := filepath.Join(cfg.Path, "schemas")
		if err := os.MkdirAll(schemasDir, 0o750); err != nil {
			return fmt.Errorf("failed to create schemas directory: %w", err)
		}

		// Remove stale schema files
		if entries, err := os.ReadDir(schemasDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
					name := strings.TrimSuffix(entry.Name(), ".yaml")
					if _, exists := schemaNames[name]; !exists {
						_ = os.Remove(filepath.Join(schemasDir, entry.Name()))
					}
				}
			}
		}

		for name, p := range spec.Ports {
			var schema *jsonschema.Schema
			if p.Schema != nil {
				// Write main schema (without $defs - they go to separate files)
				schemaToWrite := *p.Schema
				schemaToWrite.Defs = nil
				// Rewrite $refs to point to external files
				for s := range iterNestedRefs(&schemaToWrite, nil) {
					if strings.HasPrefix(s.Ref, "#/components/schemas/") {
						refName := strings.TrimPrefix(s.Ref, "#/components/schemas/")
						s.Ref = refName + ".yaml"
					} else if strings.HasPrefix(s.Ref, "#/$defs/") {
						refName := strings.TrimPrefix(s.Ref, "#/$defs/")
						s.Ref = refName + ".yaml"
					}
				}
				schemaFile := filepath.Join(schemasDir, name+".yaml")
				if err := writeYaml(schemaFile, &schemaToWrite); err != nil {
					return fmt.Errorf("failed to write schema file %s: %w", schemaFile, err)
				}

				// Write $defs as separate files
				for defName, defSchema := range p.Schema.Defs {
					defSchemaToWrite := *defSchema
					// Rewrite $refs in def schemas too
					for s := range iterNestedRefs(&defSchemaToWrite, nil) {
						if strings.HasPrefix(s.Ref, "#/components/schemas/") {
							refName := strings.TrimPrefix(s.Ref, "#/components/schemas/")
							s.Ref = refName + ".yaml"
						} else if strings.HasPrefix(s.Ref, "#/$defs/") {
							refName := strings.TrimPrefix(s.Ref, "#/$defs/")
							s.Ref = refName + ".yaml"
						}
					}
					defFile := filepath.Join(schemasDir, defName+".yaml")
					if err := writeYaml(defFile, &defSchemaToWrite); err != nil {
						return fmt.Errorf("failed to write schema file %s: %w", defFile, err)
					}
				}

				schema = &jsonschema.Schema{Ref: "schemas/" + name + ".yaml"}
			}

			ports[name] = createRawPort(spec, &p, schema)
		}

	case config.SchemaComponents:
		components = &rawComponents{Schemas: make(map[string]*jsonschema.Schema)}

		for name, p := range spec.Ports {
			var schema *jsonschema.Schema
			if p.Schema != nil {
				// Add main schema (without $defs)
				schemaToAdd := *p.Schema
				schemaToAdd.Defs = nil
				// Rewrite $refs from $defs to components
				for s := range iterNestedRefs(&schemaToAdd, nil) {
					if strings.HasPrefix(s.Ref, "#/$defs/") {
						refName := strings.TrimPrefix(s.Ref, "#/$defs/")
						s.Ref = "#/components/schemas/" + refName
					}
				}
				components.Schemas[name] = &schemaToAdd

				// Add $defs as separate component schemas
				for defName, defSchema := range p.Schema.Defs {
					defSchemaToAdd := *defSchema
					// Rewrite $refs in def schemas too
					for s := range iterNestedRefs(&defSchemaToAdd, nil) {
						if strings.HasPrefix(s.Ref, "#/$defs/") {
							refName := strings.TrimPrefix(s.Ref, "#/$defs/")
							s.Ref = "#/components/schemas/" + refName
						}
					}
					components.Schemas[defName] = &defSchemaToAdd
				}

				schema = &jsonschema.Schema{Ref: "#/components/schemas/" + name}
			}

			ports[name] = createRawPort(spec, &p, schema)
		}

	case config.SchemaInline:
		for name, p := range spec.Ports {
			ports[name] = createRawPort(spec, &p, p.Schema)
		}
	default:
		return fmt.Errorf("schema organization not supported")
	}

	raw := &rawSpec{
		OpenDPI:     spec.OpenDPI,
		Info:        rawInfo{Title: spec.Info.Title, Version: spec.Info.Version, Description: spec.Info.Description},
		Tags:        tags,
		Connections: connections,
		Ports:       ports,
		Components:  components,
	}

	specFile := filepath.Join(cfg.Path, "opendpi.yaml")
	return wr.write(specFile, raw)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeYaml(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(v)
}

func createRawPort(spec *Spec, p *Port, schema *jsonschema.Schema) rawPort {
	portConns := make([]rawPortConnection, len(p.Connections))
	for i, pc := range p.Connections {
		var connName string
		for cname, c := range spec.Connections {
			if pc.Connection.Protocol == c.Protocol && pc.Connection.Host == c.Host {
				connName = cname
				break
			}
		}
		portConns[i] = rawPortConnection{
			Connection: connectionRef{Ref: "#/connections/" + connName},
			Location:   pc.Location,
		}
	}

	return rawPort{
		Description: p.Description,
		Connections: portConns,
		Schema:      schema,
	}
}
