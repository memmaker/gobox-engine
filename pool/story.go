package pool

import (
	"strings"

	"gobox/asset"
	"gobox/core"
	"gobox/ega"
)

// Story and housekeeping routines of ovr18, ovr19 and ovr34, and the duel
// set-up of ovr08.

// ProgramEight = ovr19:02A4: the Tyranthraxus finale. The FINAL pictures
// are drawn on the page, then animated; the dragon dies, Tyranthraxus
// speaks, and the screen is rebuilt. The script waits on it (PendStory)
// while the presenter plays the animations a pass at a time (StoryStep).
func (e *Engine) ProgramEight() {
	e.story = &storyRun{savedFlag: byte(e.worldWord(core.WordBeginFlag))}
	e.setWorldWord(core.WordBeginFlag, 7)
	e.songMusic(songEnding, 0) // extra
	e.storyStart()
	if e.storyPlay() {
		return
	}
	e.wait(PendStory) // the original spins; one pass a StoryStep
}

// storyRun is the finale's place: the animation n is cycling its frames.
type storyRun struct {
	n           int // animations started
	pass, loops int
	last        int64
	cur, count  int // rec[1] and rec[0]
	savedFlag   byte
}

// storyAnim is one animation (finalAnim = the nested routine at
// ovr19:0006: cycle the frames of FINAL<block> at (x,y) until loops
// passes are done), each after what comes before it (storyBefore).
type storyAnim struct {
	x, y, block int
	slow        bool // as many passes as the message delay says, else one
}

var storyAnims = []storyAnim{{3, 5, 2, true}, {3, 1, 7, true}, {3, 1, 8, false}}

// storyPage draws FINAL<set> block at page cell (x,y) (0464:0111, 0464:0D71 mode 0).
func (e *Engine) storyPage(block, x, y int) {
	e.drawPage(x, y, e.LoadPic(e.set.File("FINAL"), block), 0, false) // Str(DS:5376) -> DS:4713, never freed
}

// storySpeak says lines at row 0x11 down, and reads them aloud as one text
// (extra, voice.go).
func (e *Engine) storySpeak(lines ...string) {
	for i, l := range lines {
		e.text(1, 0x11+i, ega.LightGreen, l)
	}
	e.speak(voiceKey(strings.Join(lines, " ")))
}

// storyBefore is what comes before animation n.
func (e *Engine) storyBefore(n int) {
	switch n {
	case 0:
		e.clearPane()
		e.storySpeak("Mortally wounded, the dragon roars!", "The spirit of Tyranthraxus flares up", "from the dragon's body.")
		e.DrawMainScreen()
		e.storyPage(1, 0, 0)
	case 1:
		e.storyPage(3, 0, 0)
		if e.worldWord(core.WordPartyCount) > 1 {
			e.storyPage(4, 0, 0xB)
		}
		if e.worldWord(core.WordPartyCount) > 2 {
			e.storyPage(5, 0xD, 9)
		}
		if e.worldWord(core.WordPartyCount) > 3 {
			e.storyPage(6, 0, 9)
		}
		e.clearPane()
		e.storySpeak("\"Fools, you have but slain the body", "I possessed.  I cannot be defeated!\"", "With the power of the pool of radiance", "which I moved here, to my lair,", "I will still rule, by possessing one", "of you!\"")
	case 2:
		e.clearPane()
		e.storySpeak("\"No Lord Bane!  I can still rule here!", "I have not failed.  Do not call me back", "through the pool!\"", "Noooo...")
	}
}

// storyStart is finalAnim's set-up for the next animation.
func (e *Engine) storyStart() {
	st := e.story
	a := storyAnims[st.n]
	e.storyBefore(st.n)
	st.n++
	st.loops = 1
	if a.slow {
		st.loops = (10 - int(e.msgDelay)) * 5 >> 1
	}
	st.pass, st.last = 0, e.Ticks()
	e.LoadAnimPic("FINAL", a.block, false)
	st.cur, st.count = e.pic.cur, len(e.pic.frames)
	if st.count > 0 { // ovr19:004C: frame 1, unmasked, then 1B52
		e.drawPage(a.x-1, a.y-1, e.pic.frames[0].Image, 0, false)
	}
}

// storyStep is one pass of finalAnim's loop; true when it ends.
func (e *Engine) storyStep() bool {
	st := e.story
	a := storyAnims[st.n-1]
	var delay int64
	if st.cur >= 1 && st.cur <= st.count {
		e.drawPicAt(a.x, a.y, true, e.pic.frames[st.cur-1].Image) // ovr29:0000 flips
		delay = int64(e.pic.frames[st.cur-1].Delay)
	}
	now := e.Ticks()
	if now-st.last > int64(e.msgDelay+3)*delay {
		st.cur++
		if st.cur > st.count {
			st.cur = 1
			st.pass++
		}
		st.last = now
	}
	return st.pass == st.loops
}

// storyPlay goes on until the next frame is needed; true when the finale is over.
func (e *Engine) storyPlay() bool {
	st := e.story
	for e.storyStep() {
		e.FreeAnimPic()
		if st.n == len(storyAnims) {
			e.setWorldWord(core.WordBeginFlag, int(st.savedFlag))
			e.DrawMainScreen()
			return true
		}
		e.storyStart()
	}
	return false
}

// StoryStep plays the finale on a pass (the presenter paces it); over,
// the script goes on.
func (e *Engine) StoryStep() error {
	if !e.waiting(PendStory) {
		return core.ErrNotNow
	}
	if e.storyPlay() {
		e.resume()
	}
	return nil
}

// recolourFrame = 0464:1E5C (from, to, frame, random, image), the EGA
// path (DS:4A1E = 1) on chunky pixels: for each of the 16 colours whose
// from and to differ, pixels that had that colour in the picture as passed
// in become `to` in a copy, which replaces the picture at the end (so one
// call never chains A->B->C). With random set, each 8-pixel cell of a row
// changes only the pixels selected by Random(256) & Random(256), leftmost
// pixel in bit 7. A negative frame means every frame.
func (e *Engine) recolourFrame(from, to []byte, frame int, random bool, im *asset.Image) {
	if im == nil {
		return
	}
	// the frames change in place: the tiles' pictures follow their pixels (picRGBA)
	count := 1
	if frame < 0 {
		frame, count = 0, im.Count
	}
	if frame+count > im.Count {
		count = im.Count - frame
	}
	if count <= 0 {
		return
	}
	if im.Alt != nil && !random { // the DOS twin alike (random draws dice: the Amiga's alone)
		e.recolourFrame(from, to, frame, random, im.Alt)
	}
	lo, hi := frame*im.BPP, (frame+count)*im.BPP
	src := append([]byte(nil), im.Data[lo:hi]...)
	dst := im.Data[lo:hi]
	for i := 0; i < 16; i++ {
		if from[i] == to[i] {
			continue
		}
		for cell := 0; cell < count*im.Height*im.Width; cell++ {
			mask := byte(0xFF)
			if random {
				a := byte(e.rollRandom(0x100))
				mask = a & byte(e.rollRandom(0x100))
			}
			for k := 0; k < 8; k++ {
				if mask&(0x80>>k) != 0 && src[cell*8+k] == from[i] {
					dst[cell*8+k] = to[i]
				}
			}
		}
	}
}

// EndgameHook = ovr19:06B3 (n): the pool of radiance fades. With n set the
// area map is drawn first and the squares change in nine steps (randomised
// recolouring of the SQRPACI tiles); a final pass recolours them fully.
func (e *Engine) EndgameHook(n int) {
	tiles := func(to1, to2 *[16]byte, random bool) {
		for j := 0; j <= 0xA; j++ {
			e.recolourSquare(EndFrom[:], to1[:], EndTiles1[j], random)
		}
		for j := 0; j <= 0x11; j++ {
			e.recolourSquare(EndFrom[:], to2[:], EndTiles2[j], random)
		}
	}
	if n != 0 {
		e.DrawAreaMap(0x11, 0xF)
		for k := 0; k <= 8; k++ {
			tiles(&EndTo1, &EndTo2, true)
			e.DrawAreaMap(int(EndMapX[k]), int(EndMapY[k]))
		}
	}
	tiles(&EndTo1, &EndTo2, false)
	if n != 0 {
		e.DrawAreaMap(0x17, 0x19)
	}
}

// recolourSquare recolours one SQRPACI frame: values above 0x7F select the
// second picture (DS:5E56), the others the first (DS:5E52).
func (e *Engine) recolourSquare(from, to []byte, v byte, random bool) {
	if v > 0x7F {
		e.recolourFrame(from, to, int(v&0x7F), random, e.sqrpaci[1])
	} else {
		e.recolourFrame(from, to, int(v), random, e.sqrpaci[0])
	}
}

// DuelSetup = ovr08:1A6F (n): only the selected member fights (the others,
// by name, are made inactive). With n set a clone of the selected member,
// "ROLF" with the CPIC picture, is appended as the opponent and gets
// copies of the member's items, in reverse order.
func (e *Engine) DuelSetup(n int) {
	e.run.duelMode = 1
	e.setWorldWord(core.WordDuel, int(int16(int8(n))))
	selName := e.ruleset().Sheet(e.selected()).Name // selected() is never NoChar
	for _, id := range e.partyIDs() {
		if e.ruleset().Sheet(id).Name != selName { // 0AF8:0C10 compares the name strings
			e.do(core.SetInactive{Who: id})
		}
	}
	if n == 0 {
		return
	}
	e.LoadIconPic("CPIC", 0xB, int(e.es.cloneIcon)) // ovr33:01C8
	if _, err := e.ruleset().Resolve(nil, core.DuelClone{Of: e.selected(), Name: "ROLF", Icon: int32(e.es.cloneIcon)}, nil); err != nil {
		panic(err) // sel is a member
	}
}
