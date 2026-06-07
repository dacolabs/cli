// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
)

type ConnectionEntry struct {
	Name string
	Path string
}

type ConnectionsCreateInput struct {
	Prj         *settings.Project
	Name        string
	Path        string
	Type        string
	Host        string
	Description string
	Variables   map[string]any
}

type ConnectionsCreateOutput struct {
	Name       string
	Path       string
	Scaffolded bool
}

func (in ConnectionsCreateInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	if in.Path == "" {
		problems["path"] = "required"
	}
	if in.Type == "" {
		problems["type"] = "required"
	}
	if in.Host == "" {
		problems["host"] = "required"
	}
	return problems
}

func ConnectionsCreate(ctx context.Context, in ConnectionsCreateInput) (*ConnectionsCreateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, exists := in.Prj.Connections[in.Name]; exists {
		return nil, fmt.Errorf("connection %q already exists", in.Name)
	}

	abs := resolveProjectPath(in.Prj, in.Path)
	scaffolded := false
	if _, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) {
		conn := opendpi.Connection{
			Type:        in.Type,
			Host:        in.Host,
			Description: in.Description,
			Variables:   in.Variables,
		}
		if err := opendpi.ScaffoldConnection(abs, conn); err != nil {
			return nil, fmt.Errorf("scaffold %s: %w", abs, err)
		}
		scaffolded = true
	} else if err != nil {
		return nil, err
	}

	in.Prj.Connections[in.Name] = in.Path
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &ConnectionsCreateOutput{Name: in.Name, Path: in.Path, Scaffolded: scaffolded}, nil
}

type ConnectionsListInput struct {
	Prj *settings.Project
}

type ConnectionsListOutput struct {
	Connections []ConnectionEntry
}

func (in ConnectionsListInput) Valid(ctx context.Context) map[string]string {
	if in.Prj == nil {
		return map[string]string{"prj": "project not loaded"}
	}
	return nil
}

func ConnectionsList(ctx context.Context, in ConnectionsListInput) (*ConnectionsListOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(in.Prj.Connections))
	for name := range in.Prj.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]ConnectionEntry, 0, len(names))
	for _, name := range names {
		out = append(out, ConnectionEntry{Name: name, Path: in.Prj.Connections[name]})
	}
	return &ConnectionsListOutput{Connections: out}, nil
}

type ConnectionsDescribeInput struct {
	Prj  *settings.Project
	Name string
}

type ConnectionsDescribeOutput struct {
	Name   string
	Path   string
	Exists bool
	Conn   *opendpi.Connection
}

func (in ConnectionsDescribeInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func ConnectionsDescribe(ctx context.Context, in ConnectionsDescribeInput) (*ConnectionsDescribeOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	path, ok := in.Prj.Connections[in.Name]
	if !ok {
		return nil, fmt.Errorf("connection %q not found", in.Name)
	}

	abs := resolveProjectPath(in.Prj, path)
	out := &ConnectionsDescribeOutput{Name: in.Name, Path: path}
	if _, err := os.Stat(abs); err == nil {
		out.Exists = true
		if conn, err := opendpi.LoadConnection(abs); err == nil {
			out.Conn = conn
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return out, nil
}

type ConnectionsDeleteInput struct {
	Prj  *settings.Project
	Name string
}

type ConnectionsDeleteOutput struct {
	Name string
}

func (in ConnectionsDeleteInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func ConnectionsDelete(ctx context.Context, in ConnectionsDeleteInput) (*ConnectionsDeleteOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, ok := in.Prj.Connections[in.Name]; !ok {
		return nil, fmt.Errorf("connection %q not found", in.Name)
	}
	delete(in.Prj.Connections, in.Name)
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &ConnectionsDeleteOutput{Name: in.Name}, nil
}