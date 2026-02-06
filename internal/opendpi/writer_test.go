package opendpi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dacolabs/jsonschema-go/jsonschema"
)

func TestVariablesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    map[string]any
		b    map[string]any
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "both empty",
			a:    map[string]any{},
			b:    map[string]any{},
			want: true,
		},
		{
			name: "nil and empty are equal",
			a:    nil,
			b:    map[string]any{},
			want: true,
		},
		{
			name: "same values",
			a:    map[string]any{"key": "value", "num": 42},
			b:    map[string]any{"key": "value", "num": 42},
			want: true,
		},
		{
			name: "different values",
			a:    map[string]any{"key": "value1"},
			b:    map[string]any{"key": "value2"},
			want: false,
		},
		{
			name: "different keys",
			a:    map[string]any{"key1": "value"},
			b:    map[string]any{"key2": "value"},
			want: false,
		},
		{
			name: "different lengths",
			a:    map[string]any{"key": "value"},
			b:    map[string]any{"key": "value", "extra": "data"},
			want: false,
		},
		{
			name: "nested maps equal",
			a:    map[string]any{"nested": map[string]any{"inner": "value"}},
			b:    map[string]any{"nested": map[string]any{"inner": "value"}},
			want: true,
		},
		{
			name: "nested maps different",
			a:    map[string]any{"nested": map[string]any{"inner": "value1"}},
			b:    map[string]any{"nested": map[string]any{"inner": "value2"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := variablesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("variablesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindConnectionName(t *testing.T) {
	connections := map[string]Connection{
		"conn1": {
			Type:        "kafka",
			Host:        "localhost:9092",
			Description: "Local Kafka",
			Variables:   nil,
		},
		"conn2": {
			Type:        "postgresql",
			Host:        "db.example.com",
			Description: "Production DB",
			Variables:   map[string]any{"ssl": true},
		},
		"conn3": {
			Type:        "s3",
			Host:        "s3.amazonaws.com",
			Description: "Data Lake",
			Variables:   map[string]any{"bucket": "my-bucket", "region": "us-east-1"},
		},
	}

	tests := []struct {
		name       string
		target     *Connection
		wantName   string
		wantErrMsg string
	}{
		{
			name: "exact match without variables",
			target: &Connection{
				Type:        "kafka",
				Host:        "localhost:9092",
				Description: "Local Kafka",
				Variables:   nil,
			},
			wantName:   "conn1",
			wantErrMsg: "",
		},
		{
			name: "exact match with variables",
			target: &Connection{
				Type:        "postgresql",
				Host:        "db.example.com",
				Description: "Production DB",
				Variables:   map[string]any{"ssl": true},
			},
			wantName:   "conn2",
			wantErrMsg: "",
		},
		{
			name: "exact match with complex variables",
			target: &Connection{
				Type:        "s3",
				Host:        "s3.amazonaws.com",
				Description: "Data Lake",
				Variables:   map[string]any{"bucket": "my-bucket", "region": "us-east-1"},
			},
			wantName:   "conn3",
			wantErrMsg: "",
		},
		{
			name: "no match - different type",
			target: &Connection{
				Type:        "http",
				Host:        "localhost:9092",
				Description: "Local Kafka",
				Variables:   nil,
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
		{
			name: "no match - different host",
			target: &Connection{
				Type:        "kafka",
				Host:        "remote:9092",
				Description: "Local Kafka",
				Variables:   nil,
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
		{
			name: "no match - different description",
			target: &Connection{
				Type:        "kafka",
				Host:        "localhost:9092",
				Description: "Remote Kafka",
				Variables:   nil,
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
		{
			name: "no match - different variables",
			target: &Connection{
				Type:        "postgresql",
				Host:        "db.example.com",
				Description: "Production DB",
				Variables:   map[string]any{"ssl": false},
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
		{
			name: "no match - missing variables",
			target: &Connection{
				Type:        "postgresql",
				Host:        "db.example.com",
				Description: "Production DB",
				Variables:   nil,
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
		{
			name: "no match - extra variables",
			target: &Connection{
				Type:        "kafka",
				Host:        "localhost:9092",
				Description: "Local Kafka",
				Variables:   map[string]any{"extra": "value"},
			},
			wantName:   "",
			wantErrMsg: "connection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := findConnectionName(connections, tt.target)
			if tt.wantErrMsg != "" {
				if err == nil {
					t.Errorf("findConnectionName() error = nil, want error containing %q", tt.wantErrMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("findConnectionName() error = %q, want error containing %q", err.Error(), tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("findConnectionName() unexpected error = %v", err)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("findConnectionName() = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}

func TestToRaw(t *testing.T) {
	tests := []struct {
		name        string
		spec        *Spec
		wantErr     bool
		errContains string
	}{
		{
			name: "valid spec with matching connection",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{
					"myconn": {
						Type: "kafka",
						Host: "localhost:9092",
					},
				},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type: "kafka",
									Host: "localhost:9092",
								},
								Location: "test.topic",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "port references non-existent connection",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{
					"myconn": {
						Type: "kafka",
						Host: "localhost:9092",
					},
				},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type: "postgresql",
									Host: "db.example.com",
								},
								Location: "test.table",
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "port \"output\": connection not found",
		},
		{
			name: "empty connections map",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type: "kafka",
									Host: "localhost:9092",
								},
								Location: "test.topic",
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "port \"output\": connection not found",
		},
		{
			name: "connection with different variables",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{
					"myconn": {
						Type:      "kafka",
						Host:      "localhost:9092",
						Variables: map[string]any{"ssl": true},
					},
				},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type:      "kafka",
									Host:      "localhost:9092",
									Variables: map[string]any{"ssl": false},
								},
								Location: "test.topic",
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "port \"output\": connection not found",
		},
		{
			name: "valid spec with variables",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{
					"myconn": {
						Type:      "kafka",
						Host:      "localhost:9092",
						Variables: map[string]any{"ssl": true, "port": 9092},
					},
				},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type:      "kafka",
									Host:      "localhost:9092",
									Variables: map[string]any{"ssl": true, "port": 9092},
								},
								Location: "test.topic",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := toRaw(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Errorf("toRaw() error = nil, want error")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("toRaw() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("toRaw() unexpected error = %v", err)
				return
			}
			if raw == nil {
				t.Errorf("toRaw() returned nil rawSpec")
			}
		})
	}
}

func TestWriterWrite(t *testing.T) {
	tests := []struct {
		name        string
		spec        *Spec
		wantErr     bool
		errContains string
	}{
		{
			name: "valid spec",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{
					"myconn": {
						Type: "kafka",
						Host: "localhost:9092",
					},
				},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type: "kafka",
									Host: "localhost:9092",
								},
								Location: "test.topic",
							},
						},
						Schema: &jsonschema.Schema{
							Type: "object",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid spec - missing connection",
			spec: &Spec{
				OpenDPI: "0.1.0",
				Info: Info{
					Title:   "Test Product",
					Version: "1.0.0",
				},
				Connections: map[string]Connection{},
				Ports: map[string]Port{
					"output": {
						Description: "Test output",
						Connections: []PortConnection{
							{
								Connection: &Connection{
									Type: "kafka",
									Host: "localhost:9092",
								},
								Location: "test.topic",
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "connection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a temporary directory for testing
			tmpDir := t.TempDir()

			err := YAMLWriter.Write(tt.spec, tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Writer.Write() error = nil, want error")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Writer.Write() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("Writer.Write() unexpected error = %v", err)
			}
		})
	}
}

func TestWritePreservesExternalSchemaRefs(t *testing.T) {
	tmpDir := t.TempDir()

	// Copy testdata spec and schema files
	specData, err := os.ReadFile("testdata/with-external-schemas.yaml")
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "opendpi.yaml"), specData, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	schemasDir := filepath.Join(tmpDir, "schemas")
	err = os.MkdirAll(schemasDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"user.yaml", "user-profile.yaml", "address.yaml"} {
		data, readErr := os.ReadFile(filepath.Join("testdata/schemas", name)) //nolint:gosec // test data
		if readErr != nil {
			t.Fatal(readErr)
		}
		writeErr := os.WriteFile(filepath.Join(schemasDir, name), data, 0o600)
		if writeErr != nil {
			t.Fatal(writeErr)
		}
	}

	// Parse the spec (resolves $ref to inline)
	f, err := os.Open(filepath.Join(tmpDir, "opendpi.yaml")) //nolint:gosec // test data
	if err != nil {
		t.Fatal(err)
	}
	spec, err := YAML.Parse(f, os.DirFS(tmpDir))
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Add a new port (simulates ports add)
	newName := "events"
	schemaRef := "schemas/" + newName + ".schema.yaml"
	spec.Ports[newName] = Port{
		Description: "Event data",
		SchemaRef:   schemaRef,
		Schema: &jsonschema.Schema{
			Ref: schemaRef,
		},
	}

	// Write back
	err = YAMLWriter.Write(spec, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "opendpi.yaml")) //nolint:gosec // test data
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)

	// All three ports should have $ref (2 existing + 1 new)
	refCount := strings.Count(output, "$ref")
	if refCount != 3 {
		t.Errorf("expected 3 $ref entries (2 existing + 1 new), got %d\nOutput:\n%s", refCount, output)
	}

	for _, path := range []string{"schemas/user.yaml", "schemas/user-profile.yaml", "schemas/events.schema.yaml"} {
		if !strings.Contains(output, path) {
			t.Errorf("expected %q in output YAML", path)
		}
	}
}
