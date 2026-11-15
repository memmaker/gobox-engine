package engine

import (
	"testing"

	"gobox/ecl"
)

// storeEngine is an engine over a layout of a byte, a word, a long and a 8-byte buffer,
// running dialect v.
func storeEngine(v *ecl.Version) *Engine {
	e := &Engine{}
	e.initVM()
	e.eclVersion = v
	e.vars, e.locals = fieldCells{}, fieldCells{}
	e.vm.Script = &ecl.Script{Layout: &ecl.Layout{Fields: []ecl.Field{
		{Name: "B", Owner: ecl.OwnerGame, Width: ecl.WidthByte, Len: 1},
		{Name: "W", Owner: ecl.OwnerGame, Width: ecl.WidthWord, Len: 1},
		{Name: "L", Owner: ecl.OwnerGame, Width: ecl.WidthDword, Len: 1},
		{Name: "Buf", Owner: ecl.OwnerLocal, Width: ecl.WidthBytes, Len: 8},
	}}}
	return e
}

func as(f ecl.FieldID, i int, w ecl.Width) ecl.Ref { return ecl.Ref{Field: f, Index: i, As: w} }

func TestStoreWidthByOperand(t *testing.T) {
	e := storeEngine(ecl.V4)
	b, w, l := as(1, 0, ecl.WidthByte), as(2, 0, ecl.WidthWord), as(3, 0, ecl.WidthDword)
	e.setVar(b, 256)
	if got := e.getVar(b); got != 0 {
		t.Errorf("byte 256 = %d", got)
	}
	e.setVar(b, e.getVar(b)-1)
	if got := e.getVar(b); got != 255 {
		t.Errorf("byte 0-1 = %d", got)
	}
	e.setVar(w, 0x12345)
	if got := e.getVar(w); got != 0x2345 {
		t.Errorf("word = %#x", got)
	}
	e.setVar(l, -5)
	if got := e.getVar(l); got != -5 {
		t.Errorf("long = %d", got)
	}
}

func TestStoreWidthBytes(t *testing.T) {
	e := storeEngine(ecl.V4)
	e.setVar(as(4, 0, ecl.WidthWord), 0x1234)
	if got := e.getVar(as(4, 0, ecl.WidthByte)); got != 0x34 {
		t.Errorf("byte 0 = %#x", got)
	}
	if got := e.getVar(as(4, 1, ecl.WidthByte)); got != 0x12 {
		t.Errorf("byte 1 = %#x", got)
	}
	e.setVar(as(4, 1, ecl.WidthByte), 0xFF)
	if got := e.getVar(as(4, 0, ecl.WidthWord)); got != 0xFF34 {
		t.Errorf("word = %#x", got)
	}
	e.setVar(as(4, 2, ecl.WidthDword), -7)
	if got := e.getVar(as(4, 2, ecl.WidthDword)); got != -7 {
		t.Errorf("long = %d", got)
	}
	for _, r := range []ecl.Ref{as(4, 5, ecl.WidthDword), as(4, 8, ecl.WidthByte)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%+v past the buffer did not fail", r)
				}
			}()
			e.getVar(r)
		}()
	}
}

// Pool: a cell is a word whatever the operand; a byte buffer is refused.
func TestStoreWidthPool(t *testing.T) {
	e := storeEngine(ecl.V1)
	e.setVar(as(1, 0, ecl.WidthByte), 0x10101)
	if got := e.getVar(as(1, 0, ecl.WidthByte)); got != 0x101 {
		t.Errorf("byte = %#x", got)
	}
	e.setVar(as(3, 0, ecl.WidthDword), 0x12345)
	if got := e.getVar(as(3, 0, ecl.WidthDword)); got != 0x12345 {
		t.Errorf("long = %#x", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("a byte buffer in a Pool dialect was not refused")
		}
	}()
	e.getVar(as(4, 0, ecl.WidthByte))
}
