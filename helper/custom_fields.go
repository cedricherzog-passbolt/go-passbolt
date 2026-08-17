package helper

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// CustomField is one custom field of a v5 resource, merged from its metadata and
// secret halves by id.
type CustomField struct {
	ID    string
	Name  string // metadata_key, else the decrypted secret_key
	Value string // secret_value, else metadata_value
}

// CustomFields is a resource's custom fields, in metadata order.
type CustomFields []CustomField

// Map flattens custom fields to name -> value, dropping unnamed ones. On duplicate
// names the last wins.
func (c CustomFields) Map() map[string]string {
	out := make(map[string]string, len(c))
	for _, cf := range c {
		if cf.Name != "" {
			out[cf.Name] = cf.Value
		}
	}
	return out
}

// validateCustomFields validates custom_fields arrays in metadata and secret maps
// before encryption. This enforces the same rules as the Passbolt web extension:
//   - Every custom field id must be a valid UUID
//   - Every custom field id must appear in both metadata and secret arrays
//   - metadata entries must have metadata_key
//   - secret entries must have secret_value
//   - key/value must not be defined on both sides for the same field id
//
// Returns nil if neither map contains custom_fields (not a custom fields resource).
func validateCustomFields(metadataFields, secretFields map[string]any) error {
	metaCF, metaOK := extractCustomFields(metadataFields)
	secretCF, secretOK := extractCustomFields(secretFields)

	// Not a custom fields resource
	if !metaOK && !secretOK {
		return nil
	}

	// If one side has custom_fields, the other must too
	if metaOK != secretOK {
		return fmt.Errorf("%w: custom_fields present in one side but not the other", ErrCustomFieldIDMismatch)
	}

	// Build indexes by id
	metaByID := make(map[string]map[string]any, len(metaCF))
	for i, cf := range metaCF {
		id, ok := cf["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("%w: metadata custom_fields[%d] has no id", ErrCustomFieldInvalidID, i)
		}
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("%w: %q (metadata custom_fields[%d])", ErrCustomFieldInvalidID, id, i)
		}
		if _, exists := metaByID[id]; exists {
			return fmt.Errorf("%w: duplicate id %q in metadata custom_fields", ErrCustomFieldInvalidID, id)
		}
		metaByID[id] = cf
	}

	secretByID := make(map[string]map[string]any, len(secretCF))
	for i, cf := range secretCF {
		id, ok := cf["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("%w: secret custom_fields[%d] has no id", ErrCustomFieldInvalidID, i)
		}
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("%w: %q (secret custom_fields[%d])", ErrCustomFieldInvalidID, id, i)
		}
		if _, exists := secretByID[id]; exists {
			return fmt.Errorf("%w: duplicate id %q in secret custom_fields", ErrCustomFieldInvalidID, id)
		}
		secretByID[id] = cf
	}

	// Strict: every id must appear in both arrays
	if len(metaByID) != len(secretByID) {
		return fmt.Errorf("%w: metadata has %d fields, secret has %d", ErrCustomFieldIDMismatch, len(metaByID), len(secretByID))
	}
	for id := range metaByID {
		if _, ok := secretByID[id]; !ok {
			return fmt.Errorf("%w: id %q is in metadata but not in secret", ErrCustomFieldIDMismatch, id)
		}
	}

	// Validate each field
	for id, metaEntry := range metaByID {
		secretEntry := secretByID[id]

		// metadata entry must have metadata_key
		if _, ok := metaEntry["metadata_key"]; !ok {
			return fmt.Errorf("%w: id %q", ErrCustomFieldMissingKey, id)
		}

		// secret entry must have secret_value
		if _, ok := secretEntry["secret_value"]; !ok {
			return fmt.Errorf("%w: id %q", ErrCustomFieldMissingValue, id)
		}

		// Cross-field: key must be on one side only
		metaHasKey := hasNonEmptyString(metaEntry, "metadata_key")
		secretHasKey := hasNonEmptyString(secretEntry, "secret_key")
		if metaHasKey && secretHasKey {
			return fmt.Errorf("%w: id %q has both metadata_key and secret_key", ErrCustomFieldCrossField, id)
		}

		// Cross-field: value must be on one side only
		metaHasValue := hasNonEmptyString(metaEntry, "metadata_value")
		secretHasValue := hasNonEmptyString(secretEntry, "secret_value")
		if metaHasValue && secretHasValue {
			return fmt.Errorf("%w: id %q has both metadata_value and secret_value", ErrCustomFieldCrossField, id)
		}
	}

	return nil
}

// ParseCustomFields merges the custom_fields arrays of the metadata and secret field
// maps returned by GetResourceFieldMaps. The metadata array decides which fields exist
// and in what order, so a resource with none there returns nil. Malformed entries are
// projected as best they can be rather than rejected, since the server cannot validate
// encrypted content.
func ParseCustomFields(metadataFields, secretFields map[string]any) CustomFields {
	metaCF, _ := extractCustomFields(metadataFields)
	if len(metaCF) == 0 {
		return nil
	}
	secretCF, _ := extractCustomFields(secretFields)

	// Duplicate ids: last wins. Blank ids are left out rather than paired with each
	// other; the schema requires the key but permits "".
	secretByID := make(map[string]map[string]any, len(secretCF))
	for _, cf := range secretCF {
		if id := GetStringField(cf, "id"); id != "" {
			secretByID[id] = cf
		}
	}

	out := make(CustomFields, 0, len(metaCF))
	for _, meta := range metaCF {
		id := GetStringField(meta, "id")
		secret := secretByID[id] // nil when unmatched; reads below are nil-safe

		// A name lives on one side only: cleartext in metadata_key, or encrypted
		// in secret_key.
		name := GetStringField(meta, "metadata_key")
		if name == "" {
			name = GetStringField(secret, "secret_key")
		}

		// First non-empty wins: a cleartext field still carries an empty
		// secret_value, so keying on presence would shadow metadata_value.
		value := stringifyCustomFieldValue(secret["secret_value"])
		if value == "" {
			value = stringifyCustomFieldValue(meta["metadata_value"])
		}

		out = append(out, CustomField{ID: id, Name: name, Value: value})
	}

	return out
}

// stringifyCustomFieldValue renders a custom field value as a string. Passbolt allows
// string, number, boolean or null; anything else is malformed.
func stringifyCustomFieldValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// extractCustomFields extracts the custom_fields array from a field map.
func extractCustomFields(fields map[string]any) ([]map[string]any, bool) {
	raw, ok := fields["custom_fields"]
	if !ok {
		return nil, false
	}
	arr, ok := raw.([]any)
	if !ok {
		// Already typed as []map[string]any (unlikely from JSON but handle it)
		if typed, ok := raw.([]map[string]any); ok {
			return typed, true
		}
		return nil, false
	}
	result := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result, len(result) > 0
}

// hasNonEmptyString checks if a map entry exists and is a non-empty string.
func hasNonEmptyString(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}
