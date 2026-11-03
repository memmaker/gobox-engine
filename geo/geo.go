// Package geo parses GEO and WALLDEF block bytes (already decoded by the
// caller's DAX reader) into plain values. No engine state, no rendering.
package geo

import "errors"

const (
	Size        = 0x402    // GEO block: 2-byte header + 4 planes
	WallSize    = 0x9C     // one wall's symbol ids
	SetSize     = 5 * 0x9C // 0x30C: one WALLDEF chunk (5 walls)
	SymWallBase = 0x2D     // symbol ids >= this are relocated per slot
	Slots       = 3        // wall slots 1..3, 5 walls each (wall values 1..15)
)

// Map is a 16x16 area map; planes are indexed y*16+x.
type Map struct {
	Plane [0x400]byte // 0 N/E walls, 0x100 S/W walls, 0x200 cells, 0x300 doors
	// SolidEdges: off-map reads answer 0. When false they wrap (>15 -> 0,
	// <0 -> 15), like cotab/sotsb outside their solid-edge areas. por: always true.
	SolidEdges bool
}

var ErrSize = errors.New("geo: bad GEO block size")

// ParseMap reads a GEO block; bytes past Size are ignored (the games report
// a long block themselves). SolidEdges is left false for the caller to set.
func ParseMap(b []byte) (*Map, error) {
	if len(b) < Size {
		return nil, ErrSize
	}
	m := &Map{}
	copy(m.Plane[:], b[2:])
	return m, nil
}

// OnMap reports whether x,y is inside the 16x16 map.
func OnMap(x, y int) bool { return x >= 0 && x <= 15 && y >= 0 && y <= 15 }

func wrap(v int) int {
	if v > 15 {
		return 0
	}
	if v < 0 {
		return 15
	}
	return v
}

// offSolid: off the map and the edges are solid.
func (m *Map) offSolid(x, y int) bool { return !OnMap(x, y) && m.SolidEdges }

// WallAt is the wall value 0..15 on side dir (0 N, 2 E, 4 S, 6 W) of a cell.
func (m *Map) WallAt(x, y, dir int) byte {
	if m.offSolid(x, y) {
		return 0
	}
	i := wrap(y)<<4 + wrap(x)
	switch dir {
	case 0:
		return m.Plane[i] >> 4
	case 2:
		return m.Plane[i] & 0xF
	case 4:
		return m.Plane[0x100+i] >> 4
	case 6:
		return m.Plane[0x100+i] & 0xF
	}
	return 0
}

// DoorAt is the 2-bit door state on that side (0 wall, 1 door/open, 2/3 locked);
// 1 for a side with no wall, 0 off a solid edge.
func (m *Map) DoorAt(x, y, dir int) byte {
	if m.offSolid(x, y) {
		return 0
	}
	if m.WallAt(x, y, dir) == 0 {
		return 1
	}
	b := m.Plane[0x300+wrap(y)<<4+wrap(x)]
	if dir < 0 || dir > 6 || dir&1 != 0 {
		return 1
	}
	return b >> uint(dir) & 3
}

// MapCell is the cell byte (plane 2); 0 off a solid edge.
func (m *Map) MapCell(x, y int) byte {
	if m.offSolid(x, y) {
		return 0
	}
	return m.Plane[0x200+wrap(y)<<4+wrap(x)]
}

// Walldef is a WALLDEF block split into 5-wall chunks of SetSize bytes.
// A trailing partial chunk is dropped: the decided behaviour (DESIGN.md, Rule 2 decisions); the original
// copies it too.
type Walldef struct{ Sets [][]byte }

// ParseWalldef splits a block; an empty or short block is an error the
// caller ignores (the original stops the game there).
func ParseWalldef(b []byte) (*Walldef, error) {
	n := len(b) / SetSize // shorter than one chunk: no sets
	w := &Walldef{}
	for k := 0; k < n; k++ {
		w.Sets = append(w.Sets, append([]byte(nil), b[k*SetSize:(k+1)*SetSize]...))
	}
	return w, nil
}

// Fits reports whether the chunks fit in slots slot..3.
func (w *Walldef) Fits(slot int) bool { return ValidWallSlot(slot) && slot+len(w.Sets) <= 4 }

// Relocate adds delta to every symbol id >= SymWallBase of chunk k
// The game computes delta from its own per-slot symbol base table (data it owns); geo holds no table.
func (w *Walldef) Relocate(k int, delta byte) {
	for i, v := range w.Sets[k] {
		if v >= SymWallBase {
			w.Sets[k][i] = v + delta
		}
	}
}

// Wall returns wall i (0..4) of chunk k as WallSize symbol ids
// (wall value v: chunk (v-1)/5 of the slot set, index (v-1)%5).
func (w *Walldef) Wall(k, i int) []byte { return w.Sets[k][i*WallSize : (i+1)*WallSize] }

// ValidWallSlot is the set {1,2,3}.
func ValidWallSlot(s int) bool { return s >= 1 && s <= 3 }

// SymbolBlock is the 8X8D block for chunk k (1-based) of an n-chunk WALLDEF
// block: block*10+k (as a byte) when n>1, else block.
func SymbolBlock(block, k, n int) int {
	if n > 1 {
		return int(byte(block*10 + k))
	}
	return block
}
