package engine

import (
	"bufio"
	"gobox/rules1e"
	"os"
	"strings"
	"testing"
)

// TestContinueBattleAgainstPascal compares the round's end with Combat.EndOfRound
// (rules1e testdata/por-cont.txt, por-re veccont): the party standing or down, the
// foes standing or down: endRound's end. Continue Battle: is cut (the won
// fight ends, CHANGES-FROM-ORIGINAL): the Pascal's rows answered Yes to it
// play on there and are left out.
func TestContinueBattleAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-cont.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		want := atois(t, strings.Fields(val))[0] == 1
		party, foe, answer := k[0] == 1, k[1] == 1, k[2] == 1
		e := devFirstTurn(t, "meleefight", false)
		for i := 1; i <= e.state().Ncomb; i++ {
			if c := e.state().CombatantSlots[i-1]; c != nil && (c.Side != 0 && !foe || c.Side == 0 && !party) {
				c.Okay, c.Status = 0, rules1e.StatusDead
			}
		}
		over := e.endRound()
		if asked := over && e.state().Rv.Side[0] > 0 && e.state().Rv.Side[1] == 0; asked && answer { // the original's Yes to Continue Battle: played on
			continue
		}
		if done := over; done != want {
			t.Errorf("%s: Go done %v", line, done)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
