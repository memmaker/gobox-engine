package importpod

// Area is where the party is.
type Area struct {
	Digit     int    // game area: the n of GEO<n>.DAX, WALLDEF<n>.DAX ...
	Block     int    // the ECL block of the area
	GeoBlock  int    // GEO<Digit>.DAX block of the map
	Walls     [3]int // WALLDEF<Digit>.DAX block in each wall slot (-1 none)
	X, Y      int
	Facing    int // 0 N, 2 E, 4 S, 6 W
	Searching bool
	Night     bool // the 3D view's night sky (O12A:0285: skygrnd 2)
	Dungeon   bool // false: the wilderness, a shop, camp ...

	Minute, Hour, Day, Month, Year int
}

// Picture is a picture asset: DAX File, Block (-1: none) and the Frame
// within the block.
type Picture struct {
	File  string
	Block int
	Frame int
}
