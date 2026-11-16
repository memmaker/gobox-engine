package rules

// The straight board (Board.Straight, extra: the options' Deskew): the original's dungeon arena draws each row of
// map squares 5 columns west of the one above, a square a 6x5 block; straight, a square is a block of StraightCell
// columns and rows, its west wall the first column, its north wall the first row, its floor 4x4 as the original's;
// the party's square starts at column StraightParty, so the board holds 10 squares by 5. The arena's geometry is
// the Gold Box engine's, shared by the Pool games and Buck Rogers.
const (
	StraightCell  = 5
	StraightParty = 20
)

// StraightX is the column on the straight board of the arena square (x, y) of the original's skewed dungeon map:
// the skew taken out (a square west for each row down), then the original's 6-wide block made 5: its two wall
// columns are the one. Floor squares keep their own square; two squares of a wall line open to the west share one.
func StraightX(x, y int) int {
	sx := x - (y - 0xA) - 0x15 + 60 // unskewed, from the block 10 cells west of the party's
	return StraightParty + (sx/6-10)*StraightCell + max(sx%6-1, 0)
}
