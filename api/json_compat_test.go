package api

import (
	"encoding/json"
	"testing"
	"time"
)

// Go 1.27 reimplemented encoding/json on top of encoding/json/v2, keeping the v1
// API and v1 semantics. These tests pin the decoder behaviors this package
// actually depends on, so a future change in the backend (or the removal of the
// GOEXPERIMENT=nojsonv2 opt-out) shows up here rather than in production.
//
// They are deliberately not behind a build tag: they must hold under both
// backends. Run them both ways:
//
//	go test ./api/
//	GOEXPERIMENT=nojsonv2 go test ./api/

// TestJSONDuplicateNames_LastWins pins that a duplicated object name is accepted
// and the last occurrence wins, for both struct and map targets.
//
// This matters because these payloads are server-controlled: response envelopes,
// resource metadata, and decrypted secret JSON all arrive from the Passbolt
// server. v2's own semantics reject duplicate names outright, so if the v1
// compatibility layer ever stopped applying, every one of those decodes would
// begin erroring instead of silently taking the last value. Pinning it makes that
// a test failure rather than a field-reported outage.
//
// Note this is a statement of current behavior, not an endorsement: a server
// sending "password" twice is anomalous, and last-wins means the earlier value is
// discarded without a signal.
func TestJSONDuplicateNames_LastWins(t *testing.T) {
	t.Parallel()

	t.Run("into a struct", func(t *testing.T) {
		t.Parallel()
		var got APIHeader
		if err := json.Unmarshal([]byte(`{"message":"first","message":"second"}`), &got); err != nil {
			t.Fatalf("Unmarshal: unexpected error: %v", err)
		}
		if got.Message != "second" {
			t.Errorf("Message = %q, want %q (last duplicate wins)", got.Message, "second")
		}
	})

	t.Run("into a map", func(t *testing.T) {
		t.Parallel()
		// map[string]any is how decrypted secret and metadata JSON is decoded.
		var got map[string]any
		if err := json.Unmarshal([]byte(`{"password":"first","password":"second"}`), &got); err != nil {
			t.Fatalf("Unmarshal: unexpected error: %v", err)
		}
		if got["password"] != "second" {
			t.Errorf("password = %v, want %q (last duplicate wins)", got["password"], "second")
		}
	})

	t.Run("nested in a metadata object", func(t *testing.T) {
		t.Parallel()
		var got map[string]any
		raw := `{"object_type":"PASSBOLT_RESOURCE_METADATA","custom":{"k":"a","k":"b"}}`
		if err := json.Unmarshal([]byte(raw), &got); err != nil {
			t.Fatalf("Unmarshal: unexpected error: %v", err)
		}
		nested, ok := got["custom"].(map[string]any)
		if !ok {
			t.Fatalf("custom is %T, want map[string]any", got["custom"])
		}
		if nested["k"] != "b" {
			t.Errorf("custom.k = %v, want %q", nested["k"], "b")
		}
	})
}

// TestJSONRawMessage_StaysRawAndUnescaped pins that json.RawMessage still round
// trips bytes verbatim. APIResponse.Body is a RawMessage that is decoded a second
// time by every caller, and in Go 1.27 RawMessage became an alias for
// jsontext.Value, so this is worth stating explicitly.
func TestJSONRawMessage_StaysRawAndUnescaped(t *testing.T) {
	t.Parallel()

	// HTML-significant characters in a secret must survive an envelope round trip
	// unchanged, and the envelope's own encoding must not double-escape them.
	const body = `{"password":"a<b>c&d"}`
	env := APIResponse{
		Header: APIHeader{Status: "success", Code: 200},
		Body:   json.RawMessage(body),
	}

	encoded, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("Marshal envelope: %v", err)
	}

	var round APIResponse
	if err := json.Unmarshal(encoded, &round); err != nil {
		t.Fatalf("Unmarshal envelope: %v", err)
	}

	var secret map[string]string
	if err := json.Unmarshal(round.Body, &secret); err != nil {
		t.Fatalf("Unmarshal body %q: %v", round.Body, err)
	}
	if secret["password"] != "a<b>c&d" {
		t.Errorf("password = %q, want %q", secret["password"], "a<b>c&d")
	}
}

// TestTime_UnmarshalJSON_RejectsEscapedInput pins a limitation of Time's
// hand-rolled unquoting.
//
// UnmarshalJSON is handed the raw JSON bytes and does
// strings.Trim(string(buf), `"`) rather than a real JSON unquote, so any escape
// sequence in the value survives into time.Parse and the parse fails. Both
// backends agree on this, so it is not a v1-versus-v2 difference.
//
// It is harmless for the format actually in use - RFC3339 contains no character
// that needs escaping, so a well-formed timestamp never arrives escaped - and
// failing closed is the right direction for a decoder. Pinned so that a future
// rewrite of Time.UnmarshalJSON to a real unquote is a deliberate, visible
// change rather than an accidental loosening.
func TestTime_UnmarshalJSON_RejectsEscapedInput(t *testing.T) {
	t.Parallel()

	// The canonical form parses.
	want := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	var normal Time
	if err := json.Unmarshal([]byte(`"2024-03-01T12:00:00Z"`), &normal); err != nil {
		t.Fatalf("Unmarshal canonical: %v", err)
	}
	if !normal.Equal(want) {
		t.Errorf("canonical parsed to %v, want %v", normal.Time, want)
	}

	// A JSON string carrying escapes does not: the backslashes reach time.Parse.
	// Note Trim strips the outer quotes but cannot undo the \" escapes.
	var escaped Time
	err := json.Unmarshal([]byte(`"\"2024-03-01T12:00:00Z\""`), &escaped)
	if err == nil {
		t.Fatalf("Unmarshal of an escaped timestamp succeeded, giving %v; want a parse error", escaped.Time)
	}
	if !escaped.IsZero() {
		t.Errorf("failed Unmarshal left a non-zero time %v; want the zero value", escaped.Time)
	}
}
