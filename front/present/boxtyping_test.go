package present

import (
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestBoxQuestionTyping: a typed line whose question is in the text box (no
// Prompt of its own) is still drawn.
func TestBoxQuestionTyping(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	draw := func(ask *front.Ask) int {
		dst := gpu.NewImage(1280, 800)
		defer dst.Deallocate()
		p := &Presenter{}
		p.Draw(dst, note, front.View{Adventure: true, Pane: &front.Pane{Lines: [][]front.PaneRun{{{Text: "CHARACTER NAME:"}}}}, Ask: ask})
		return lit(dst)
	}
	if bare, typed := draw(nil), draw(&front.Ask{Kind: front.AskText, Max: 15, Text: "ZED"}); typed <= bare {
		t.Fatalf("the line being typed is not drawn: %d lit with it, %d without", typed, bare)
	}
}
