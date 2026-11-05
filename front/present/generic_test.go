package present

import (
	"image/color"
	"testing"

	"gobox/front"
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

// A View with only pane text and a "press Enter" Ask (the text after a
// fight, no Adventure, no Battle): the text shows in the pane, the prompt
// on the page.
func TestGenericPaneAndPrompt(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	white := color.RGBA{255, 255, 255, 255}
	pane := &front.Pane{Lines: [][]front.PaneRun{{{Text: "THE PARTY HAS WON.", Fg: white, Bg: color.RGBA{A: 255}}}}}
	ask := &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "PRESS ENTER", Items: []front.MenuItem{{Text: "CONTINUE"}}}}
	p := &Presenter{}
	p.Draw(dst, note, front.View{Pane: pane})
	textOnly := lit(dst)
	if textOnly == 0 {
		t.Fatal("the pane text is not drawn")
	}
	p.Draw(dst, note, front.View{Pane: pane, Ask: ask})
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
	p := &Presenter{}
	p.Draw(dst, note, front.View{Ask: &front.Ask{Kind: front.AskText, Menu: front.Menu{Prompt: "Character name:"}, Max: 15}})
	if lit(dst) == 0 {
		t.Fatal("the name question is not drawn")
	}
}
