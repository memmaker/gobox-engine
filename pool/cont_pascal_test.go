package pool

import (
	"bufio"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// TestContinueBattleAgainstPascal compares the round's end with Combat.EndOfRound
// (rules1e testdata/por-cont.txt, por-re veccont): the party standing or down, the
// foes standing or down, and the answer to Continue Battle: endRound's end and
// whether the fight asks, with the answer applied as dContinueFight does.
func TestContinueBattleAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-cont.txt")
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
		asks := over && e.state().Rv.Side[0] > 0 && e.state().Rv.Side[1] == 0 // fightEnd's Continue Battle: (combatflow.go)
		done := over && !(asks && answer)                                     // dContinueFight: Yes plays on (TestContinueBattle drives it)
		if done != want {
			t.Errorf("%s: Go done %v", line, done)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
