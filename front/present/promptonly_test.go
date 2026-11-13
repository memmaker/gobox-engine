package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// found is a message with no choice.
var found = &play.Ask{Kind: game.Choice, Prompt: "You found some loot!"}

// TestPromptOnlyAskDrawn: a message with no choice ("You found some
// loot!") is drawn until a key, never invisible.
func TestPromptOnlyAskDrawn(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	askedOn(found).drawPageMenu(dst, note)
	if s := strings.ToUpper(drawn.String()); !strings.Contains(s, "FOUND SOME LOOT") {
		t.Fatalf("drawn %q", drawn.String())
	}
}

// TestPromptOnlyAskEnter: Enter (or Space) dismisses a message with no
// choice: the game gets OK.
func TestPromptOnlyAskEnter(t *testing.T) {
	for _, k := range []uint16{'\r', ' '} {
		p := askedOn(found)
		p.Key(k)
		if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActOK {
			t.Errorf("key %d: acts %v, want OK", k, p.inv.acts)
		}
	}
}
