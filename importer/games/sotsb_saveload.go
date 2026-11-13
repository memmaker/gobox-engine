package games

// The saved game's layout: area, game state, aux, then the position and
// states, the wall slots and the party's file names.
const (
	sotsbSaveNameSize  = 41 // string[40]
	sotsbSaveNameSlots = 8
	sotsbSavePosition  = 5 // x, y, facing, the wall and the cell there
	saveWallSlots      = 3
	sotsbSaveGameSize  = 1 + 2048 + 2048 + 1024 + sotsbSavePosition + 2 + saveWallSlots*4 + 1 + sotsbSaveNameSize*sotsbSaveNameSlots
)
