package opendpi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite_NilSpec(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{Path: tmpDir}

			err := tt.writer.Write(nil, cfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "nil spec")
		})
	}
}

func TestWrite_NilConfig(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test", Version: "1.0.0"},
			}

			err := tt.writer.Write(spec, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "nil config")
		})
	}
}

func TestWrite_DirectoryDoesNotExist(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Path: "/nonexistent/path/that/does/not/exist",
				Schema: config.SchemaConfig{
					Organization: config.SchemaInline,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test", Version: "1.0.0"},
			}

			err := tt.writer.Write(spec, cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "spec directory does not exist")
		})
	}
}

func TestWrite_Inline(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaInline,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0", Description: "A test product"},
				Tags: []Tag{
					{Name: "customer-data", Description: "Customer related data"},
				},
				Connections: map[string]Connection{
					"db": {
						Protocol:    "postgresql",
						Host:        "localhost:5432",
						Description: "Main database",
						Variables:   map[string]any{"schema": "public"},
					},
				},
				Ports: map[string]Port{
					"users": {
						Description: "User data",
						Connections: []PortConnection{
							{Connection: &Connection{
								Protocol:    "postgresql",
								Host:        "localhost:5432",
								Description: "Main database",
								Variables:   map[string]any{"schema": "public"},
							}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"id":   {Type: "integer"},
								"name": {Type: "string"},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify the spec file was created
			specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
			data, err := os.ReadFile(specFile) //nolint:gosec // test file
			require.NoError(t, err)

			content := string(data)
			// Schema should be inline in the port
			assert.Contains(t, content, "object")
			assert.Contains(t, content, "integer")
			assert.Contains(t, content, "string")
			// Metadata should be present
			assert.Contains(t, content, "customer-data")
			assert.Contains(t, content, "Main database")
			assert.Contains(t, content, "User data")
			// Verify schema structure is present (format-agnostic)
			assert.True(t, strings.Contains(content, `"id"`) || strings.Contains(content, "id:"))
			assert.True(t, strings.Contains(content, `"name"`) || strings.Contains(content, "name:"))

			// Verify no schemas directory was created
			_, err = os.Stat(filepath.Join(tmpDir, "schemas"))
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestWrite_Modular(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaModular,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"id":      {Type: "integer"},
								"address": {Ref: "#/$defs/Address"},
							},
							Defs: map[string]*jsonschema.Schema{
								"Address": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"street": {Type: "string"},
									},
								},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify schemas directory was created
			schemasDir := filepath.Join(tmpDir, "schemas")
			_, err = os.Stat(schemasDir)
			require.NoError(t, err)

			// Verify users.yaml was created with rewritten refs
			usersFile := filepath.Join(schemasDir, "users.yaml")
			data, err := os.ReadFile(usersFile) //nolint:gosec // test file
			require.NoError(t, err)
			usersContent := string(data)
			assert.Contains(t, usersContent, "Address.yaml")
			assert.NotContains(t, usersContent, "#/$defs/")

			// Verify Address.yaml was created
			addressFile := filepath.Join(schemasDir, "Address.yaml")
			_, err = os.Stat(addressFile)
			require.NoError(t, err)

			// Verify the opendpi.yaml references the schema file
			specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
			data, err = os.ReadFile(specFile) //nolint:gosec // test file
			require.NoError(t, err)
			content := string(data)
			assert.Contains(t, content, "schemas/users.yaml")
		})
	}
}

func TestWrite_Modular_NestedRefs(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaModular,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"orders": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.orders"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"customer": {Ref: "#/$defs/Customer"},
							},
							Defs: map[string]*jsonschema.Schema{
								"Customer": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"address": {Ref: "#/$defs/Address"},
									},
								},
								"Address": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"street": {Type: "string"},
									},
								},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify Customer.yaml has rewritten nested refs
			customerFile := filepath.Join(tmpDir, "schemas", "Customer.yaml")
			data, err := os.ReadFile(customerFile) //nolint:gosec // test file
			require.NoError(t, err)
			customerContent := string(data)
			assert.Contains(t, customerContent, "Address.yaml")
			assert.NotContains(t, customerContent, "#/$defs/")
		})
	}
}

func TestWrite_Modular_ComponentsRefRewrite(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaModular,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"profile": {Ref: "#/components/schemas/Profile"},
							},
							Defs: map[string]*jsonschema.Schema{
								"Profile": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"bio": {Type: "string"},
									},
								},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify users.yaml has #/components/schemas/ refs rewritten to .yaml
			usersFile := filepath.Join(tmpDir, "schemas", "users.yaml")
			data, err := os.ReadFile(usersFile) //nolint:gosec // test file
			require.NoError(t, err)
			usersContent := string(data)
			assert.Contains(t, usersContent, "Profile.yaml")
			assert.NotContains(t, usersContent, "#/components/schemas/")
		})
	}
}

func TestWrite_ModularCleansStaleFiles(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaModular,
				},
			}

			// Create schemas directory with a stale file
			schemasDir := filepath.Join(tmpDir, "schemas")
			require.NoError(t, os.MkdirAll(schemasDir, 0o750))
			staleFile := filepath.Join(schemasDir, "OldSchema.yaml")
			require.NoError(t, os.WriteFile(staleFile, []byte("type: object\n"), 0o600))

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{Type: "object"},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify stale file was removed
			_, err = os.Stat(staleFile)
			assert.True(t, os.IsNotExist(err), "stale file should be removed")

			// Verify users.yaml still exists
			usersSchemaFile := filepath.Join(schemasDir, "users.yaml")
			_, err = os.Stat(usersSchemaFile)
			require.NoError(t, err)
		})
	}
}

func TestWrite_Components(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaComponents,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"id":      {Type: "integer"},
								"address": {Ref: "#/$defs/Address"},
							},
							Defs: map[string]*jsonschema.Schema{
								"Address": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"street": {Type: "string"},
									},
								},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Read the spec file
			specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
			data, err := os.ReadFile(specFile) //nolint:gosec // test file
			require.NoError(t, err)
			content := string(data)

			// Should have components section (format-agnostic)
			assert.True(t, strings.Contains(content, `"components"`) || strings.Contains(content, "components:"))
			assert.True(t, strings.Contains(content, `"schemas"`) || strings.Contains(content, "schemas:"))
			// Port should reference component schema
			assert.Contains(t, content, "#/components/schemas/users")
			// $defs refs should be rewritten to components
			assert.Contains(t, content, "#/components/schemas/Address")
			// Old $defs refs should NOT be present
			assert.NotContains(t, content, "#/$defs/")

			// Verify no schemas directory was created
			_, err = os.Stat(filepath.Join(tmpDir, "schemas"))
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestWrite_Components_NestedRefs(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaComponents,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"orders": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.orders"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Properties: map[string]*jsonschema.Schema{
								"customer": {Ref: "#/$defs/Customer"},
							},
							Defs: map[string]*jsonschema.Schema{
								"Customer": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"address": {Ref: "#/$defs/Address"},
									},
								},
								"Address": {
									Type: "object",
									Properties: map[string]*jsonschema.Schema{
										"street": {Type: "string"},
									},
								},
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Read the spec file
			specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
			data, err := os.ReadFile(specFile) //nolint:gosec // test file
			require.NoError(t, err)
			content := string(data)

			// All nested refs should be rewritten to components format
			assert.Contains(t, content, "#/components/schemas/Customer")
			assert.Contains(t, content, "#/components/schemas/Address")
			// No $defs refs should remain
			assert.NotContains(t, content, "#/$defs/")
		})
	}
}

func TestWrite_DuplicateSchemaNames(t *testing.T) {
	tests := []struct {
		name         string
		writer       Writer
		organization config.SchemaOrganization
	}{
		{"YAML/Modular", YAMLWriter, config.SchemaModular},
		{"JSON/Modular", JSONWriter, config.SchemaModular},
		{"YAML/Components", YAMLWriter, config.SchemaComponents},
		{"JSON/Components", JSONWriter, config.SchemaComponents},
		{"YAML/Inline", YAMLWriter, config.SchemaInline},
		{"JSON/Inline", JSONWriter, config.SchemaInline},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: tt.organization,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Defs: map[string]*jsonschema.Schema{
								"Address": {Type: "object"},
							},
						},
					},
					"profiles": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.profiles"},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
							Defs: map[string]*jsonschema.Schema{
								"Address": {Type: "object"}, // Duplicate!
							},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate schema name")
			assert.Contains(t, err.Error(), "Address")
		})
	}
}

func TestWrite_UnsupportedOrganization(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: "unsupported-mode",
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "schema organization not supported")
		})
	}
}

func TestWrite_PortsWithoutSchemas(t *testing.T) {
	tests := []struct {
		name   string
		writer Writer
	}{
		{"YAML", YAMLWriter},
		{"JSON", JSONWriter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: config.SchemaInline,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Description: "Has schema",
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{Type: "object"},
					},
					"logs": {
						Description: "No schema",
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.logs"},
						},
						Schema: nil, // No schema
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			// Verify both ports are written
			specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
			data, err := os.ReadFile(specFile) //nolint:gosec // test file
			require.NoError(t, err)

			content := string(data)
			// Both ports should be present (format-agnostic)
			assert.True(t, strings.Contains(content, `"users"`) || strings.Contains(content, "users:"))
			assert.True(t, strings.Contains(content, `"logs"`) || strings.Contains(content, "logs:"))
		})
	}
}

func TestWrite_MultiplePorts(t *testing.T) {
	tests := []struct {
		name         string
		writer       Writer
		organization config.SchemaOrganization
	}{
		{"YAML/Modular", YAMLWriter, config.SchemaModular},
		{"JSON/Modular", JSONWriter, config.SchemaModular},
		{"YAML/Components", YAMLWriter, config.SchemaComponents},
		{"JSON/Components", JSONWriter, config.SchemaComponents},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &config.Config{
				Path: tmpDir,
				Schema: config.SchemaConfig{
					Organization: tt.organization,
				},
			}

			spec := &Spec{
				OpenDPI: "1.0.0",
				Info:    Info{Title: "Test Product", Version: "1.0.0"},
				Connections: map[string]Connection{
					"db": {Protocol: "postgresql", Host: "localhost:5432"},
				},
				Ports: map[string]Port{
					"users": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.users"},
						},
						Schema: &jsonschema.Schema{Type: "object"},
					},
					"orders": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.orders"},
						},
						Schema: &jsonschema.Schema{Type: "object"},
					},
					"products": {
						Connections: []PortConnection{
							{Connection: &Connection{Protocol: "postgresql", Host: "localhost:5432"}, Location: "public.products"},
						},
						Schema: &jsonschema.Schema{Type: "object"},
					},
				},
			}

			err := tt.writer.Write(spec, cfg)
			require.NoError(t, err)

			if tt.organization == config.SchemaModular {
				// Verify all schema files were created
				schemasDir := filepath.Join(tmpDir, "schemas")
				for _, portName := range []string{"users", "orders", "products"} {
					schemaFile := filepath.Join(schemasDir, portName+".yaml")
					_, err := os.Stat(schemaFile)
					require.NoError(t, err, "schema file %s should exist", portName)
				}
			} else {
				// Verify all schemas are in components
				specFile := filepath.Join(tmpDir, "opendpi"+tt.writer.extension)
				data, err := os.ReadFile(specFile) //nolint:gosec // test file
				require.NoError(t, err)

				content := string(data)
				assert.Contains(t, content, "#/components/schemas/users")
				assert.Contains(t, content, "#/components/schemas/orders")
				assert.Contains(t, content, "#/components/schemas/products")
			}
		})
	}
}
