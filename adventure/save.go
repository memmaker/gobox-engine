package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gobox/rules"
)

// The canonical save (section 6 of gobox ARCHITECTURE.md): the engine's part
// of a slot, written as JSON. The script's state travels by field name
// (layout/), the engine's words by their own names; no address and no record
// of the original's reaches the file. The members are their own files. The
// engine writes it; the importer writes it once from each original save.

// GameFile is the canonical save's file in a slot.
const GameFile = "GAME.JSON"

// SavesDir is the adventure folder's shipped characters and saves (the
// player's save folder is seeded from it); its SlotsDir/<name> are the slots.
const SavesDir, SlotsDir = "saves", "saves"

// Save is the canonical save.
type Save struct {
	Set                      rules.Set
	Rules                    json.RawMessage    // the ruleset's part (the ruleset's Save)
	Fields                   map[string][]int64 // the game's fields, by name
	Locals                   map[string][]int64 // the running area's locals, by name
	X, Y, Facing, Pos3, Pos4 byte
	ViewMode, BeginFlag      byte
	Party                    []string // the members' file bases (CHRDATA1..)
}

// ReadSave reads the canonical save at path.
func ReadSave(path string) (Save, error) {
	var g Save
	b, err := os.ReadFile(path)
	if err != nil {
		return g, err
	}
	if err := json.Unmarshal(b, &g); err != nil {
		return g, fmt.Errorf("%s: %w", path, err)
	}
	return g, nil
}

// MarshalSave is g as its file holds it.
func MarshalSave(g Save) ([]byte, error) { return json.MarshalIndent(g, "", "\t") }

// WriteSave writes g to path.
func WriteSave(path string, g Save) error {
	b, err := MarshalSave(g)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// MakeFileName is the DOS file name of a character name: the separators are
// removed, at most 8 characters are kept, upper case.
func MakeFileName(s string) string {
	for _, c := range " .*,?/\\:;|" {
		s = rules.RemoveSubstr(s, string(c))
	}
	if len(s) > 8 {
		s = s[:8]
	}
	return strings.ToUpper(s)
}
