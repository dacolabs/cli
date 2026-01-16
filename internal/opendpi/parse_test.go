package opendpi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Minimal(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "minimal.yaml", YAML},
		{"JSON", "minimal.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f)
			require.NoError(t, err)

			// Check basic fields
			assert.Equal(t, "1.0.0", spec.OpenDPI)
			assert.Equal(t, "My Data Product", spec.Info.Title)
			assert.Equal(t, "1.0.0", spec.Info.Version)

			// Check connection
			require.Len(t, spec.Connections, 1)
			conn, ok := spec.Connections["database"]
			require.True(t, ok, "connection 'database' not found")
			assert.Equal(t, "postgresql", conn.Protocol)
			assert.Equal(t, "localhost:5432", conn.Host)

			// Check port
			require.Len(t, spec.Ports, 1)
			port, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")
			require.Len(t, port.Connections, 1)

			// Check connection ref was resolved
			pc := port.Connections[0]
			require.NotNil(t, pc.Connection, "port connection was not resolved")
			assert.Equal(t, "postgresql", pc.Connection.Protocol)
			assert.Equal(t, "users", pc.Location)

			// Check inline schema
			require.NotNil(t, port.Schema)
			assert.Equal(t, "object", port.Schema.Type)
		})
	}
}

func TestParse_MultiPort(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "multi-port.yaml", YAML},
		{"JSON", "multi-port.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f)
			require.NoError(t, err)

			// Check info
			assert.Equal(t, "E-Commerce Analytics", spec.Info.Title)

			// Check tags
			assert.Len(t, spec.Tags, 3)

			// Check connections
			require.Len(t, spec.Connections, 2)
			warehouse, ok := spec.Connections["warehouse"]
			require.True(t, ok, "connection 'warehouse' not found")
			assert.Equal(t, "postgresql", warehouse.Protocol)
			assert.Len(t, warehouse.Variables, 2)

			// Check ports
			require.Len(t, spec.Ports, 3)

			// Check daily_orders port resolves to warehouse connection
			dailyOrders, ok := spec.Ports["daily_orders"]
			require.True(t, ok, "port 'daily_orders' not found")
			assert.Equal(t, "postgresql", dailyOrders.Connections[0].Connection.Protocol)

			// Check order_events port resolves to event_stream connection
			orderEvents, ok := spec.Ports["order_events"]
			require.True(t, ok, "port 'order_events' not found")
			assert.Equal(t, "kafka", orderEvents.Connections[0].Connection.Protocol)
		})
	}
}

func TestParse_WithComponents(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "with-components.yaml", YAML},
		{"JSON", "with-components.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f)
			require.NoError(t, err)

			// Check ports with schema refs were resolved
			usersPort, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")

			// Schema should be resolved (not a ref anymore)
			require.NotNil(t, usersPort.Schema)
			assert.Empty(t, usersPort.Schema.Ref, "schema ref should be empty after resolution")
			assert.Equal(t, "object", usersPort.Schema.Type)

			// Check that Defs was populated with component schemas
			require.NotNil(t, usersPort.Schema.Defs)
			assert.NotEmpty(t, usersPort.Schema.Defs)

			// Check specific component schemas are in Defs
			expectedSchemas := []string{"User", "UserProfile", "Address", "Timestamp", "BaseEvent", "ActivityEvent", "AuthEvent"}
			for _, name := range expectedSchemas {
				assert.Contains(t, usersPort.Schema.Defs, name, "schema %q not found in Defs", name)
			}

			// Check activity_log port has schema with nested ref (allOf with $ref)
			activityPort, ok := spec.Ports["activity_log"]
			require.True(t, ok, "port 'activity_log' not found")
			require.NotNil(t, activityPort.Schema)
			// ActivityEvent uses allOf with $ref to BaseEvent - this should be in Defs
			assert.Contains(t, activityPort.Schema.Defs, "BaseEvent")
		})
	}
}

func TestParse_NilReader(t *testing.T) {
	_, err := JSON.Parse(nil)
	assert.Error(t, err)

	_, err = YAML.Parse(nil)
	assert.Error(t, err)
}
