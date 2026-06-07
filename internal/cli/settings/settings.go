// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	UserDir     = ".daco"
	UserFile    = "settings.yaml"
	ProjectFile = "daco.yaml"
)

var (
	ErrProjectNotFound = errors.New("project file not found")
	ErrUserNotFound    = errors.New("user file not found")
)

type User struct {
	// Projects maps project IDs to their location.
	Projects map[string]string `yaml:"projects"`
}

func (u *User) ensureDefaults() {
	if u.Projects == nil {
		u.Projects = make(map[string]string)
	}
}

func userConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error fetching user home dir: %w", err)
	}
	return filepath.Join(home, UserDir, UserFile), nil
}

func LoadUser() (*User, error) {
	path, err := userConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err != nil {
		return nil, ErrUserNotFound
	}

	usr, err := load[User](path)
	if err != nil {
		return nil, err
	}
	usr.ensureDefaults()
	return usr, nil
}

func (u *User) Save() error {
	path, err := userConfigPath()
	if err != nil {
		return err
	}
	return save(path, u)
}

type Project struct {
	// Path tracks the absolute path to the project file on disk.
	Path string `yaml:"-"`

	// Name of the product.
	Name string `yaml:"name"`

	// Products maps product IDs to their location.
	Products map[string]string `yaml:"products"`

	// Connections maps connection IDs to their location.
	Connections map[string]string `yaml:"connections"`

	// Schemas maps schema IDs to their location.
	Schemas map[string]string `yaml:"schemas"`
}

func (p *Project) ensureDefaults() {
	if p.Products == nil {
		p.Products = make(map[string]string)
	}
	if p.Connections == nil {
		p.Connections = make(map[string]string)
	}
	if p.Schemas == nil {
		p.Schemas = make(map[string]string)
	}
}

func LoadProject(usr *User) (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error fetching current working directory: %w", err)
	}
	path := filepath.Join(cwd, ProjectFile)

	if _, err := os.Stat(path); err == nil {
		prj, err := load[Project](path)
		if err != nil {
			return nil, err
		}
		prj.Path = path
		prj.ensureDefaults()
		return prj, nil
	}

	if usr == nil || len(usr.Projects) == 0 {
		return nil, ErrProjectNotFound
	}

	bestMatch := ""
	maxLen := -1

	for _, prjPath := range usr.Projects {
		absPrjPath, err := filepath.Abs(prjPath)
		if err != nil {
			continue
		}

		rel, err := filepath.Rel(absPrjPath, cwd)
		if err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
			infPath := filepath.Join(absPrjPath, ProjectFile)
			if _, err := os.Stat(infPath); err == nil {
				if len(absPrjPath) > maxLen {
					maxLen = len(absPrjPath)
					bestMatch = infPath
				}
			}
		}
	}

	if bestMatch == "" {
		return nil, ErrProjectNotFound
	}

	prj, err := load[Project](bestMatch)
	if err != nil {
		return nil, err
	}
	prj.Path = bestMatch
	prj.ensureDefaults()
	return prj, nil
}

// LoadProjectAt loads a project from an explicit directory (the dir containing
// daco.yaml), bypassing cwd and the user registry. Used by the TUI to open a
// project picked from the registry without chdir'ing.
func LoadProjectAt(dir string) (*Project, error) {
	path := filepath.Join(dir, ProjectFile)
	prj, err := load[Project](path)
	if err != nil {
		return nil, err
	}
	prj.Path = path
	prj.ensureDefaults()
	return prj, nil
}

func (p *Project) Save() error {
	if p.Path == "" {
		return errors.New("cannot save: project configuration has no path assigned")
	}
	return save(p.Path, p)
}

func save[T any](path string, out *T) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(out)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", path, err)
	}

	return nil
}

func load[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	var out T
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("error parsing YAML in %s: %w", path, err)
	}

	return &out, nil
}