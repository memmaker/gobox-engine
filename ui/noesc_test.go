package ui

import (
	"testing"

	"gobox/game"
)

// TestQuestionNoEsc: the game's NoEsc reaches the Ask.
func TestQuestionNoEsc(t *testing.T) {
	r := &frontState{}
	if a := r.question(&game.Question{YesNo: true, NoEsc: true}, false); !a.NoEsc {
		t.Error("NoEsc lost")
	}
}
