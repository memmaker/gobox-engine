package present

import (
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestCreationTyping: a typed question over the creation sheet (the name
// after Keep) is drawn over it, not hidden under the sheet.
func TestCreationTyping(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	litWith := func(ask *front.Ask) int {
		dst := gpu.NewImage(1280, 800)
		defer dst.Deallocate()
		p := &Presenter{}
		p.Draw(dst, note, front.View{Creation: &front.Creation{}, Ask: ask})
		return lit(dst)
	}
	bare := litWith(nil)
	if asked := litWith(&front.Ask{Kind: front.AskText, Max: 15, Menu: front.Menu{Prompt: "Character name: "}}); asked <= bare {
		t.Fatalf("the name question is not drawn: %d lit with it, %d without", asked, bare)
	}
}
