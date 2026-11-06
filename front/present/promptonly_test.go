package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestPromptOnlyAskDrawn: a message with no choice ("You found some
// loot!") is drawn until a key, never invisible.
func TestPromptOnlyAskDrawn(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	v := front.View{Exploring: true, Ask: &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "You found some loot!"}}}
	(&Presenter{}).drawPageMenu(dst, note, v)
	if s := strings.ToUpper(drawn.String()); !strings.Contains(s, "FOUND SOME LOOT") {
		t.Fatalf("drawn %q", drawn.String())
	}
}

// TestPromptOnlyAskEnter: Enter (or Space) dismisses a message with no
// choice: the game gets OK.
func TestPromptOnlyAskEnter(t *testing.T) {
	v := front.View{Adventure: true, Pictures: []front.Picture{{}}, Ask: &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "You found some loot!"}}}
	for _, k := range []uint16{'\r', ' '} {
		p := &Presenter{shownView: v}
		p.Key(k, v)
		if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActOK {
			t.Errorf("key %d: acts %v, want OK", k, p.inv.acts)
		}
	}
}
