// Package grid is the combat map geometry the Pool games share: the line
// walker, line of sight, the quarter arcs, the near list order and the
// direction between two squares. Pure functions on squares; the map, its
// ground tables and the combatants stay with the game, which passes in what
// differs (map size, ground heights, slopes) as parameters.
package grid

// StepDirs is the direction taken by a step of (dx, dy) in -1..1, indexed
// [dy+1][dx+1]; 8 = none. Directions run 0 N, 1 NE, 2 E ... 7 NW.
var StepDirs = [3][3]byte{
	{7, 0, 1},
	{6, 8, 2},
	{5, 4, 3},
}

// StepX and StepY are the compass steps, indexed by direction 0..8 (8 = none).
var (
	StepX = [9]int{0, 1, 1, 1, 0, -1, -1, -1, 0}
	StepY = [9]int{-1, -1, 0, 1, 1, 1, 0, -1, 0}
)

// Line is the game's line walker record: NewLine starts it, Step takes one
// step. Err counts in half units of the major axis; Steps counts half
// squares, three for a diagonal.
type Line struct {
	X1, Y1, Err, Dx, Dy, X, Y, Sx, Sy, Steps int
	Dir                                      byte // the direction of the last step
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// NewLine starts a line from (x0, y0) to (x1, y1).
func NewLine(x0, y0, x1, y1 int) *Line {
	return &Line{X1: x1, Y1: y1, Dx: abs(x1 - x0), Dy: abs(y1 - y0), Sx: sign(x1 - x0), Sy: sign(y1 - y0), X: x0, Y: y0}
}

// Step goes one square along the major axis, and along the minor one when
// the error reaches the major distance; Dir is the direction taken.
func (l *Line) Step() bool {
	mx, my, moved := 0, 0, false
	if l.Dx >= l.Dy {
		if l.X != l.X1 {
			l.X += l.Sx
			mx = l.Sx
			l.Err += 2 * l.Dy
			l.Steps += 2
			if l.Err >= l.Dx {
				l.Err -= 2 * l.Dx
				l.Y += l.Sy
				my = l.Sy
				l.Steps++
			}
			moved = true
		}
	} else if l.Y != l.Y1 {
		l.Y += l.Sy
		my = l.Sy
		l.Err += 2 * l.Dx
		l.Steps += 2
		if l.Err >= l.Dy {
			l.Err -= 2 * l.Dy
			l.X += l.Sx
			mx = l.Sx
			l.Steps++
		}
		moved = true
	}
	l.Dir = StepDirs[my+1][mx+1]
	return moved
}

// Sight is what line of sight reads of the map.
type Sight struct {
	SeeAll bool                             // walls do not block
	Height func(x, y int) (view, block int) // the ground's view height and blocking height at a square
}

// LineOfSight walks from (x,y) towards (*tx,*ty) until the target, a wall
// higher than the line or the range (in half squares) stops it. On a stop
// short of the target (*tx,*ty) is the square it reached; *rng is the half
// squares walked.
func LineOfSight(s Sight, rng *int, tx, ty *int, x, y int) bool {
	limit := *rng*2 + 1
	a := NewLine(x, y, *tx, *ty)
	h, _ := s.Height(x, y)
	b := NewLine(0, h, max(a.Dx, a.Dy), h)
	for {
		_, block := s.Height(a.X, a.Y)
		if (!s.SeeAll && block > b.Y) || a.Steps&0xFF > limit {
			*tx, *ty = a.X, a.Y
			*rng = a.Steps & 0xFF
			return false
		}
		b.Step()
		if !a.Step() {
			*rng = a.Steps & 0xFF
			return true
		}
	}
}

// InArc reports whether (x2,y2) lies in the quarter facing dir from
// (x1,y1) (dir 8 or 0xFF: anywhere) on a map of w x h squares; a square off
// the map is in no arc.
func InArc(w, h, x1, y1, x2, y2, dir int) bool {
	if x1 < 0 || x1 >= w || y1 < 0 || y1 >= h || x2 < 0 || x2 >= w || y2 < 0 || y2 >= h {
		return false
	}
	if dir == 0xFF {
		dir = 8
	}
	nx, ny := x1+StepX[dir], y1+StepY[dir]
	if (x1 == x2 && y1 == y2) || (nx == x2 && ny == y2) {
		return true
	}
	switch dir {
	case 0:
		return (x2 >= nx && y2 <= ny+nx-x2) || (x2 <= nx && y2 <= ny+x2-nx)
	case 1:
		return (x2 >= nx && y2 <= ny+nx-x2) || (x2 >= nx+ny-y2 && y2 <= ny)
	case 2:
		return (x2 >= nx+ny-y2 && y2 <= ny) || (x2 >= nx+y2-ny && y2 >= ny)
	case 3:
		return (x2 >= nx+y2-ny && y2 >= ny) || (x2 >= nx && y2 >= ny+x2-nx)
	case 4:
		return (x2 >= nx && y2 >= ny+x2-nx) || (x2 <= nx && y2 >= ny+nx-x2)
	case 5:
		return (x2 <= nx && y2 >= ny+nx-x2) || (x2 <= nx+ny-y2 && y2 >= ny)
	case 6:
		return (x2 <= nx+ny-y2 && y2 >= ny) || (x2 <= nx+y2-ny && y2 <= ny)
	case 7:
		return (x2 <= nx+y2-ny && y2 <= ny) || (x2 <= nx && y2 <= ny+x2-nx)
	}
	return true
}

// NearSort orders a near list by distance. On a tie the games differ: POR
// and COTAB put the lower direction first, SOTSB and POD the higher
// (lowDirFirst). key gives an entry's distance and direction.
func NearSort[T any](l []T, key func(T) (dist, dir byte), lowDirFirst bool) {
	for i := 0; i < len(l)-1; i++ {
		for j := i + 1; j < len(l); j++ {
			a, b := l[i], l[j]
			ad, adir := key(a)
			bd, bdir := key(b)
			swap := bd < ad
			if bd == ad {
				if lowDirFirst {
					swap = bdir < adir && bdir%2 <= adir%2
				} else {
					swap = adir < bdir && adir%2 <= bdir%2
				}
			}
			if swap {
				l[i], l[j] = b, a
			}
		}
	}
}

// FormationEdge reports a square outside a w x h formation on both axes.
func FormationEdge(x, y, w, h int) bool {
	return (x < 0 || x >= w) && (y < 0 || y >= h)
}

// DirectionTo is the direction 0..7 from (x1,y1) to (x2,y2): the plane is
// cut into eight sectors by two slopes (steep and shallow, in 1/scale). On a
// square level with the origin on one axis the games differ: SOTSB and POD
// leave it to the straight sectors (strict), COTAB gives it to the diagonal.
func DirectionTo(x1, y1, x2, y2, steepSlope, shallowSlope, scale int, strict bool) int {
	dx, dy := abs(x2-x1), abs(y2-y1)
	steep := steepSlope * dx / scale
	shallow := shallowSlope * dx / scale
	diagonal := steep >= dy && shallow <= dy
	before := func(a, b int) bool { return a < b || !strict && a == b }
	for dir := 0; dir < 8; dir++ {
		var fits bool
		switch dir {
		case 0:
			fits = y2 <= y1 && steep <= dy
		case 1:
			fits = before(y2, y1) && before(x1, x2) && diagonal
		case 2:
			fits = x2 >= x1 && shallow >= dy
		case 3:
			fits = before(y1, y2) && before(x1, x2) && diagonal
		case 4:
			fits = y2 >= y1 && steep <= dy
		case 5:
			fits = before(y1, y2) && before(x2, x1) && diagonal
		case 6:
			fits = x2 <= x1 && shallow >= dy
		case 7:
			fits = before(y2, y1) && before(x2, x1) && diagonal
		}
		if fits {
			return dir
		}
	}
	return 8 // never: the eight sectors cover the plane
}
