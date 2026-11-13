package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"gobox/asset"
	"gobox/ecl"
	"gobox/rules"
	"io/fs"
	"os"
	"path/filepath"
)

// POR's picture numbers as the shared collection's core.PicID (e.pics, the
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

// PortraitsFile is the portrait tables in the adventure's data folder
// (DataDir), written by the game's import.
const PortraitsFile = "portraits.json"

// GameTablesFile is the game's GameTables other than its portraits in the adventure's
// data folder (DataDir), written by the game's import.
const GameTablesFile = "gametables.json"

// ReadGameTables reads the game's tables from the adventure folder adv, as the
// game's import wrote them: GameTablesFile and the portrait tables (PortraitsFile);
// a game writes either or both. Script is the Dialect's.
func ReadGameTables(adv string) (GameTables, error) {
	var t GameTables
	found := false
	for _, f := range []string{GameTablesFile, PortraitsFile} {
		b, err := os.ReadFile(filepath.Join(adv, DataDir, f))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return GameTables{}, err
		}
		if err := json.Unmarshal(b, &t); err != nil {
			return GameTables{}, fmt.Errorf("%s: %w", f, err)
		}
		found = true
	}
	if !found {
		return GameTables{}, fmt.Errorf("%s and %s: not found (run the game's import)", GameTablesFile, PortraitsFile)
	}
	if t.Dialect != 0 {
		t.Script = ecl.Versions[ecl.VersionID(t.Dialect)]
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
