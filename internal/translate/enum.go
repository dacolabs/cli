// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import "fmt"

// EnumStringSymbols returns the constraint's enum as []string if every value is a string,
// otherwise nil. Targets with native enum types (Avro, protobuf) only accept string identifiers.
func EnumStringSymbols(c Constraints) []string {
	if len(c.Enum) == 0 {
		return nil
	}
	out := make([]string, 0, len(c.Enum))
	for _, v := range c.Enum {
		s, ok := v.(string)
		if !ok {
			return nil
		}
		out = append(out, s)
	}
	return out
}

// FormatLiteral renders a single value as a Python/TypeScript-style literal —
// strings get double-quoted, booleans become True/False, numbers print plainly.
func FormatLiteral(v any) string {
	switch x := v.(type) {
	case string:
		return fmt.Sprintf("%q", x)
	case bool:
		if x {
			return "True"
		}
		return "False"
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return fmt.Sprintf("%v", x)
	case nil:
		return "None"
	default:
		return fmt.Sprintf("%v", x)
	}
}
