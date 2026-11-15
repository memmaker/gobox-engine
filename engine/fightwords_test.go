package engine

import (
	"fmt"
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// fightScene starts a plain fight of three foes (the dev scene "combat") after set ran, and plays to the
// party's first turn.
func fightScene(t *testing.T, set func(e *Engine)) *Engine {
	t.Helper()
	e := testEngine(t)
	e.Explore = true
	e.DevAction = func() {
		idx := 1
		if first := e.ruleset().Check(rules.FirstMonster{}, nil); first.Yes {
			idx = int(first.Value)
		}
		e.addMonsters(idx, 3, 1)
		set(e)
		e.push(cont{kind: kDevFought})
		e.pushCombat()
	}
	seed := devSeed
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	toDev(e)
	for range 400 {
		if e.cs.turn != nil && e.cs.actor != rules.NoChar {
			return e
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	t.Fatal("no party turn")
	return e
}

func combatSquares(e *Engine) []string {
	var out []string
	for _, c := range e.state().CombatantSlots {
		if c != nil {
			out = append(out, fmt.Sprintf("%d:%d,%d", c.Side, e.state().CombatX(c), e.state().CombatY(c)))
		}
	}
	return out
}

// The values below were recorded on the code before the words moved out of the
// ruleset (the same seed, devSeed).
var monsterSquaresByDist = map[uint16][]string{ // the three foes' squares by the encounter distance
	0: {"1:26,12", "1:27,13", "1:25,12"},
	1: {"1:20,12", "1:21,13", "1:19,12"},
	2: {"1:14,12", "1:15,13", "1:13,12"},
	3: {"1:8,12", "1:9,13", "1:7,12"},
}

// TestFightWordsMonsterDist: the monsters stand MonsterDist squares off, as the ruleset placed them
// when it read the word itself.
func TestFightWordsMonsterDist(t *testing.T) {
	for d, want := range monsterSquaresByDist {
		e := fightScene(t, func(e *Engine) { e.world.MonsterDist = d })
		got := combatSquares(e)
		if len(got) < len(want) || !slices.Equal(got[len(got)-len(want):], want) {
			t.Errorf("distance %d: squares %v, want the last three %v", d, got, want)
		}
	}
}

// TestFightWordsMoraleAndSurprise: a Morale above 100 reads 100 once the fight began, a Surprise
// bit reads 0 once the first round began.
func TestFightWordsMoraleAndSurprise(t *testing.T) {
	e := fightScene(t, func(e *Engine) {
		e.world.Morale, e.world.Surprise = 150, 3
		if got := e.fightTerms().Surprised; !slices.Equal(got, []rules.Side{rules.SideParty, rules.SideMonsters}) {
			t.Errorf("surprise bits 3 hand over %v", got)
		}
	})
	if e.world.Morale != 100 || e.world.Surprise != 0 {
		t.Errorf("morale %d, surprise %d after the first round began, want 100 and 0", e.world.Morale, e.world.Surprise)
	}
}

// TestFightWordsEnd: after a won fight Beaten is the beaten count and CombatEnd 0; after a fled one
// CombatEnd is 0x81.
func TestFightWordsEnd(t *testing.T) {
	e := devFirstTurn(t, "winfight", false)
	for range 300 {
		s := e.build()
		if q := commonOf(s).Ask; q != nil && q.Text == "Continue Battle:" {
			e.doNow(no(q))
			break
		}
		e.doNow(game.Quick{})
	}
	if e.world.Beaten != 1 || e.world.CombatEnd != 0 || e.world.AttackedAlly != 0 {
		t.Errorf("won: beaten %d, end %#x, attacked an ally %d; want 1, 0, 0", e.world.Beaten, e.world.CombatEnd, e.world.AttackedAlly)
	}
	e = devFirstTurn(t, "fleefight", false)
	for range 300 {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			break
		}
		if q := fightQ(s); q != nil {
			e.doNow(yes(q))
		} else if myTurn(s) {
			e.doNow(game.Flee{})
		} else if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	if e.world.CombatEnd != combatEndFled || e.world.Beaten != 0 {
		t.Errorf("fled: end %#x, beaten %d; want %#x, 0", e.world.CombatEnd, e.world.Beaten, combatEndFled)
	}
}
