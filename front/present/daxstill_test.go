package present

import (
	"testing"
	"time"

	"gobox/asset"
	"gobox/front"
)

// A two-frame animation block shows frame 0 for its delay (tenths of a second), then frame 1, then loops.
func TestSpriteFrameDelays(t *testing.T) {
	a := &spriteAnim{delays: []time.Duration{300 * time.Millisecond, 100 * time.Millisecond}}
	t0 := time.Unix(0, 0)
	a.start, a.seen = t0, t0
	spriteAnims[daxKey{file: "PIC9", block: 1}] = a
	old := daxArt
	daxArt.art = asset.OpenArt(t.TempDir())
	defer func() { daxArt = old; delete(spriteAnims, daxKey{file: "PIC9", block: 1}) }()
	for _, c := range []struct {
		ms, want int
	}{{0, 0}, {100, 0}, {200, 0}, {300, 1}, {400, 0}} {
		if f, ok := spriteFrame(front.Picture{File: "PIC9", Block: 1}, t0.Add(time.Duration(c.ms)*time.Millisecond)); !ok || f != c.want {
			t.Errorf("%d ms: frame %d %v, want %d", c.ms, f, ok, c.want)
		}
	}
}

// animFrame plays the frames once from the first, then over from loop.
func TestAnimFrameLoop(t *testing.T) {
	d := []time.Duration{100, 100, 100, 100} // ns
	for _, c := range []struct {
		t          time.Duration
		loop, want int
	}{{0, 0, 0}, {350, 0, 3}, {450, 0, 0}, {250, 2, 2}, {450, 2, 2}, {550, 2, 3}, {650, 2, 2}} {
		if f := animFrame(d, c.loop, c.t); f != c.want {
			t.Errorf("t %d loop %d: frame %d, want %d", c.t, c.loop, f, c.want)
		}
	}
}

// A sprite (SPRIT*) block never plays itself: the game picks its frame.
func TestSpriteNotAutoPlayed(t *testing.T) {
	if _, ok := spriteFrame(front.Picture{File: "SPRIT1", Block: 1}, time.Now()); ok {
		t.Fatal("SPRIT played")
	}
}

// A VGA still (PoD's TITLE pictures) is not an animation: never a frame
// past its one, which decodes as nothing (the title went black).
func TestVGAStillNotAnimated(t *testing.T) {
	if !useArt(t, "../../data/pod/sources/dos") || artBlock("TITLE", 2) == nil {
		t.Skip("no PoD data")
	}
	if f, ok := spriteFrame(front.Picture{File: "TITLE", Block: 2}, time.Now()); ok {
		t.Fatalf("a VGA still plays as an animation (frame %d)", f)
	}
}
