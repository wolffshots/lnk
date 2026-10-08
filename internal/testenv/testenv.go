// Package testenv provides environment helpers for tests.
package testenv

import "testing"

// SetHome points os.UserHomeDir at dir for the duration of the test.
// Go reads HOME on Unix and USERPROFILE on Windows, so both are set.
func SetHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}
