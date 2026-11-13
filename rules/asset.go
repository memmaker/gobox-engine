package rules

import "image"

// Asset ids are the adventure's own names. Decoded data only: no file, block
// or format crosses to the engine.
type (
	PictureID string // "head2.3", per ARCHITECTURE.md's asset plan
	MapID     string // "geo3.1"
)

// Picture is one still or animation in the canonical format.
type Picture struct {
	Frames []image.NRGBA
	Delay  []int32 // milliseconds per frame; nil for a still
}

// Sound is PCM, mono, 16 bit.
type Sound struct {
	Rate    int32
	Samples []int16
}

// Edge is one wall side (por geo.go).
type Edge struct {
	Wall   int32 // a wall set's picture; 0 none
	Door   bool
	Locked bool
}

// Walls is a walking map, row by row: four edges a square.
type Walls struct {
	Size  image.Point
	Edges [][4]Edge
}
