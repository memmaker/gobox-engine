package importpod

import (
	"gobox/char"
)

// The saved game's layout: the game block (p87F8), the position (x, y,
// facing, the wall and the cell there), the last and current game state,
// the wall set's two words, then the party's file names.
const (
	gameBlockSize     = 1024
	saveNameSize      = 41 // string[40]
	saveNameSlots     = 8
	savePosition      = 5
	saveGameSize      = gameBlockSize + savePosition + 2 + 2*2 + 1 + saveNameSize*saveNameSlots
	saveNamesAt       = saveGameSize - 1 - saveNameSize*saveNameSlots
	silverBladesParty = 5140 // where a Silver Blades saved game's party starts
)

// OriginalSave is savgamX.dat as above: no lead byte, the game block's bytes (script address =
// offset + 1, eclvars.go), no code image, the position and states, the wall set's two words.
var OriginalSave = char.SaveLayout{
	Records:   []char.SaveRecord{{Base: 1, Cells: gameBlockSize, CellBytes: 1}},
	CodeBase:  int(eclLayout.Code.First),
	CodeSize:  scriptCodeSize,
	WallBytes: 2 * 2,
}
