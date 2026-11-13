package engine

import (
	"gobox/asset"
	"gobox/rules"
)

// Story and housekeeping routines of ovr18, ovr19 and ovr34, and the duel
// set-up of ovr08.

// StoryStep is front.Game's finale pass: POR's finale is its manifest's End now (SetEnd,
// ending), so no pass is ever due.
func (e *Engine) StoryStep() error { return rules.ErrNotNow }

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
	e.tilesChanged()
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
	e.setWorldWord(rules.WordDuel, int(int16(int8(n))))
	selName := e.ruleset().Sheet(e.selected()).Name // selected() is never NoChar
	for _, id := range e.partyIDs() {
		if e.ruleset().Sheet(id).Name != selName { // 0AF8:0C10 compares the name strings
			e.do(rules.SetInactive{Who: id})
		}
	}
	if n == 0 {
		return
	}
	e.LoadIconPic("CPIC", 0xB, int(e.es.cloneIcon)) // ovr33:01C8
	if _, err := e.ruleset().Resolve(nil, rules.DuelClone{Of: e.selected(), Name: "ROLF", Icon: int32(e.es.cloneIcon)}, nil); err != nil {
		panic(err) // sel is a member
	}
}
