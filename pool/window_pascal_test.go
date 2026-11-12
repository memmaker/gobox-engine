package pool

import (
	"bufio"
	"fmt"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// TestWindowAgainstPascal compares the selected member's window writes
// (writeSpecialGlobal) with the Pascal's SpecialWriteZ1 (rules1e
// testdata/por-window.txt, por-re vecwin): each address the Pascal gives a
// meaning with a set of values, on a fresh member. The coins go to the party
// purse (the money unification) and are not here. The "rd" lines read each
// window address a script reads back (readSpecialGlobal against SpecialReadZ1).
func TestWindowAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-window.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, want, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		c := e.member(e.selected())
		if strings.HasPrefix(line, "rd ") {
			r := byte(k[0])
			c.Race, c.Stats[rules1e.StatCon], c.Saves[1], c.Control, c.Alignment, c.Okay, c.Move = rules1e.RaceID(r), rules1e.Score(10+r), rules1e.Score(5+r), r, r, r%2, 6+r
			for i := range c.Thief {
				c.Thief[i] = rules1e.Percent(21 + byte(i) + r)
			}
			got := "0" // no field there: no script reads it
			if r, ok := refAt(t, e, k[1]); ok {
				got = fmt.Sprint(e.host.Get(r))
			}
			if got != want {
				t.Errorf("%s: Go %s", line, got)
			}
			cases++
			continue
		}
		c.Memorized = [rules1e.MemorySlots]rules1e.SpellID{}
		c.Control, c.ExpValue, c.ExpPerHP, c.Okay, c.Status, c.Side, c.Auto = 0, 0, 0, 1, rules1e.StatusOkay, rules1e.SideParty, 0
		e.es.memberGone, e.es.memberDied = 0, 0
		if r, ok := refAt(t, e, k[0]); ok { // no field there: no script writes it, nothing changes
			e.host.Set(r, int64(uint16(k[1])))
		}
		var b strings.Builder
		for i := 1; i <= 20; i++ {
			fmt.Fprint(&b, c.Memorized[i].Byte(), " ")
		}
		fmt.Fprint(&b, c.Control, c.ExpValue, c.ExpPerHP, c.Okay, c.Status, c.Side, c.Auto, e.es.memberGone, e.es.memberDied)
		if got := b.String(); got != strings.Join(strings.Fields(want), " ") {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
