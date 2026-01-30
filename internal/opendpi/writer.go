package opendpi

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

// Writer encodes an OpenDPI spec to a file.
type Writer struct {
	write     func(path string, v any) error
	extension string
}

var (
	// JSONWriter writes OpenDPI specs as JSON.
	JSONWriter = Writer{writeJSON, ".json"}
	// YAMLWriter writes OpenDPI specs as YAML.
	YAMLWriter = Writer{writeYaml, ".yaml"}
)

// Write encodes the spec to the given path.
func (wr Writer) Write(spec *Spec, specDir string) error {
	panic("not implemented")
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeYaml(path string, v any) error {
	f, err := os.Create(path) //nolint:gosec // path is from config
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(v)
}
