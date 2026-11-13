package engine

import (
	"gobox/asset"
	"gobox/rules"
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
