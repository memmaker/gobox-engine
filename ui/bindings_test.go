package ui

import (
	"testing"
)

func TestTable(t *testing.T) {
	tb := &Table{Order: []string{"A", "C"}, Combat: "C", Rows: []Binding{
		{Ctx: "A", Keys: "Up", Desc: "up", Pad: "dpad", Codes: []Code{{0x4800, Up}}},
		{Ctx: "C", Keys: "Up", Desc: "cursor", Codes: []Code{{0x4800, CursorUp}}},
	}}
	if tb.Act(0x4800) != Up || tb.CombatAct(0x4800) != CursorUp || tb.Act('x') != None {
		t.Fatal("lookup")
	}
	if tb.KeyOf(Up) != 0x4800 || tb.KeyOf(CursorUp) != 0 {
		t.Fatal("KeyOf")
	}
	if h := tb.HelpLines(); len(h) != 9 || h[0][0] != "A" || h[1] != [2]string{"Up", "up"} {
		t.Fatalf("%q", h)
	}
	if Keypad('1') != End || Keypad('5') != None || Keypad('a') != None {
		t.Fatal("Keypad")
	}
}
