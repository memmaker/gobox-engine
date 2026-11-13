package importsotsb

import "gobox/rules1e"

// SotSB's combat ground constants (START.EXE).
const (
	cellRubble       = 0x40 // the map cell's flag for rubble in the room
	groundRubble     = 26
	groundRubbleEdge = 27
	tileOpenGround   = 22 // the groundInfo tile of open ground (things may be put there)
)

// CombatGround is how SotSB lays its combat ground (rules.json combatGround):
// the dungeon only, outdoors the map stays bare.
func CombatGround() rules1e.GroundGen {
	return rules1e.GroundGen{
		Corners: rules1e.CornerTiles{NWDoorOpen: 0xD, NEWallOpen: 7, NEDoorOpen: 1},
		Rubble: &rules1e.RubbleGen{CellFlag: cellRubble, Tile: tileOpenGround,
			Heap: groundRubble, Edge: groundRubbleEdge, HeapOdds: 5, EdgeOdds: 9},
	}
}
