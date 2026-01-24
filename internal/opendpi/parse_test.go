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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check ports with schema refs were resolved
			usersPort, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")

			// Schema should be resolved (not a ref anymore)
			require.NotNil(t, usersPort.Schema)
			assert.Empty(t, usersPort.Schema.Ref, "schema ref should be empty after resolution")
			assert.Equal(t, "object", usersPort.Schema.Type)

			// User schema has no component refs, so Defs should be empty
			assert.Empty(t, usersPort.Schema.Defs, "User schema has no refs, Defs should be empty")

			// Check user_profiles port - UserProfile references Address
			profilesPort, ok := spec.Ports["user_profiles"]
			require.True(t, ok, "port 'user_profiles' not found")
			require.NotNil(t, profilesPort.Schema)
			require.NotNil(t, profilesPort.Schema.Defs)
			assert.Len(t, profilesPort.Schema.Defs, 1, "UserProfile only references Address")
			assert.Contains(t, profilesPort.Schema.Defs, "Address")

			// Check activity_log port - ActivityEvent references BaseEvent via allOf
			activityPort, ok := spec.Ports["activity_log"]
			require.True(t, ok, "port 'activity_log' not found")
			require.NotNil(t, activityPort.Schema)
			require.NotNil(t, activityPort.Schema.Defs)
			assert.Len(t, activityPort.Schema.Defs, 1, "ActivityEvent only references BaseEvent")
			assert.Contains(t, activityPort.Schema.Defs, "BaseEvent")
		})
	}
}

func TestParse_ExternalSchemas(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "with-external-schemas.yaml", YAML},
		{"JSON", "with-external-schemas.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check users port - schema loaded from external file
			usersPort, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")
			require.NotNil(t, usersPort.Schema)
			assert.Equal(t, "object", usersPort.Schema.Type)
			assert.Contains(t, usersPort.Schema.Required, "id")
			assert.Contains(t, usersPort.Schema.Required, "email")

			// Check user_profiles port - schema loaded from external file with nested refs
			profilesPort, ok := spec.Ports["user_profiles"]
			require.True(t, ok, "port 'user_profiles' not found")
			require.NotNil(t, profilesPort.Schema)
			assert.Equal(t, "object", profilesPort.Schema.Type)

			// Check that nested $ref to address.yaml was resolved
			billingAddr, ok := profilesPort.Schema.Properties["billing_address"]
			require.True(t, ok, "billing_address property not found")
			assert.Equal(t, "object", billingAddr.Type, "billing_address should be resolved to object type")
			assert.Contains(t, billingAddr.Properties, "street", "address schema should have street property")
			assert.Contains(t, billingAddr.Properties, "city", "address schema should have city property")
		})
	}
}

func TestParse_MixedInlineComponent(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "mixed-inline-component.yaml", YAML},
		{"JSON", "mixed-inline-component.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check orders port - inline schema with component refs
			ordersPort, ok := spec.Ports["orders"]
			require.True(t, ok, "port 'orders' not found")
			require.NotNil(t, ordersPort.Schema)

			// Schema should be inline (not resolved from ref)
			assert.Equal(t, "object", ordersPort.Schema.Type)
			assert.Contains(t, ordersPort.Schema.Required, "order_id")

			// Check that component refs inside inline schema are in Defs
			require.NotNil(t, ordersPort.Schema.Defs, "Defs should contain referenced components")

			// Should contain Customer, OrderItem, Address, and Money (transitive from OrderItem)
			assert.Contains(t, ordersPort.Schema.Defs, "Customer", "Customer should be in Defs")
			assert.Contains(t, ordersPort.Schema.Defs, "OrderItem", "OrderItem should be in Defs")
			assert.Contains(t, ordersPort.Schema.Defs, "Address", "Address should be in Defs")
			assert.Contains(t, ordersPort.Schema.Defs, "Money", "Money should be in Defs (transitive via OrderItem)")

			// Verify the Defs count - should be exactly 4
			assert.Len(t, ordersPort.Schema.Defs, 4, "Should have exactly 4 component schemas in Defs")
		})
	}
}

func TestParse_ComponentExternalRef(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "component-external-ref.yaml", YAML},
		{"JSON", "component-external-ref.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check users port - component that refs external file
			usersPort, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")
			require.NotNil(t, usersPort.Schema)

			// Schema should be resolved from external file via component
			assert.Equal(t, "object", usersPort.Schema.Type)
			assert.Contains(t, usersPort.Schema.Required, "id")
			assert.Contains(t, usersPort.Schema.Required, "email")

			// Check profiles port - component refs external file with nested refs
			profilesPort, ok := spec.Ports["profiles"]
			require.True(t, ok, "port 'profiles' not found")
			require.NotNil(t, profilesPort.Schema)
			assert.Equal(t, "object", profilesPort.Schema.Type)

			// Check that nested ref to address.yaml was resolved
			billingAddr, ok := profilesPort.Schema.Properties["billing_address"]
			require.True(t, ok, "billing_address property not found")
			assert.Equal(t, "object", billingAddr.Type, "billing_address should be resolved")
			assert.Contains(t, billingAddr.Properties, "street")
		})
	}
}

func TestParse_DeepNestedRefs(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "deep-nested-refs.yaml", YAML},
		{"JSON", "deep-nested-refs.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check entities port - uses schema from nested directory
			entitiesPort, ok := spec.Ports["entities"]
			require.True(t, ok, "port 'entities' not found")
			require.NotNil(t, entitiesPort.Schema)
			assert.Equal(t, "object", entitiesPort.Schema.Type)

			// Check that ../address.yaml ref was resolved correctly
			addr, ok := entitiesPort.Schema.Properties["address"]
			require.True(t, ok, "address property not found")
			assert.Equal(t, "object", addr.Type, "address should be resolved from parent directory")
			assert.Contains(t, addr.Properties, "street")
			assert.Contains(t, addr.Properties, "city")

			// Check mixed port - inline schema with both external and component refs
			mixedPort, ok := spec.Ports["mixed"]
			require.True(t, ok, "port 'mixed' not found")
			require.NotNil(t, mixedPort.Schema)

			// Check external ref was resolved
			directAddr, ok := mixedPort.Schema.Properties["direct_address"]
			require.True(t, ok, "direct_address property not found")
			assert.Equal(t, "object", directAddr.Type)
			assert.Contains(t, directAddr.Properties, "street")

			// Check component ref is in Defs
			require.NotNil(t, mixedPort.Schema.Defs)
			assert.Contains(t, mixedPort.Schema.Defs, "Timestamp")

			// Check nested inline schema with external ref
			nestedData, ok := mixedPort.Schema.Properties["nested_data"]
			require.True(t, ok, "nested_data property not found")
			innerAddr, ok := nestedData.Properties["inner_address"]
			require.True(t, ok, "inner_address property not found")
			assert.Equal(t, "object", innerAddr.Type)
			assert.Contains(t, innerAddr.Properties, "city")
		})
	}
}

func TestParse_CircularRefs(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "circular-refs.yaml", YAML},
		{"JSON", "circular-refs.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			// Should not hang or panic on circular refs
			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check nodes port - Node refs itself
			nodesPort, ok := spec.Ports["nodes"]
			require.True(t, ok, "port 'nodes' not found")
			require.NotNil(t, nodesPort.Schema)

			// Node should be in Defs (self-referential)
			require.NotNil(t, nodesPort.Schema.Defs)
			assert.Contains(t, nodesPort.Schema.Defs, "Node")

			// Should only have Node, not duplicated
			assert.Len(t, nodesPort.Schema.Defs, 1, "Self-ref should only add schema once")
		})
	}
}

func TestParse_RefOnlySchema(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "ref-only-schema.yaml", YAML},
		{"JSON", "ref-only-schema.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check users port - uses UserWrapper which is just a ref to BaseUser
			usersPort, ok := spec.Ports["users"]
			require.True(t, ok, "port 'users' not found")
			require.NotNil(t, usersPort.Schema)

			// The schema IS UserWrapper, so UserWrapper doesn't need to be in Defs.
			// But UserWrapper refs BaseUser, so BaseUser should be in Defs.
			require.NotNil(t, usersPort.Schema.Defs)
			assert.Contains(t, usersPort.Schema.Defs, "BaseUser")
			assert.Len(t, usersPort.Schema.Defs, 1, "Only BaseUser should be in Defs")

			// Verify the schema is the UserWrapper (which is a ref-only schema)
			// After resolution, the Ref field should still contain the ref string
			assert.Equal(t, "#/components/schemas/BaseUser", usersPort.Schema.Ref)

			// Check direct port - uses BaseUser directly
			directPort, ok := spec.Ports["direct"]
			require.True(t, ok, "port 'direct' not found")
			require.NotNil(t, directPort.Schema)

			// BaseUser has no refs, so Defs should be empty
			assert.Empty(t, directPort.Schema.Defs, "BaseUser has no refs")
		})
	}
}

func TestParse_CrossFormatRefs(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML spec with mixed refs", "cross-format-yaml.yaml", YAML},
		{"JSON spec with mixed refs", "cross-format-json.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Check that YAML schema reference was resolved
			usersYAML, ok := spec.Ports["users_yaml"]
			require.True(t, ok, "port 'users_yaml' not found")
			require.NotNil(t, usersYAML.Schema)
			assert.Equal(t, "object", usersYAML.Schema.Type)
			assert.Contains(t, usersYAML.Schema.Required, "id")
			assert.Contains(t, usersYAML.Schema.Required, "email")

			// Check that JSON schema reference was resolved
			usersJSON, ok := spec.Ports["users_json"]
			require.True(t, ok, "port 'users_json' not found")
			require.NotNil(t, usersJSON.Schema)
			assert.Equal(t, "object", usersJSON.Schema.Type)
			assert.Contains(t, usersJSON.Schema.Required, "id")
			assert.Contains(t, usersJSON.Schema.Required, "email")

			// Check YAML profile with nested YAML refs
			profilesYAML, ok := spec.Ports["profiles_yaml"]
			require.True(t, ok, "port 'profiles_yaml' not found")
			require.NotNil(t, profilesYAML.Schema)
			billingYAML, ok := profilesYAML.Schema.Properties["billing_address"]
			require.True(t, ok, "billing_address not found in YAML profile")
			assert.Equal(t, "object", billingYAML.Type)
			assert.Contains(t, billingYAML.Properties, "street")

			// Check JSON profile with nested JSON refs
			profilesJSON, ok := spec.Ports["profiles_json"]
			require.True(t, ok, "port 'profiles_json' not found")
			require.NotNil(t, profilesJSON.Schema)
			billingJSON, ok := profilesJSON.Schema.Properties["billing_address"]
			require.True(t, ok, "billing_address not found in JSON profile")
			assert.Equal(t, "object", billingJSON.Type)
			assert.Contains(t, billingJSON.Properties, "street")

			// Check inline schema with mixed format refs
			mixed, ok := spec.Ports["mixed_inline"]
			require.True(t, ok, "port 'mixed_inline' not found")
			require.NotNil(t, mixed.Schema)

			yamlAddr, ok := mixed.Schema.Properties["yaml_address"]
			require.True(t, ok, "yaml_address not found")
			assert.Equal(t, "object", yamlAddr.Type)
			assert.Contains(t, yamlAddr.Properties, "city")

			jsonAddr, ok := mixed.Schema.Properties["json_address"]
			require.True(t, ok, "json_address not found")
			assert.Equal(t, "object", jsonAddr.Type)
			assert.Contains(t, jsonAddr.Properties, "city")
		})
	}
}

func TestParse_MissingExternalRef(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		parser       Parser
		errorContain string
	}{
		{"YAML", "missing-external-ref.yaml", YAML, "does-not-exist.yaml"},
		{"JSON", "missing-external-ref.json", JSON, "does-not-exist.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			_, err = tt.parser.Parse(f, os.DirFS("testdata"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContain)
		})
	}
}

func TestParse_MissingComponentRef(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		parser Parser
	}{
		{"YAML", "missing-component-ref.yaml", YAML},
		{"JSON", "missing-component-ref.json", JSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			defer f.Close() //nolint:errcheck

			_, err = tt.parser.Parse(f, os.DirFS("testdata"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "DoesNotExist")
		})
	}
}

func TestParse_NilReader(t *testing.T) {
	_, err := JSON.Parse(nil, os.DirFS("testdata"))
	assert.Error(t, err)

	_, err = YAML.Parse(nil, os.DirFS("testdata"))
	assert.Error(t, err)
}

func TestParse_NilFS(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "minimal.yaml"))
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck

	_, err = YAML.Parse(f, nil)
	assert.Error(t, err)
}

func TestParse_SchemasMap(t *testing.T) {
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Schemas map should be populated
			require.NotNil(t, spec.Schemas)
			require.Len(t, spec.Schemas, 1)

			// Schema should be keyed by port name for inline schemas
			schema, ok := spec.Schemas["users"]
			require.True(t, ok, "schema 'users' should exist in Schemas map")
			assert.Equal(t, "object", schema.Type)
		})
	}
}

func TestParse_SchemasMap_Components(t *testing.T) {
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Schemas map should contain component schemas
			require.NotNil(t, spec.Schemas)

			// Component schemas should be in the map
			_, ok := spec.Schemas["User"]
			assert.True(t, ok, "component schema 'User' should exist in Schemas map")

			_, ok = spec.Schemas["Address"]
			assert.True(t, ok, "component schema 'Address' should exist in Schemas map")
		})
	}
}

func TestParse_SchemasMap_MultiPort(t *testing.T) {
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Schemas map should contain all port schemas
			require.NotNil(t, spec.Schemas)
			require.Len(t, spec.Schemas, 3) // daily_orders, user_segments, order_events

			for portName := range spec.Ports {
				_, ok := spec.Schemas[portName]
				assert.True(t, ok, "schema for port %q should exist in Schemas map", portName)
			}
		})
	}
}

func TestParse_SchemasResolved(t *testing.T) {
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

			spec, err := tt.parser.Parse(f, os.DirFS("testdata"))
			require.NoError(t, err)

			// Schemas in the map should be fully resolved (self-contained)
			userSchema, ok := spec.Schemas["User"]
			require.True(t, ok)

			// Schema should have type resolved (not just a ref)
			assert.Equal(t, "object", userSchema.Type)
			assert.NotNil(t, userSchema.Properties)
		})
	}
}
