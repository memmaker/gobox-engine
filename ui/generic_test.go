package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// lit counts the pixels of dst that are not black.
func lit(dst *gpu.Image) (n int) {
	px := make([]byte, 4*dst.Bounds().Dx()*dst.Bounds().Dy())
	dst.ReadPixels(px)
	for i := 0; i < len(px); i += 4 {
		if px[i]|px[i+1]|px[i+2] != 0 {
			n++
		}
	}
	return
}

// A state with only pane text and a "press Enter" Ask (the text after a
// fight, no Adventure, no Battle): the text shows in the pane, the prompt
// on the page.
func TestGenericPaneAndPrompt(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	won := game.Common{Text: []string{"THE PARTY HAS WON."}}
	p, f := presenterOn(game.Scene{Common: won})
	p.Draw(dst, note)
	textOnly := lit(dst)
	if textOnly == 0 {
		t.Fatal("the pane text is not drawn")
	}
	won.Ask = &game.Question{Text: "PRESS ENTER", Replies: []string{"CONTINUE"}}
	f.st = game.Scene{Common: won}
	p.Draw(dst, note)
	if both := lit(dst); both <= textOnly {
		t.Fatalf("the prompt is not drawn: %d lit with it, %d without", both, textOnly)
	}
}

// A typed answer with nothing else up (BR2's name after the sheet) is
// drawn: its question and line, never a black screen.
func TestGenericTyping(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	p, _ := presenterOn(game.Scene{Common: game.Common{Ask: &game.Question{Text: "Character name:", Typed: true, Max: 15}}})
	p.Draw(dst, note)
	if lit(dst) == 0 {
		t.Fatal("the name question is not drawn")
	}
}

// TestInventoryGameAsk: a question the game asks while its items are up
// (COTAB's "Use it?") is drawn by Draw, not hidden under the inventory.
func TestInventoryGameAsk(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	p, _ := presenterOn(game.Exploring{Common: game.Common{Party: []game.Member{{Name: "A"}}, Ask: &game.Question{Text: "Use it? ", Replies: []string{"Yes", "No"}}}})
	p.eng.r.items = game.ItemsOwn
	if p.Inventory(dst, Note{}) {
		t.Fatal("the items hide the game's question")
	}
}
