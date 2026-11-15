package ui

import (
	"fmt"
	"os"
	"slices"
	"testing"
)

// stuckFrames are the dead screens the run's frames drew (OnStuck), each once.
var stuckFrames []string

// TestMain fails the run when any test's frame was drawn stuck: a question
// of the game's on no screen (Presenter.stuck), whatever the test checked.
func TestMain(m *testing.M) {
	OnStuck = func(what string) {
		if !slices.Contains(stuckFrames, what) {
			stuckFrames = append(stuckFrames, what)
		}
	}
	code := m.Run()
	for _, s := range stuckFrames {
		fmt.Fprintln(os.Stderr, "FAIL: a frame drawn stuck:", s)
	}
	if len(stuckFrames) > 0 && code == 0 {
		code = 1
	}
	os.Exit(code)
}
