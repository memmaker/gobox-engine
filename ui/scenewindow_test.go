package ui

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestSceneWindowEmpty: a scene with nothing over the view (Rolf's tour
// between his pictures: the 3D view, the text below) shows the 3D view, not
// an empty whole-window text box; text over the view is WindowText.
func TestSceneWindowEmpty(t *testing.T) {
	if w := windowOf(play.Scene{Text: []string{"YOU NOW FACE AN ENTRANCE"}}); w != game.WindowScene {
		t.Errorf("scene with no pictures and no view text: window %v, want WindowScene", w)
	}
	if w := windowOf(play.Scene{Placed: []play.Placed{{At: image.Pt(2, 5), Text: "RUNES"}}}); w != game.WindowText {
		t.Errorf("scene with view text: window %v, want WindowText", w)
	}
}
