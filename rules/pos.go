package rules

// MapPos is a square of an area map: X east, Y south, from the map's top left.
type MapPos struct{ X, Y int }

// BoardPos is a square of the combat board: X east, Y south, from the board's
// top left. A board square is never a map square: the two are separate types.
type BoardPos struct{ X, Y int }

// Add, Sub and Mul are vector arithmetic on board squares (steps, offsets).
func (p BoardPos) Add(q BoardPos) BoardPos { return BoardPos{p.X + q.X, p.Y + q.Y} }
func (p BoardPos) Sub(q BoardPos) BoardPos { return BoardPos{p.X - q.X, p.Y - q.Y} }
func (p BoardPos) Mul(k int) BoardPos      { return BoardPos{p.X * k, p.Y * k} }

// Size is a width and height in squares.
type Size struct{ W, H int }
