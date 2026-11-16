package ecl

import "testing"

// TestLayoutFlat: a table index past a field's end continues into the field that holds the address, as the original's flat memory does.
func TestLayoutFlat(t *testing.T) {
	l := &Layout{Fields: []Field{
		{Name: "Base", Owner: OwnerGame, Width: WidthByte, From: 0x4B00, Len: 1},
		{Name: "Next", Owner: OwnerGame, Width: WidthByte, From: 0x4B01, Len: 3},
		{Name: "Table", Owner: OwnerConst, Width: WidthByte, From: 0x4B10, Len: 2},
	}}
	base := Ref{Field: 1, As: WidthWord}
	for _, c := range []struct {
		in   Ref
		want Ref
	}{
		{base, base}, // inside its field
		{base.Plus(2), Ref{Field: 2, Index: 1, As: WidthWord}}, // 4B02 is Next[1]
		{base.Plus(4), base.Plus(4)},                           // no field holds 4B04: unchanged, the engine says the layout is short
		{Ref{Field: 3}.Plus(5), Ref{Field: 3}.Plus(5)},         // a constant table of the code does not continue
	} {
		if got := l.Flat(c.in); got != c.want {
			t.Errorf("Flat(%+v) = %+v, want %+v", c.in, got, c.want)
		}
	}
}
