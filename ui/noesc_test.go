package ui

import (
	"testing"

	"gobox/game/play"
)

// TestQuestionNoEsc: the game's NoEsc reaches the Ask.
func TestQuestionNoEsc(t *testing.T) {
	r := &frontState{}
	if a := r.question(&play.Question{YesNo: true, NoEsc: true}, false); !a.NoEsc {
		t.Error("NoEsc lost")
	}
}
