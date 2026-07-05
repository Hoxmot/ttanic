package engine

import "os"

// brokenGate exists only to prove CI fails on lint errors; this file is reverted.
func brokenGate() {
	f, err := os.Open("nope")
	if err != nil {
		return
	}
	f.Close()
}
