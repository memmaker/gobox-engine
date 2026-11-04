package ui

import (
	"testing"

	"gobox/gpu"
)

// A border with no pieces (the game's symbols not loaded yet) draws no
// ribbon rather than crashing.
func TestWindowNoPieces(t *testing.T) {
	dst := gpu.NewImage(64, 64)
	if r := Window(dst, Ribbon{}, gpu.GeoM{}, 8, 8, nil); r.Dx() != 48 {
		t.Fatalf("inside %v", r)
	}
}
