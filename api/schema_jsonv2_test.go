//go:build goexperiment.jsonv2

// This file is behind the jsonv2 build tag so that the module still builds and
// tests under GOEXPERIMENT=nojsonv2, the escape hatch back to the pre-1.27
// encoding/json implementation. jsontext exists only when jsonv2 is enabled
// (the default in Go 1.27), and under nojsonv2 json.RawMessage is a plain
// []byte with no IsValid method, so neither the import nor the call would
// compile. Fold this into schema_test.go once the opt-out is removed.

package api

import (
	"encoding/json/jsontext"
	"testing"
)

// TestResourceJsonSchemaNoDuplicateKeys guards the hand-written fallback schemas
// in schema.go against a duplicated object name.
//
// TestResourceJsonSchema cannot catch this: encoding/json keeps the last of a set
// of duplicate names without complaining, so a schema that accidentally defines
// "password" twice unmarshals cleanly with one definition silently discarded.
// Since Go 1.27 json.RawMessage is an alias for jsontext.Value, whose IsValid
// rejects duplicate names at any depth by default — exactly the check these blobs
// want, since they are edited by hand, run to several hundred lines, and only take
// effect against a broken v5.0 server, so a mistake would surface late and rarely.
func TestResourceJsonSchemaNoDuplicateKeys(t *testing.T) {
	for slug, schema := range ResourceSchemas {
		if schema.IsValid() {
			continue
		}
		// Narrow it down: valid once duplicates are permitted means the problem
		// is a duplicate name rather than a syntax error.
		if schema.IsValid(jsontext.AllowDuplicateNames(true)) {
			t.Errorf("Resource Schema %v contains a duplicate object name; one definition is being silently discarded", slug)
		} else {
			t.Errorf("Resource Schema %v is not valid JSON", slug)
		}
	}
}
