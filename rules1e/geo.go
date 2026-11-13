package rules1e

import "gobox/core/grid"

// Area map (GEO) and the 3D view's state, ported from ovr30. The view
// itself is the remaster's (presentdata, gobox-engine/blobber): the
// original's wall drawing is gone, the engine only says where the party
// stands and looks (View.X, Y, Facing, Area).
//
// A GEO block is 2 header bytes and four 16x16 planes indexed y*16+x:
//
//	0x000: north wall (high nibble), east wall (low nibble)
//	0x100: south wall (high nibble), west wall (low nibble)
//	0x200: cell byte (ECL reads it as DS:6AB1)
//	0x300: door bits, 2 per direction: N 0..1, E 2..3, S 4..5, W 6..7
//
// A wall value 1..15 indexes the 15 wall definitions of the 3 WALLDEF
// slots; only the slots' block numbers are kept (the save file holds them).

// DirDX/DirDY = DS:2736 / DS:273F, indexed by direction 0..8 (0 N, 2 E, 4 S, 6 W).
var DirDX, DirDY = grid.StepX, grid.StepY

// OffMapSolid is the guard shared by the ovr30 accessors: off the map the
// answer is 0 (no wall, no way through). extra: in every area; the original
// did so only when DS:84DC was 0 or 0xA and wrapped the coordinates
// elsewhere (ovr30:037B), an oversight the port does not keep.
func (e *RuleState) OffMapSolid(x, y int) bool { return !grid.OnMap(x, y) }
