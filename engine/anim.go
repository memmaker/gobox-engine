package engine

import (
	"slices"
	"time"

	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// Pictures shown in the view window, ported from ovr29: animated PIC /
// SPRIT / FINAL blocks and the HEAD+BODY portraits.

// animPic = the DS:6ABE record: up to 0x42 bytes holding a frame count and
// per frame a delay and a picture pointer.
type animPic struct {
	name   string // DS:6B0A cache key
	block  int    // DS:6B1C (-1 = nothing loaded)
	frames []asset.Frame
	cur    int   // DS:6ABF frame shown while waiting for input (1-based)
	last   int64 // ovr26:03BD [bp-57]: tick of the last frame change in the current DisplayInput
}

// portrait = DS:6B00..6B09: the cached HEAD and BODY pictures.
type portrait struct {
	headID, bodyID   int       // DS:6B00 / 6B05 (-1 = none)
	headSet, bodySet rules.Set // the sets they came from
	head, body       *asset.Image
}

// picFamily is the view window's pictures' file family (PIC<set>.DAX).
const picFamily = "PIC"

// animFile is the file of family name's block: the set in play's, or for a
// picture the adventure keeps elsewhere (GameTables.PictureSets) that set's.
func (e *Engine) animFile(name string, block int) string {
	if s, ok := e.game().PictureSets[block]; ok && name == picFamily {
		return s.File(name)
	}
	return e.set.File(name)
}

// LoadAnimPic = ovr29:00A2 (name, block, masked): load <name><gfx>.DAX
// block into the view-window picture, unless it is already there. PIC and
// FINAL frames after the first are stored XORed against the first; the
// first PIC block is cut to 4 frames. Sprites are masked on colour 0.
func (e *Engine) LoadAnimPic(name string, block int, masked bool) {
	a := &e.pic
	if a.name == name && a.block == block {
		return
	}
	e.FreeAnimPic()
	a.name, a.block = name, block
	file := e.animFile(name, block)
	b, dos := e.artBlocks(file, block)
	if b == nil {
		return
	}
	mask := asset.NoMask
	if masked {
		mask = 0
	}
	fr := b.Anim(mask) // nil: no animation block (EGA's or VGA's)
	if name == "PIC" && block == 1 && len(fr) > 4 {
		fr = fr[:4]
	}
	var alt []asset.Frame // the DOS animation an Amiga one stands in for (F12)
	if dos != nil {
		alt = dos.Anim(mask)
	}
	for i, f := range fr {
		named(f.Image, file, block).Base = i
		if len(alt) > 0 { // ponytail: frame for frame, the Amiga's count and delays; the DOS one's own timing if they ever differ visibly
			twin(f.Image, named(alt[min(i, len(alt)-1)].Image, file, block))
		}
	}
	if masked && b.Kind == asset.Anim && (e.Cfg.Graphics == 'E' || e.Cfg.Graphics == 'T') { // DS:4A1E != 0; EGA art's colours only (a VGA block's indices are its own palette's)
		for _, f := range fr { // 0464:1E5C(DS:0C6C, DS:0C7C, 0, 0): colour 13 -> opaque black
			e.recolourFrame(maskFrom[:], maskTo[:], 0, false, f.Image)
		}
	}
	a.frames = fr
	a.cur = 1
}

var maskFrom = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} // DS:0C6C
var maskTo = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 0, 14, 15}    // DS:0C7C

// FreeAnimPic = ovr29:048F.
func (e *Engine) FreeAnimPic() {
	e.pic = animPic{block: -1}
}

// drawPicAt = ovr29:0000 (x, y, overlay, pic): the picture in the view
// window (the original's cell (x,y) and overlay are where and how it drew
// it: the presenter's).
func (e *Engine) drawPicAt(im *asset.Image) {
	if im == nil {
		return
	}
	e.showPic(im)
}

// nextPictureFrame = pod's CALL 3: the view picture steps to its next frame, wrapping at the
// frame count, and is drawn as the animation draws it.
func (e *Engine) nextPictureFrame() {
	a := &e.pic
	if len(a.frames) == 0 {
		return
	}
	a.cur = a.cur%len(a.frames) + 1
	e.drawPicAt(a.frames[a.cur-1].Image)
}

// DrawAnimFrame = ovr29:0708 (pic, frame): frame 1..3 of a sprite at its
// own cell offset inside the view window.
func (e *Engine) DrawAnimFrame(frame int) {
	if frame < 1 || frame > 3 {
		e.PromptRow24(game.InkWhite, "Illegal range in Show3DSprite.")
		e.Exit()
		return
	}
	if frame > len(e.pic.frames) {
		return
	}
	e.drawPicAt(e.pic.frames[frame-1].Image) // 0464:0D71 at (x+3-1, y+3-1): the art's own cell offset (game.Picture.CellX)
}

// DrawAnimFirst draws frame 1 of the loaded picture the way the PICTURE
// and APPROACH code does: ovr29:0000(3, 3, overlay, frame 1).
func (e *Engine) DrawAnimFirst() {
	if len(e.pic.frames) > 0 {
		e.drawPicAt(e.pic.frames[0].Image)
	}
}

// LoadPortrait = ovr29:05D9 (head, body): set's HEAD/BODY blocks, kept
// while they do not change; 0xFF leaves that half alone.
func (e *Engine) LoadPortrait(set rules.Set, head, body int) {
	p := &e.port
	if head != 0xFF && (p.headID != head || p.headSet != set) {
		p.head = e.LoadPic(set.File("HEAD"), head)
		p.headID, p.headSet = head, set
	}
	if body != 0xFF && (p.bodyID != body || p.bodySet != set) {
		p.body = e.LoadPic(set.File("BODY"), body)
		p.bodyID, p.bodySet = body, set
	}
}

// DrawPortrait = ovr29:0697 (3, 3, both): the head and the body under it
// in the view window (nested drawers ovr29:0050 / ovr29:0070).
func (e *Engine) DrawPortrait() {
	if e.port.head == nil && e.port.body == nil {
		return
	}
	e.setWin(WinPicture)
	e.snap.shown = viewShown{portrait: true}
}

// ShowPortrait = ovr08:0511 (head, body).
func (e *Engine) ShowPortrait(head, body int) {
	e.es.picAnimating = 0
	e.es.picHead, e.es.picBody = byte(head), byte(body)
	e.LoadPortrait(e.set, head, body)
	e.DrawPortrait()
}

// ShowEncounter = ovr08:054E (sprite, pic, dist, flags): draw the approaching
// monster sprite for its distance, then at distance 0 its close-up picture
// (or the portrait game[0x5C2] when one is set). flags is DS:84C8: [0] the
// sprite is loaded, [1] the picture is shown.
func (e *Engine) ShowEncounter(sprite, pic, dist int, flags *[2]byte) {
	if flags[1] == 0 && flags[0] != 0 {
		e.Redraw()
	}
	frame := dist + 1
	if flags[1] == 0 {
		if flags[0] == 0 {
			e.LoadAnimPic("SPRIT", sprite, true)
			flags[0] = 1
			e.es.spriteLoaded = 1
		}
		if e.viewMode == 1 {
			e.DrawAnimFrame(frame)
		} else {
			e.DrawAnimFrame(2)
		}
	}
	face := uint16(e.world.Portrait)
	if flags[1] != 0 && uint16(e.es.shownFace) == face {
		return
	}
	if dist != 0 || e.viewMode != 1 || (e.es.inEncounterMenu != 0 && e.viewBit1 == 0) {
		return
	}
	e.es.shownFace = byte(face)
	e.es.picShown = 1
	if face == 0xFF {
		e.LoadAnimPic("PIC", pic, false)
		flags[1] = 1
		e.DrawAnimFirst()
	} else {
		e.ShowPortrait(int(face), pic)
		flags[1] = 1
		e.es.picAnimating = 0
	}
}

// animating: the picture animates while the engine waits.
func (e *Engine) animating() bool {
	a := &e.pic
	return (e.viewAnim != 0 || e.animInput) && a.cur != 0 && a.cur <= len(a.frames)
}

// animateWhileWaiting = the wait loop of ovr26:03BD: until a key arrives,
// cycle the view picture's frames at (3,3), each shown for its delay in
// tenths of a second (area+0x3FE forces a change every pass). Runs when
// the caller asked for animation or area+0x3FE is set; false when nothing animates.
func (e *Engine) animateWhileWaiting() bool {
	a := &e.pic
	force := e.viewAnim != 0
	if !e.animating() {
		return false
	}
	// ovr26:04CC: the current frame is redrawn on every poll
	e.drawPicAt(a.frames[a.cur-1].Image)
	if e.Ticks()-a.last >= int64(a.frames[a.cur-1].Delay)*10 || force {
		a.cur++
		if a.cur > len(a.frames) {
			a.cur = 1
		}
		a.last = e.Ticks()
	}
	return true
}

// animFrameTime is an animation frame's Delay unit (tenths of a second);
// animNextPoll is a 0 delay's time: the wait loop's next poll.
const (
	animFrameTime = 100 * time.Millisecond
	animNextPoll  = 10 * time.Millisecond
)

// viewState is what the view window shows as the State says it: the
// picture up (the big picture over it), the frames it cycles while the
// wait animates it, or the encounter portrait.
func (e *Engine) viewState() (pic *game.Picture, frames []game.Picture, delays []time.Duration, pt *game.Portrait) {
	if im := e.es.bigPic; im != nil {
		g := framePicture(im, 0, -1)
		return &g, nil, nil, nil
	}
	sh := e.snap.shown
	if sh.portrait {
		return nil, nil, nil, &game.Portrait{Head: framePicture(e.port.head, 0, -1), Body: framePicture(e.port.body, 0, -1)}
	}
	if sh.pic == nil {
		return nil, nil, nil, nil
	}
	g := framePicture(sh.pic, 0, sh.pic.Mask)
	if !e.animating() || !slices.ContainsFunc(e.pic.frames, func(f asset.Frame) bool { return f.Image == sh.pic }) {
		return &g, nil, nil, nil
	}
	for _, f := range e.pic.frames {
		frames = append(frames, framePicture(f.Image, 0, f.Image.Mask))
		delays = append(delays, max(time.Duration(f.Delay)*animFrameTime, animNextPoll))
	}
	return &g, frames, delays, nil
}
