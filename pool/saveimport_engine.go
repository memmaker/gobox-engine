package pool

import (
	"errors"
	"gobox/char"
	"gobox/core/goldbox"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"

	"gobox/ecl/script"
	grules "gobox/rules"
)

// readSlotSave is the canonical save of the save directory; a slot holding
// only an original save is imported.
func (e *Engine) readSlotSave() (gameSave, error) {
	g, err := readGameSave(e.savePath(gameFile))
	if !errors.Is(err, fs.ErrNotExist) {
		return g, err
	}
	b, err := os.ReadFile(e.savePath(origSaveFile))
	if err != nil {
		return g, err
	}
	return e.importSave(b)
}

// importSlots gives every slot holding only an original save its canonical
// save, once a session: the slot list sees the canonical saves only.
func (e *Engine) importSlots() {
	if e.slotsImported {
		return
	}
	e.slotsImported = true
	ents, err := os.ReadDir(e.slotRoot())
	if err != nil {
		return // no slots yet
	}
	for _, en := range ents {
		dir := filepath.Join(e.slotRoot(), en.Name())
		if !en.IsDir() {
			continue
		}
		e.inDir(dir, func() {
			if _, err := os.Stat(e.savePath(gameFile)); !errors.Is(err, fs.ErrNotExist) {
				return
			}
			if _, err := os.Stat(e.savePath(origSaveFile)); err != nil {
				return // not a slot
			}
			g, err := e.readSlotSave()
			if err == nil {
				err = writeGameSave(filepath.Join(dir, gameFile), g)
			}
			if err != nil {
				log.Printf("save slot %s not imported: %v", en.Name(), err)
			}
		})
	}
}

// importSave turns an original SAVGAMA.DAT image into the canonical save:
// gobox/char reads the image (ReadOriginalSave); the fields' cells come
// out of its record words by their layout addresses (the game's fields, and
// the locals of the area it was saved in).
func (e *Engine) importSave(b []byte) (gameSave, error) {
	o, err := char.ReadOriginalSave(b)
	if err != nil {
		return gameSave{}, err
	}
	g := gameSave{
		Set: goldbox.Set(o.Set), Rules: o.Rules, Fields: fieldCells{}, Locals: fieldCells{},
		X: o.X, Y: o.Y, Facing: o.Facing, Pos3: o.Pos3, Pos4: o.Pos4,
		ViewMode: o.ViewMode, BeginFlag: o.BeginFlag,
	}
	for _, fl := range e.gameFields() {
		if c := decodeCells(&o, fl); c != nil {
			g.Fields[fl.Name] = c
		}
	}
	l, err := script.ReadLayout(os.DirFS(e.scriptDir()), int(o.AreaKind))
	if err != nil {
		return gameSave{}, err
	}
	for _, fl := range l.Fields {
		if fl.Owner == script.OwnerLocal && e.saved(fl) {
			if c := decodeCells(&o, fl); c != nil {
				g.Locals[fl.Name] = c
			}
		}
	}
	for _, n := range o.Party {
		g.Party = append(g.Party, grules.MakeFileName(goldbox.RemoveSubstr(n, ".")))
	}
	return g, nil
}

// origSaveFile is the original's save of a slot (ovr18 SaveGame), read only
// here: the import turns it into the canonical save (savegame.go).
const origSaveFile = char.OriginalSaveFile

// recordCell is the save record word at script address a, nil outside the
// records.
func recordCell(o *char.ImportedSave, a int) *uint16 {
	switch {
	case a >= areaVarBase && a < areaVarBase+len(o.Area):
		return &o.Area[a-areaVarBase]
	case a >= gameVarBase && a < gameVarBase+len(o.Game):
		return &o.Game[a-gameVarBase]
	case a >= auxVarBase && a < auxVarBase+len(o.Aux):
		return &o.Aux[a-auxVarBase]
	}
	return nil
}

// decodeCells are field fl's cells in o's records (a text up to its NUL);
// nil when all are zero.
func decodeCells(o *char.ImportedSave, fl script.Field) []int64 {
	var c []int64
	n := fl.Len
	if fl.Width == script.WidthText {
		n = 0x100
	}
	for i := range n {
		p := recordCell(o, fl.From+i)
		if p == nil || fl.Width == script.WidthText && *p == 0 {
			break
		}
		c = append(c, int64(*p))
	}
	if !slices.ContainsFunc(c, func(v int64) bool { return v != 0 }) {
		return nil
	}
	return c
}
