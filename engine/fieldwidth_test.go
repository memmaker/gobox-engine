package engine

import (
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/ecl"
)

// TestFieldWidths: a write keeps what the field's width holds: a dword past
// 0xFFFF; a byte field a whole word, as the original's word cells did
// (cutToWidth), wrapping at 0x10000. A save and load keeps it.
func TestFieldWidths(t *testing.T) {
	e := &Engine{}
	e.initVM()
	e.vars, e.locals = fieldCells{}, fieldCells{}
	e.vm.Script = &ecl.Script{Layout: &ecl.Layout{Fields: []ecl.Field{
		{Name: "TestDword", Owner: ecl.OwnerGame, Width: ecl.WidthDword, Len: 1},
		{Name: "TestByte", Owner: ecl.OwnerGame, Width: ecl.WidthByte, Len: 1},
		{Name: "TestWord", Owner: ecl.OwnerGame, Width: ecl.WidthWord, Len: 1},
	}}}
	dword, byt, word := ecl.Ref{Field: 1}, ecl.Ref{Field: 2}, ecl.Ref{Field: 3}
	e.setVar(dword, 0x12345)
	e.setVar(byt, 0x10101)
	e.setVar(word, 0x10002)
	path := filepath.Join(t.TempDir(), adventure.GameFile)
	if err := adventure.WriteSave(path, adventure.Save{Fields: e.vars, Locals: e.locals}); err != nil {
		t.Fatal(err)
	}
	g, err := adventure.ReadSave(path)
	if err != nil {
		t.Fatal(err)
	}
	e.vars = g.Fields
	for _, c := range []struct {
		r    ecl.Ref
		want int64
	}{{dword, 0x12345}, {byt, 0x101}, {word, 2}} {
		if got := e.getVar(c.r); got != c.want {
			t.Errorf("%s = %#x, want %#x", e.vm.Script.Layout.Field(c.r).Name, got, c.want)
		}
	}
}
