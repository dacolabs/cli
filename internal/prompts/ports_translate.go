package prompts

import "github.com/charmbracelet/huh"

// RunTranslateFormatSelect returns a select field for choosing translation output format.
func RunTranslateFormatSelect(value *string, formats []string) *huh.Select[string] {
	options := make([]huh.Option[string], len(formats))
	for i, f := range formats {
		options[i] = huh.NewOption(f, f)
	}
	return huh.NewSelect[string]().
		Title("Output format").
		Options(options...).
		Value(value)
}
