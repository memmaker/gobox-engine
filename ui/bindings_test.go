package ui

import (
	"testing"

	"gobox/input"
)

func TestTable(t *testing.T) {
	tb := &Table{Order: []string{"A", "C"}, Combat: "C", Rows: []Binding{
		{Ctx: "A", Keys: "Up", Desc: "up", Pad: "dpad", Codes: []Code{{0x4800, input.Up}}},
		{Ctx: "C", Keys: "Up", Desc: "cursor", Codes: []Code{{0x4800, input.CursorUp}}},
	}}
	if tb.Act(0x4800) != input.Up || tb.CombatAct(0x4800) != input.CursorUp || tb.Act('x') != input.None {
		t.Fatal("lookup")
	}
	if tb.KeyOf(input.Up) != 0x4800 || tb.KeyOf(input.CursorUp) != 0 {
		t.Fatal("KeyOf")
	}
	if h := tb.HelpLines(); len(h) != 9 || h[0][0] != "A" || h[1] != [2]string{"Up", "up"} {
		t.Fatalf("%q", h)
	}
	if Keypad('1') != input.End || Keypad('5') != input.None || Keypad('a') != input.None {
		t.Fatal("Keypad")
	}
}
