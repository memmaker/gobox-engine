package ui

import (
	"testing"

	"gobox/game"
)

// TestSceneWindowEmpty: a scene with nothing over the view (Rolf's tour
// between his pictures: the 3D view, the text below) shows the 3D view, not
// an empty whole-window text box; text over the view is WindowText.
func TestSceneWindowEmpty(t *testing.T) {
	if w := windowOf(game.Scene{Text: []string{"YOU NOW FACE AN ENTRANCE"}}); w != game.WindowScene {
		t.Errorf("scene with no pictures and no view text: window %v, want WindowScene", w)
	}
	if w := windowOf(game.Scene{Placed: [][]game.TextRun{{{Text: "RUNES"}}}}); w != game.WindowText {
		t.Errorf("scene with view text: window %v, want WindowText", w)
	}
}
