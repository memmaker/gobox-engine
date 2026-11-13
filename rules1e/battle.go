package rules1e

// GroundSquare is ground g as published: Wall is the blocking flag the
// movement rule tests (0xFF), Tile and Height its picture and height.
func (e *RuleState) GroundSquare(g byte) Square {
	r := e.GroundInfo(Ground(g))
	return Square{Tile: int(r[3]), Wall: r[0] == 0xFF, Height: int(r[2])}
}
