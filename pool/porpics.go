package pool

import (
	"encoding/json"
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/importer/pictures"
	"os"
	"path/filepath"
)

// POR's picture numbers as the shared collection's core.PicID (e.pics, the
// adventure's data/pictures.json) and back: PortHead and PortBody index
// portraitHead and portraitBody, IconHead and IconBody are CHEAD and CBODY
// blocks (the large icon's 0x40 on). "" and false: no such picture. Number 0
// is none (pictures.Set's none).

// portraitPics is the portrait table of family "HEAD" or "BODY".
func (e *Engine) portraitPics(family string) []goldbox.PortraitPic {
	if family == "HEAD" {
		return e.game().PortraitHead
	}
	return e.game().PortraitBody
}

// portPic is portrait number b of family ("HEAD", "BODY") as a PicID.
func (e *Engine) portPic(family string, b byte) core.PicID {
	t := e.portraitPics(family)
	if b == 0 || int(b) >= len(t) {
		return ""
	}
	return e.pics.ID(pictures.DAXRef{File: t[b].Set.File(family) + ".DAX", Block: int(t[b].Id), Frame: pictures.NoFrame})
}

// portByte is portPic's inverse: the first number showing id.
func (e *Engine) portByte(family string, id core.PicID) (byte, bool) {
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
func (e *Engine) iconPic(file string, b, size byte) core.PicID {
	return e.pics.ID(pictures.DAXRef{File: file + ".DAX", Block: int(b) + iconBlockOff(size), Frame: pictures.NoFrame})
}

// iconByte is iconPic's inverse at size.
func (e *Engine) iconByte(file string, id core.PicID, size byte) (byte, bool) {
	for b := 0; id != "" && b < 0x40; b++ {
		if e.iconPic(file, byte(b), size) == id {
			return byte(b), true
		}
	}
	return 0, false
}

// PortraitsFile is the portrait tables in the adventure's data folder
// (DataDir), written by the game's import.
const PortraitsFile = "portraits.json"

// ReadGameTables reads the game's tables from the adventure folder adv: the
// portrait tables (PortraitsFile), as the game's import wrote them.
func ReadGameTables(adv string) (GameTables, error) {
	b, err := os.ReadFile(filepath.Join(adv, DataDir, PortraitsFile))
	if err != nil {
		return GameTables{}, fmt.Errorf("%w (run the game's import)", err)
	}
	var t GameTables
	if err := json.Unmarshal(b, &t); err != nil {
		return GameTables{}, fmt.Errorf("%s: %w", PortraitsFile, err)
	}
	return t, nil
}

// TreasureSetsFile is GameTables.TreasureSets in the adventure's data folder (DataDir),
// written by the game's import (pod).
const TreasureSetsFile = "treasuresets.json"

// ReadTreasureSets reads the treasure sets (TreasureSetsFile) from the adventure folder adv.
func ReadTreasureSets(adv string) ([][]int32, error) {
	b, err := os.ReadFile(filepath.Join(adv, DataDir, TreasureSetsFile))
	if err != nil {
		return nil, fmt.Errorf("%w (run the game's import)", err)
	}
	var sets [][]int32
	if err := json.Unmarshal(b, &sets); err != nil {
		return nil, fmt.Errorf("%s: %w", TreasureSetsFile, err)
	}
	return sets, nil
}
