// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionsCreate_Happy(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	out, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{
		Prj:         prj,
		Name:        "db",
		Path:        "connections/db.yaml",
		Type:        "postgresql",
		Host:        "localhost:5432",
		Description: "primary",
		Variables:   map[string]any{"db": "analytics"},
	})
	require.NoError(t, err)
	assert.True(t, out.Scaffolded)
	assert.FileExists(t, filepath.Join(prjDir, "connections/db.yaml"))
	assert.Equal(t, "connections/db.yaml", prj.Connections["db"])

	body, err := readFileString(filepath.Join(prjDir, "connections/db.yaml"))
	require.NoError(t, err)
	assert.Contains(t, body, "type: postgresql")
	assert.Contains(t, body, "host: localhost:5432")
	assert.Contains(t, body, "description: primary")
	assert.Contains(t, body, "db: analytics")
}

func TestConnectionsCreate_PreservesExistingFile(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "connections/db.yaml"), "hand-written\n")
	out, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{
		Prj: prj, Name: "db", Path: "connections/db.yaml", Type: "x", Host: "y",
	})
	require.NoError(t, err)
	assert.False(t, out.Scaffolded)
	body, err := readFileString(filepath.Join(prjDir, "connections/db.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "hand-written\n", body)
}

func TestConnectionsCreate_Duplicate(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	_, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{
		Prj: prj, Name: "db", Path: "connections/dup.yaml", Type: "x", Host: "y",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConnectionsCreate_Validation(t *testing.T) {
	_, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"host: required", "name: required", "path: required", "prj: project not loaded", "type: required"},
		errs,
	)
}

func TestConnectionsList_SortedAndEmpty(t *testing.T) {
	_, prj, _ := newTestProject(t)
	out, err := ConnectionsList(context.Background(), ConnectionsListInput{Prj: prj})
	require.NoError(t, err)
	assert.Empty(t, out.Connections)

	mustCreateConnection(t, prj, "zebra", "connections/zebra.yaml", "kafka", "k:9092")
	mustCreateConnection(t, prj, "apple", "connections/apple.yaml", "postgresql", "p:5432")
	out, err = ConnectionsList(context.Background(), ConnectionsListInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Connections, 2)
	assert.Equal(t, "apple", out.Connections[0].Name)
	assert.Equal(t, "zebra", out.Connections[1].Name)
}

func TestConnectionsList_Validation(t *testing.T) {
	_, err := ConnectionsList(context.Background(), ConnectionsListInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prj: project not loaded")
}

func TestConnectionsDescribe_HappyPath(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{
		Prj: prj, Name: "db", Path: "connections/db.yaml",
		Type: "postgresql", Host: "localhost:5432", Description: "primary",
		Variables: map[string]any{"db": "x"},
	})
	require.NoError(t, err)
	out, err := ConnectionsDescribe(context.Background(), ConnectionsDescribeInput{Prj: prj, Name: "db"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	require.NotNil(t, out.Conn)
	assert.Equal(t, "postgresql", out.Conn.Type)
	assert.Equal(t, "localhost:5432", out.Conn.Host)
	assert.Equal(t, "primary", out.Conn.Description)
	assert.Equal(t, map[string]any{"db": "x"}, out.Conn.Variables)
}

func TestConnectionsDescribe_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ConnectionsDescribe(context.Background(), ConnectionsDescribeInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConnectionsDescribe_ParseTolerantOnBadFile(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "connections/db.yaml"), "hand-written\n")
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	// File contains non-connection yaml — describe should not error.
	out, err := ConnectionsDescribe(context.Background(), ConnectionsDescribeInput{Prj: prj, Name: "db"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	assert.Nil(t, out.Conn)
}

func TestConnectionsDescribe_RefFile(t *testing.T) {
	// Connection file is purely a $ref. Current behavior: out.Conn.Ref
	// populated, Type/Host empty (no auto-resolve).
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "connections/db.yaml"), "$ref: ../other.yaml\n")
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "x", "y")
	out, err := ConnectionsDescribe(context.Background(), ConnectionsDescribeInput{Prj: prj, Name: "db"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	require.NotNil(t, out.Conn)
	assert.Equal(t, "../other.yaml", out.Conn.Ref)
	assert.Empty(t, out.Conn.Type)
	assert.Empty(t, out.Conn.Host)
}

func TestConnectionsDelete_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	out, err := ConnectionsDelete(context.Background(), ConnectionsDeleteInput{Prj: prj, Name: "db"})
	require.NoError(t, err)
	assert.Equal(t, "db", out.Name)
	assert.FileExists(t, filepath.Join(prjDir, "connections/db.yaml"))
	assert.NotContains(t, prj.Connections, "db")
}

func TestConnectionsDelete_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ConnectionsDelete(context.Background(), ConnectionsDeleteInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}