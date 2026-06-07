// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dacolabs/daco/internal/cli/settings"
)

type InitInput struct {
	Usr  *settings.User
	Cwd  string
	Name string
}

type InitOutput struct {
	ProjectPath string
	Name        string
	Created     bool
}

func (in InitInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Usr == nil {
		problems["usr"] = "user settings not loaded"
	}
	if in.Cwd == "" {
		problems["cwd"] = "required"
	}
	return problems
}

func Init(ctx context.Context, in InitInput) (*InitOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}

	absCwd, err := filepath.Abs(in.Cwd)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	cfgPath := filepath.Join(absCwd, settings.ProjectFile)

	if _, err := os.Stat(cfgPath); err == nil {
		if in.Name != "" {
			return nil, fmt.Errorf("%s already exists in this directory", settings.ProjectFile)
		}
		prj, err := settings.LoadProject(in.Usr)
		if err != nil {
			return nil, err
		}
		in.Usr.Projects[prj.Name] = absCwd
		if err := in.Usr.Save(); err != nil {
			return nil, err
		}
		return &InitOutput{ProjectPath: cfgPath, Name: prj.Name, Created: false}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	prj := &settings.Project{
		Path:        cfgPath,
		Name:        in.Name,
		Products:    map[string]string{},
		Connections: map[string]string{},
		Schemas:     map[string]string{},
	}
	if err := prj.Save(); err != nil {
		return nil, err
	}
	in.Usr.Projects[in.Name] = absCwd
	if err := in.Usr.Save(); err != nil {
		return nil, err
	}
	return &InitOutput{ProjectPath: cfgPath, Name: in.Name, Created: true}, nil
}