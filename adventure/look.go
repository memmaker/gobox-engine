package adventure

import (
	"encoding/json"
	"errors"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
)

// LookFile is the game's presentation values in the adventure's data folder
// (DataDir), written by the game's import; absent: every field is gobox's own.
const LookFile = "look.json"

// Look is what an original shows differently from the others, as data the
// front reads without knowing the game. A nil or empty field: gobox's own.
type Look struct {
	ReadyMark   string      `json:",omitempty"` // shown before a readied item's name
	Door        *color.RGBA `json:",omitempty"` // a door's colour on the maps
	AmigaGround *color.RGBA `json:",omitempty"` // the combat tiles' bare ground in the Amiga art (the DOS art's is black)
}

// ReadLook reads the adventure folder adv's LookFile; the zero Look when it has none.
func ReadLook(adv string) (Look, error) {
	b, err := os.ReadFile(filepath.Join(adv, DataDir, LookFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Look{}, nil
	}
	if err != nil {
		return Look{}, err
	}
	var l Look
	if err := json.Unmarshal(b, &l); err != nil {
		return Look{}, errors.Join(errors.New(LookFile), err)
	}
	return l, nil
}
