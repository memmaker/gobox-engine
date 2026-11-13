package importsotsb

// The saved game's layout: area, game state, aux, then the position and
// states, the wall slots and the party's file names.
const (
	saveNameSize  = 41 // string[40]
	saveNameSlots = 8
	savePosition  = 5 // x, y, facing, the wall and the cell there
	saveWallSlots = 3
	saveGameSize  = 1 + 2048 + 2048 + 1024 + savePosition + 2 + saveWallSlots*4 + 1 + saveNameSize*saveNameSlots
)
