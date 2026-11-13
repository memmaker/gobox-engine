package importer

import (
	"path/filepath"
	"reflect"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// rawBlit is one frame-drawing call as the Amiga program makes it (the
// pushed y and x, in 2-pixel units from 8000 above 6000) and the piece's
// header (hot spot hotY, hotX; size w x h).
type rawBlit struct {
	at          string // the call
	entry       int    // the library entry
	y, x        int
	hotY, hotX  int
	w, h        int
	masked      bool
	repeatYFrom int // a loop over y: from, to, step (0: one call)
	repeatYTo   int
	repeatStep  int
	pairEntry   int // a loop's second call at each y: this entry (8x16, no hot spot) at x pairX (0: none)
	pairX       int
}

// framePx is the program's coordinate in pixels (Curse F_3e520).
func framePx(v int) int {
	const scaled, origin, unit = 6000, 8000, 2
	if v > scaled {
		return (v - origin) * unit
	}
	return v
}

// TestAmigaFrameLayout: each game's AmigaFrame screens are the calls its
// frame routines make, in order, each piece at the call's position less
// its hot spot, and the imported pieces have the entries' sizes.
func TestAmigaFrameLayout(t *testing.T) {
	const o = 0x1f40 // 8000: pixel 0
	// Curse 0x398b0 (F_3bcf8 calls)
	curseFull := []rawBlit{
		{at: "0x398dc", entry: 71, y: o, x: o, w: 320, h: 8},
		{at: "0x398fa/0x39912", entry: 72, y: o + 4, x: o, w: 8, h: 16, repeatYFrom: o + 4, repeatYTo: o + 19*4, repeatStep: 8, pairEntry: 74, pairX: 0x1fdc}, // d2 1..19 step 2: y 8000+4*d2
		{at: "0x39930", entry: 73, y: 0x1f94, x: o, w: 8, h: 16},
		{at: "0x39942", entry: 106, y: 0x1f94, x: 0x1fdc, w: 8, h: 16},
		{at: "0x39954", entry: 75, y: o, x: o, hotY: -184, w: 320, h: 16},
	}
	curseView := []rawBlit{
		{at: "0x399a6", entry: 98, y: o, x: o, hotY: -16, hotX: -16, w: 104, h: 104},
		{at: "0x399b8", entry: 77, y: o, x: 0x1f80, w: 8, h: 8},
		{at: "0x399d6", entry: 76, y: o + 4, x: 0x1f80, w: 8, h: 16, repeatYFrom: o + 4, repeatYTo: o + 13*4, repeatStep: 8},
		{at: "0x399f2", entry: 96, y: 0x1f80, x: o, w: 320, h: 8},
		{at: "0x39a04", entry: 78, y: 0x1f7c, x: 0x1f80, w: 8, h: 16},
	}
	curseRule := []rawBlit{{at: "0x39b10", entry: 96, y: 0x1f80, x: o, w: 320, h: 8}}
	// Silver Blades 0x3da76 (F_3ff40 calls); 0x3da58(16): entry 41 at y 8000+4*16
	sotsbFull := []rawBlit{
		{at: "0x3da9e", entry: 37, y: o, x: o, w: 320, h: 8},
		{at: "0x3dab0", entry: 38, y: o, x: o, w: 8, h: 184},
		{at: "0x3dac2", entry: 39, y: o, x: 0x1fdc, w: 8, h: 183},
		{at: "0x3dad4", entry: 40, y: o, x: o, hotY: -184, w: 320, h: 16},
	}
	sotsbRule := rawBlit{at: "0x3da6e", entry: 41, y: o + 4*16, x: o, w: 320, h: 7, masked: true}
	sotsbView := []rawBlit{sotsbRule, {at: "0x3db4c", entry: 68, y: o, x: o, w: 136, h: 135}}
	// Pools of Darkness 0x3df80 (F_405e6 calls, library 1: the FRAME set)
	podFull := []rawBlit{
		{at: "0x3dfaa", entry: 1, y: o, x: o, w: 320, h: 8},
		{at: "0x3dfbe", entry: 2, y: o, x: o, w: 8, h: 184},
		{at: "0x3dfd2", entry: 3, y: o, x: o, hotX: -312, w: 8, h: 184},
		{at: "0x3dfe6", entry: 4, y: o, x: o, hotY: -184, w: 320, h: 16},
	}
	podRule := rawBlit{at: "0x3df4e", entry: 9, y: o, x: o, hotY: -128, w: 320, h: 7}
	podView := []rawBlit{podRule, {at: "0x3e04a", entry: 12, y: o, x: o, w: 136, h: 136}}
	cat := func(ls ...[]rawBlit) []rawBlit {
		var out []rawBlit
		for _, l := range ls {
			out = append(out, l...)
		}
		return out
	}
	for _, c := range []struct {
		game    string
		frame   *game.AmigaFrame
		screens [][]rawBlit
		frameOf func(entry int) (block, frame int) // the imported piece of an entry
	}{
		{"cotab", cotabAmigaFrame(), [][]rawBlit{curseFull, cat(curseFull, curseView), cat(curseFull, curseRule)}, func(e int) (int, int) { return e, 0 }},
		{"sotsb", sotsbAmigaFrame(), [][]rawBlit{sotsbFull, cat(sotsbFull, sotsbView), cat(sotsbFull, []rawBlit{sotsbRule})}, func(e int) (int, int) { return e, 0 }},
		{"pod", podAmigaFrame(), [][]rawBlit{podFull, cat(podFull, podView), cat(podFull, []rawBlit{podRule})}, func(e int) (int, int) { return podAmigaFrameSets[0], e - 1 }},
	} {
		art := asset.OpenArt(filepath.Join("..", "data", c.game, "art"))
		if !art.Has(asset.Amiga) {
			t.Skipf("%s: no imported Amiga art", c.game)
		}
		var want [][]game.FrameBlit
		for _, raws := range c.screens {
			var s []game.FrameBlit
			for _, r := range raws {
				ys := []int{r.y}
				if r.repeatStep > 0 {
					ys = nil
					for y := r.repeatYFrom; y <= r.repeatYTo; y += r.repeatStep {
						ys = append(ys, y)
					}
				}
				block, frame := c.frameOf(r.entry)
				for _, y := range ys {
					s = append(s, game.FrameBlit{Block: block, Frame: frame, X: framePx(r.x) - r.hotX, Y: framePx(y) - r.hotY, Masked: r.masked})
					if r.pairEntry != 0 {
						pb, pf := c.frameOf(r.pairEntry)
						s = append(s, game.FrameBlit{Block: pb, Frame: pf, X: framePx(r.pairX), Y: framePx(y)})
					}
				}
				for _, set := range append([]int{block}, c.frame.Sets...) {
					b := art.Block(asset.Amiga, c.frame.File, set)
					if b == nil || frame >= len(b.Frames) {
						t.Errorf("%s %s: no imported piece %s block %d frame %d", c.game, r.at, c.frame.File, set, frame)
						continue
					}
					if im := b.Frames[frame].Image; im.Width*8 != r.w || im.Height != r.h {
						t.Errorf("%s %s: entry %d (block %d frame %d) is %dx%d, want %dx%d", c.game, r.at, r.entry, set, frame, im.Width*8, im.Height, r.w, r.h)
					}
				}
			}
			want = append(want, s)
		}
		if !reflect.DeepEqual(c.frame.Screens, want) {
			t.Errorf("%s: screens\n%v\nwant\n%v", c.game, c.frame.Screens, want)
		}
	}
}
