package script

import (
	"strings"
	"testing"
)

// TestGoldCoins: a platinum operand names the gold cell and the command's
// literals grow by its value; in place when the field has no gold cell; a
// computed operand is kept and noted; "platinum > 0" is "gold > 4"; an
// equality is noted.
func TestGoldCoins(t *testing.T) {
	l, err := ParseLayout(strings.NewReader("0100 9 word member Coins coins=c,-,10c,-,100c,-,g,-,5g\n0200 2 word member Plat coins=5g,-\n0300 1 word local X\n"))
	if err != nil {
		t.Fatal(err)
	}
	v := func(a int) Arg { return Arg{Type: ArgVarWord, N: a} }
	s := &Script{Code: []Instr{
		{Op: OpCompare, Args: []Arg{{Type: ArgByte, N: 60}, v(0x108)}},      // platinum >= 60
		{Op: OpSub, Args: []Arg{v(0x108), {Type: ArgByte, N: 1}, v(0x108)}}, // take 1 platinum
		{Op: OpAdd, Args: []Arg{v(0x200), {Type: ArgWord, N: 5}, v(0x200)}}, // in place
		{Op: OpSave, Args: []Arg{{Type: ArgByte, N: 3}, v(0x102)}},          // 3 silver
		{Op: OpSave, Args: []Arg{v(0x108), v(0x300)}},                       // computed
		{Op: OpCompare, Args: []Arg{{Type: ArgByte, N: 7}, v(0x106)}},       // gold: untouched
	}}
	GoldCoins(s, l)
	want := []Instr{
		{Op: OpCompare, Args: []Arg{{Type: ArgWord, N: 300}, v(0x106)}},
		{Op: OpSub, Args: []Arg{v(0x106), {Type: ArgByte, N: 5}, v(0x106)}},
		{Op: OpAdd, Args: []Arg{v(0x200), {Type: ArgWord, N: 25}, v(0x200)}},
		{Op: OpSave, Args: []Arg{{Type: ArgByte, N: 30}, v(0x100)}},
		{Op: OpSave, Args: []Arg{v(0x106), v(0x300)}},
		{Op: OpCompare, Args: []Arg{{Type: ArgByte, N: 7}, v(0x106)}},
		{Op: OpCompare, Args: []Arg{v(0x106), {Type: ArgByte, N: 4}}},
		{Op: OpIfGt},
		{Op: OpGoto, Args: []Arg{{Type: ArgVar, N: 0, Label: true}}},
		{Op: OpCompare, Args: []Arg{v(0x106), {Type: ArgByte, N: 10}}},
		{Op: OpIfEq},
	}
	for i, in := range s.Code {
		w := want[i]
		for k := range w.Args {
			if in.Args[k] != w.Args[k] {
				t.Errorf("command %d operand %d: %+v, want %+v", i, k, in.Args[k], w.Args[k])
			}
		}
		if noted := in.Comment != ""; noted != (i == 0 || i == 4 || i == 9) {
			t.Errorf("command %d comment %q", i, in.Comment)
		}
	}
	if f := l.Fields[0]; !f.Moved(8) || f.Moved(6) || l.Fields[1].Moved(0) {
		t.Error("Moved")
	}
}
