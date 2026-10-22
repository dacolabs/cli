package apply

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dacolabs/cli/internal/catalogapi"
	"gopkg.in/yaml.v3"
)

// Unit is one Dataset to apply, identified by urn+version.
type Unit struct {
	Source string
	Input  catalogapi.DatasetInput
}

type rawDoc struct {
	Kind string `yaml:"kind"`
}

// Load reads -f paths (files or recursive dirs of *.yaml/*.yml), splits on ---,
// and merges Dataset documents. Conflicting unequal duplicates fail.
func Load(paths []string) ([]Unit, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("apply requires at least one -f path")
	}
	files, err := expandPaths(paths)
	if err != nil {
		return nil, err
	}
	type keyed struct {
		unit Unit
		raw  []byte
	}
	byKey := map[string]keyed{}
	order := []string{}

	for _, file := range files {
		docs, err := splitYAMLDocs(file)
		if err != nil {
			return nil, err
		}
		for i, doc := range docs {
			src := fmt.Sprintf("%s#%d", file, i+1)
			var head rawDoc
			if err := yaml.Unmarshal(doc, &head); err != nil {
				return nil, fmt.Errorf("%s: %w", src, err)
			}
			if head.Kind == "" {
				return nil, fmt.Errorf("%s: missing kind", src)
			}
			if head.Kind != "Dataset" {
				return nil, fmt.Errorf("%s: unknown kind %q", src, head.Kind)
			}
			var input catalogapi.DatasetInput
			if err := yaml.Unmarshal(doc, &input); err != nil {
				return nil, fmt.Errorf("%s: %w", src, err)
			}
			norm, err := canonicalJSON(input)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", src, err)
			}
			key := input.Urn + "\x00" + string(input.Version)
			if prev, ok := byKey[key]; ok {
				if !bytes.Equal(prev.raw, norm) {
					return nil, fmt.Errorf("conflict: dataset %s@version %s declared differently in %s and %s",
						input.Urn, input.Version, prev.unit.Source, src)
				}
				continue
			}
			byKey[key] = keyed{unit: Unit{Source: src, Input: input}, raw: norm}
			order = append(order, key)
		}
	}

	out := make([]Unit, 0, len(order))
	for _, key := range order {
		out = append(out, byKey[key].unit)
	}
	return out, nil
}

func expandPaths(paths []string) ([]string, error) {
	var files []string
	seen := map[string]struct{}{}
	add := func(p string) {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		files = append(files, p)
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			add(p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".yaml" || ext == ".yml" {
				add(path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML files found in -f paths")
	}
	return files, nil
}

func splitYAMLDocs(path string) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var docs [][]byte
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if node.Kind == 0 && len(node.Content) == 0 {
			continue
		}
		buf, err := yaml.Marshal(&node)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if len(bytes.TrimSpace(buf)) == 0 {
			continue
		}
		docs = append(docs, buf)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("%s: empty YAML", path)
	}
	return docs, nil
}

func canonicalJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var normalized any
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}
