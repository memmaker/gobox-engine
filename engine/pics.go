package engine

import (
	"gobox/asset"
	"gobox/rules"
	"slices"
)

// The game's picture numbers (GameTables' portrait tables, CHEAD/CBODY) as the shared collection's core.PicID (e.pics, the
// adventure's data/pictures.json) and back: PortHead and PortBody index
// portraitHead and portraitBody, IconHead and IconBody are CHEAD and CBODY
// blocks (the large icon's 0x40 on). "" and false: no such picture. Number 0
// is none (pictures.Set's none).

// portraitPics is the portrait table of family "HEAD" or "BODY".
func (e *Engine) portraitPics(family string) []rules.PortraitPic {
	if family == "HEAD" {
		return e.game().PortraitHead
	}
	return e.game().PortraitBody
}

// portPic is portrait number b of family ("HEAD", "BODY") as a PicID.
func (e *Engine) portPic(family string, b byte) rules.PicID {
	t := e.portraitPics(family)
	if b == 0 || int(b) >= len(t) {
		return ""
	}
	return e.pics.ID(asset.DAXRef{File: t[b].Set.File(family) + ".DAX", Block: int(t[b].Id), Frame: asset.NoFrame})
}

// portByte is portPic's inverse: the first number showing id.
func (e *Engine) portByte(family string, id rules.PicID) (byte, bool) {
	for b := 1; id != "" && b < len(e.portraitPics(family)); b++ {
		if e.portPic(family, byte(b)) == id {
			return byte(b), true
		}
	}
	return 0, false
}

// iconBlockOff is where size's blocks start in CHEAD and CBODY (LoadIconPic).
func iconBlockOff(size byte) int {
	if size == 2 {
		return 0x40
	}
	return 0
}

// iconPic is combat icon part b of file ("CHEAD", "CBODY") at size as a PicID.
func (e *Engine) iconPic(file string, b, size byte) rules.PicID {
	return e.pics.ID(asset.DAXRef{File: file + ".DAX", Block: int(b) + iconBlockOff(size), Frame: asset.NoFrame})
}

// iconByte is iconPic's inverse at size.
func (e *Engine) iconByte(file string, id rules.PicID, size byte) (byte, bool) {
	for b := 0; id != "" && b < 0x40; b++ {
		if e.iconPic(file, byte(b), size) == id {
			return byte(b), true
		}
	}
	return 0, false
}

// charHead is the head the player chose for c, as the sheet shows it
// (SheetPortrait), as View pixels in the EGA's own colours (Member.HeadPix;
// none without one).
func (e *Engine) charHead(cID rules.CharID) Picture {
	h := e.fact(cID, rules.PortraitFact{}).(rules.Portrait).Head
	if h == 0 || h >= len(e.game().PortraitHead) {
		return Picture{}
	}
	return e.head(e.game().PortraitHead[h])
}

// head is portrait head p's pixels (cached by id, one picture in every
// set; none if missing).
func (e *Engine) head(p rules.PortraitPic) Picture {
	id := int(p.Id)
	if DOSArt.Load() {
		id |= 1 << 16 // the DOS twin's
	}
	if pc, ok := e.snap.heads[id]; ok {
		return pc
	}
	if e.snap.heads == nil {
		e.snap.heads = map[int]Picture{}
	}
	var out Picture
	im := e.art(e.LoadPic(p.Set.File("HEAD"), int(p.Id)), 0)
	if im != nil && im.Count > 0 {
		out = Picture{Pix: slices.Clone(im.Frame(0)), PixW: im.Width * 8, Palette: im.Palette}
	}
	e.snap.heads[id] = out
	return out
}

// borderTiles is DrawBorder's style 0 cells (0x114 corner, 0x115 side,
// 0x116 top), taken once the symbols are loaded, as the border's pieces by
// place (ui.TopLeft ..): View pixels in the EGA's own colours.
func (e *Engine) borderTiles() [8]Picture {
	if cells := e.artNames().BorderCells; e.snap.tiles[0].Pix == nil && e.sym[4] != nil && len(cells) == len(e.snap.tiles) {
		im := e.sym[4]
		for i, c := range cells {
			e.snap.tiles[i] = Picture{Pix: slices.Clone(im.Frame(c - symFix[4])), PixW: im.Width * 8, Palette: im.Palette}
		}
	}
	c, side, top := e.snap.tiles[0], e.snap.tiles[1], e.snap.tiles[2]
	return [8]Picture{c, c, c, c, top, top, side, side}
}
