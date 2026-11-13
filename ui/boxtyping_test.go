package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestBoxQuestionTyping: a typed line whose question is in the text box (no
// Prompt of its own) is still drawn.
func TestBoxQuestionTyping(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	draw := func(ask *game.Question) int {
		dst := gpu.NewImage(1280, 800)
		defer dst.Deallocate()
		g, _ := playGame(game.Exploring{Common: game.Common{Adventure: true, Text: []string{"CHARACTER NAME:"}, Ask: ask}})
		p := &Presenter{}
		p.SetEngine(g)
		p.Draw(dst, note)
		return lit(dst)
	}
	if bare, typed := draw(nil), draw(&game.Question{Typed: true, Max: 15}); typed <= bare {
		t.Fatalf("the line being typed is not drawn: %d lit with it, %d without", typed, bare)
	}
}
