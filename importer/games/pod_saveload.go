package games

import (
	"gobox/rules1e"
)

// The saved game's layout: the game block (p87F8), the position (x, y,
// facing, the wall and the cell there), the last and current game state,
// the wall set's two words, then the party's file names.
const (
	gameBlockSize     = 1024
	podSaveNameSize   = 41 // string[40]
	podSaveNameSlots  = 8
	podSavePosition   = 5
	podSaveGameSize   = gameBlockSize + podSavePosition + 2 + 2*2 + 1 + podSaveNameSize*podSaveNameSlots
	saveNamesAt       = podSaveGameSize - 1 - podSaveNameSize*podSaveNameSlots
	silverBladesParty = 5140 // where a Silver Blades saved game's party starts
)

// OriginalSave is savgamX.dat as above: no lead byte, the game block's bytes (script address =
// offset + 1, eclvars.go), no code image, the position and states, the wall set's two words.
var PodOriginalSave = rules1e.SaveLayout{
	Records:   []rules1e.SaveRecord{{Base: 1, Cells: gameBlockSize, CellBytes: 1}},
	CodeBase:  int(podEclLayout.Code.First),
	CodeSize:  podScriptCodeSize,
	WallBytes: 2 * 2,
}
