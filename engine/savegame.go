package engine

import (
	"encoding/json"
	"fmt"
	"gobox/core/goldbox"
	"os"
)

// The canonical save (section 6 of gobox ARCHITECTURE.md): the engine's
// part of a slot, written as JSON. The script's state travels by field name
// (layout/), the engine's words by their own names; no address and no
// record of the original's reaches the file. The members are their own
// files (SaveCharacter). An original SAVGAMA.DAT is read once, by the import
// (saveimport.go), and nothing writes one.

// gameFile is the canonical save's file in a slot.
const gameFile = "GAME.JSON"

// gameSave is the canonical save.
type gameSave struct {
	Set                      goldbox.Set
	Rules                    json.RawMessage // the ruleset's part (the ruleset's Save)
	Fields                   fieldCells      // the game's fields, by name
	Locals                   fieldCells      // the running area's locals, by name
	X, Y, Facing, Pos3, Pos4 byte
	ViewMode, BeginFlag      byte
	Party                    []string // the members' file bases (CHRDATA1..)
}

// legacySave is a save written before the ruleset's part was its own blob:
// Area, Game and Purse at the top level. readGameSave moves them into Rules.
type legacySave struct{ Area, Game, Purse json.RawMessage }

// readGameSave reads the canonical save at path.
func readGameSave(path string) (gameSave, error) {
	var g gameSave
	b, err := os.ReadFile(path)
	if err != nil {
		return g, err
	}
	if err := json.Unmarshal(b, &g); err != nil {
		return g, fmt.Errorf("%s: %w", path, err)
	}
	if g.Rules == nil {
		var l legacySave
		if err := json.Unmarshal(b, &l); err != nil {
			return g, fmt.Errorf("%s: %w", path, err)
		}
		if g.Rules, err = json.Marshal(l); err != nil {
			return g, err
		}
	}
	return g, nil
}

// writeGameSave writes g to path.
func writeGameSave(path string, g gameSave) error {
	b, err := marshalGameSave(g)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func marshalGameSave(g gameSave) ([]byte, error) { return json.MarshalIndent(g, "", "\t") }
