package programs

import (
	"fmt"
	"strings"
)

// ParamExtractor provides type-safe parameter extraction from LLM-provided
// parameter maps. Handles common type conversions and provides sensible defaults.
type ParamExtractor struct {
	params map[string]interface{}
}

// NewParamExtractor creates a new parameter extractor for the given params map
func NewParamExtractor(params map[string]interface{}) *ParamExtractor {
	return &ParamExtractor{params: params}
}

// String extracts a string parameter from the params map.
// Returns empty string if parameter doesn't exist or isn't a string.
func (e *ParamExtractor) String(name string) string {
	if val, ok := e.params[name]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// StringLower extracts a string parameter and converts it to lowercase.
// Useful for case-insensitive matching (contexts, statuses, etc.)
func (e *ParamExtractor) StringLower(name string) string {
	return strings.ToLower(e.String(name))
}

// StringTrim extracts a string parameter and trims whitespace
func (e *ParamExtractor) StringTrim(name string) string {
	return strings.TrimSpace(e.String(name))
}

// Int extracts an integer parameter from the params map.
// Handles float64 (JSON default), int, and string conversions.
// Returns 0 if parameter doesn't exist or can't be converted.
func (e *ParamExtractor) Int(name string) int {
	if val, ok := e.params[name]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return 0
}

// Bool extracts a boolean parameter from the params map.
// Returns false if parameter doesn't exist or isn't a boolean.
func (e *ParamExtractor) Bool(name string) bool {
	if val, ok := e.params[name]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Has checks if a parameter exists in the map
func (e *ParamExtractor) Has(name string) bool {
	_, ok := e.params[name]
	return ok
}

// Require validates that one or more required parameters are present and non-empty.
// Returns error if any parameter is missing or empty.
//
// Example:
//   if err := extractor.Require("command", "project_id"); err != nil {
//       return nil, err
//   }
func (e *ParamExtractor) Require(names ...string) error {
	for _, name := range names {
		val := e.String(name)
		if val == "" {
			// Check if it's an int that might be valid
			if intVal := e.Int(name); intVal != 0 {
				continue
			}
			return fmt.Errorf("required parameter missing or empty: %s", name)
		}
	}
	return nil
}

// RequireAny validates that at least one of the provided parameters is present.
// Returns error if all parameters are missing or empty.
//
// Example:
//   if err := extractor.RequireAny("project_id", "project_name"); err != nil {
//       return nil, err
//   }
func (e *ParamExtractor) RequireAny(names ...string) error {
	for _, name := range names {
		if e.String(name) != "" || e.Int(name) != 0 {
			return nil
		}
	}
	return fmt.Errorf("at least one required parameter must be provided: %v", names)
}

// ==================== Legacy Helper Functions ====================
// These functions maintain backward compatibility with existing code.
// New code should use ParamExtractor instead.

// getStringParam extracts a string parameter from the params map.
// Returns empty string if parameter doesn't exist or isn't a string.
//
// Deprecated: Use ParamExtractor.String() instead
func getStringParam(params map[string]interface{}, name string) string {
	extractor := NewParamExtractor(params)
	return extractor.String(name)
}

// getIntParam extracts an integer parameter from the params map.
// Handles float64 (JSON default), int, and string conversions.
// Returns 0 if parameter doesn't exist or can't be converted.
//
// Deprecated: Use ParamExtractor.Int() instead
func getIntParam(params map[string]interface{}, name string) int {
	extractor := NewParamExtractor(params)
	return extractor.Int(name)
}

// getBoolParam extracts a boolean parameter from the params map.
// Returns false if parameter doesn't exist or isn't a boolean.
//
// Deprecated: Use ParamExtractor.Bool() instead
func getBoolParam(params map[string]interface{}, name string) bool {
	extractor := NewParamExtractor(params)
	return extractor.Bool(name)
}
