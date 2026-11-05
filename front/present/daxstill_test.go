package present

import (
	"testing"
	"time"

	"gobox/dax"
	"gobox/front"
)

// An animation block (PIC*: u8 count, a 21-byte frame header, packed pixels) decodes as its frame, not as nothing.
func TestDaxStillSprite(t *testing.T) {
	b := []byte{1, 0, 0, 0, 0, 2, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b = append(b, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0)
	im := daxStill("PIC1", b, 0)
	if im == nil || im.Width != 1 || im.Height != 2 || im.Frame(0)[0] != 1 {
		t.Fatalf("%+v", im)
	}
}

// A two-frame animation block shows frame 0 for its delay (tenths of a second), then frame 1, then loops.
func TestSpriteFrameDelays(t *testing.T) {
	a := &spriteAnim{delays: []time.Duration{300 * time.Millisecond, 100 * time.Millisecond}}
	t0 := time.Unix(0, 0)
	a.start, a.seen = t0, t0
	spriteAnims[daxKey{file: "PIC9", block: 1}] = a
	old := daxLoader
	daxLoader = &dax.Loader{}
	defer func() { daxLoader = old; delete(spriteAnims, daxKey{file: "PIC9", block: 1}) }()
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
