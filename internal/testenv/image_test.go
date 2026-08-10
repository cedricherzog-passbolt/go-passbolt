package testenv

import "testing"

// PassboltImage is what scripts/test-matrix.sh drives to replay the suite
// against each release, so the fallback and trimming behavior is worth
// pinning: a silently-ignored override would report one version's results
// under another version's name.
//
// The CLI's copy of passbolt.go carries identical logic; `make
// check-testenv-sync` is what keeps the two from drifting.
func TestPassboltImage(t *testing.T) {
	t.Run("defaults to the pinned image", func(t *testing.T) {
		t.Setenv(PassboltImageEnv, "")
		if got := PassboltImage(); got != defaultPassboltImage {
			t.Errorf("PassboltImage() = %q, want the pinned default %q", got, defaultPassboltImage)
		}
	})

	t.Run("honors the override", func(t *testing.T) {
		const want = "passbolt/passbolt:5.14.0-1-ce"
		t.Setenv(PassboltImageEnv, want)
		if got := PassboltImage(); got != want {
			t.Errorf("PassboltImage() = %q, want %q", got, want)
		}
	})

	t.Run("treats a blank override as unset", func(t *testing.T) {
		// Makefile recipes export PASSBOLT_TEST_IMAGE unconditionally, so an
		// unset override arrives as an empty (or whitespace) value rather than
		// an absent variable.
		t.Setenv(PassboltImageEnv, "   \t ")
		if got := PassboltImage(); got != defaultPassboltImage {
			t.Errorf("PassboltImage() = %q, want the pinned default %q", got, defaultPassboltImage)
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Setenv(PassboltImageEnv, "  passbolt/passbolt:5.0.0-1-ce\n")
		if got, want := PassboltImage(), "passbolt/passbolt:5.0.0-1-ce"; got != want {
			t.Errorf("PassboltImage() = %q, want %q", got, want)
		}
	})
}
