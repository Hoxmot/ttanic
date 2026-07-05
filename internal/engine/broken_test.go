package engine

import "testing"

// TestBrokenGate exists only to prove CI fails on failing tests; this file is reverted.
func TestBrokenGate(t *testing.T) {
	t.Fatal("deliberate failure to prove the CI gate")
}
