package pool

import (
	"fmt"
	"os"
	"gobox/rules1e"
	"testing"

	"gobox/dax"
	"gobox/ecl/script"
)

// TestArea9CoinLoop: area 9's special cases type the coin loop (SelCoins[i] for
// SelName[187+i] and its high byte). Under the address VM the original byte
// code and the patched script take the same coins: TakeCoins runs from the
// same purse and cost in both, and the purse and the cost left are equal.
func TestArea9CoinLoop(t *testing.T) {
	const scripts, area = "../data/por/sources/assets/scripts", 9
	l, err := dax.NewLoader("../data/por/sources/dos")
	if err != nil {
		t.Skip(err)
	}
	names, err := script.ReadNames(os.DirFS(scripts), area, true, script.V1)
	if err != nil {
		t.Fatal(err)
	}
	var raw []byte
	for set := 1; set <= 8 && raw == nil; set++ {
		if b := l.BlockShared(fmt.Sprintf("ECL%d", set), area); len(b) >= 2 && set == names.Set {
			raw = b[2:]
		}
	}
	orig, err := script.Decode(raw, names)
	if err != nil {
		t.Fatal(err)
	}
	// The byte code indexes the coins from 6B01, a word no field holds (the
	// text's SelCoins[i] replaced it): this test's own field gives that base
	// its Ref and reads on past it into the coins as the original did
	// (coinWindow).
	orig = resolvedOver(t, orig, area, "6B01 1 byte member CoinLoopBase\n")
	dir := scripts // imported, special cases applied, by the TestMain import (por-go cmd/importstale)
	text, err := os.ReadFile(script.AreaPath(dir, area))
	if err != nil {
		t.Fatal(err)
	}
	patched, err := script.Parse(string(text), names)
	if err != nil {
		t.Fatal(err)
	}
	patched = resolved(t, patched, area)
	run := func(s *script.Script, p rules1e.Purse, cost uint16) string {
		e := quickStarted(t)
		at, ok := s.Labels["TakeCoins"]
		if !ok {
			t.Fatal("no TakeCoins")
		}
		s.Code = append(s.Code, script.Instr{Op: script.OpGosub, Args: []script.Arg{{Type: script.ArgVar, N: at, Label: true}}}, script.Instr{Op: script.OpExit})
		defer func() { s.Code = s.Code[:len(s.Code)-2] }()
		e.vm.Script = s
		e.engineFields["SelName"] = coinWindow(e, "SelName", 0x6B00)
		e.engineFields["CoinLoopBase"] = coinWindow(e, "CoinLoopBase", 0x6B01)
		clear(e.vm.Code)
		copy(e.vm.Code, s.Data)
		e.state().Purse = p
		scratch4 := e.field("Scratch4") // the cost left, in copper-worth units of the kind being taken
		e.host.Set(scratch4, int64(cost))
		e.vm.PC = uint16(len(s.Code) - 2)
		e.vm.Stop = false
		e.eclRun()
		return fmt.Sprint(e.state().Purse, e.host.Get(scratch4))
	}
	cases := 0
	for _, p := range []rules1e.Purse{{}, {Copper: 7}, {Gold: 3, Copper: 40}, {Gold: 12}, {Gold: 300, Copper: 999}, {Gold: 5000}} {
		for _, cost := range []uint16{0, 1, 3, 10, 60, 250} {
			if a, b := run(orig, p, cost), run(patched, p, cost); a != b {
				t.Errorf("purse %+v cost %d: original %s, patched %s", p, cost, a, b)
			}
			cases++
		}
	}
	t.Log(cases, "cases")
}

// coinWindow is field name at script address from indexed on into the
// selected member's coins at 6BBB (SelCoins), as the original's byte code
// read them from 6B00 (SelName) and 6B01; the other words past the field
// read 0.
func coinWindow(e *Engine, name string, from int) engineField {
	const first, n = 0x6BBB, 9
	coin := e.field("SelCoins")
	own := e.engineFields[name]
	f := own // CoinLoopBase is no field of the game's: none of its own words kept
	f.get = func(i int) (int64, bool) {
		switch a := from + i; {
		case a >= first && a < first+n:
			return e.host.Get(coin.Plus(a - first)), true
		case i > 0 || !own.stored:
			return 0, true
		case own.get != nil:
			return own.get(i)
		}
		return 0, false
	}
	f.set = func(i int, v int64) {
		if a := from + i; a >= first && a < first+n {
			e.host.Set(coin.Plus(a-first), v)
		} else if own.set != nil {
			own.set(i, v)
		}
	}
	return f
}
