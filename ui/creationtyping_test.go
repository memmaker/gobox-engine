package ui

import (
	"strings"
	"testing"

	"gobox/game/play"
	"gobox/gpu"
)

// TestCreationTyping: a typed question over the creation (the name after
// Keep) is drawn, not hidden under the sheet.
func TestCreationTyping(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	p := draftAsked(&play.Draft{}, &play.Question{Text: "Character name: ", Typed: true, Max: 15})
	p.Draw(dst, note)
	if s := strings.ToUpper(drawn.String()); !strings.Contains(s, "CHARACTER NAME") {
		t.Fatalf("the name question is not drawn: %q", drawn.String())
	}
}
