package pool

import (
	"gobox/core"
	"rules1e"
	"testing"
)

// A human fighter 5 becomes a magic-user 1; the fighter's levels come back
// once the magic-user reaches level 6, and only the magic-user trains.
func TestChangeClass(t *testing.T) {
	e := testEngine(t)
	e.state().Debug = true
	c := &rules1e.Character{}
	c.Race, c.Class, c.Alignment = rules1e.RaceHuman, rules1e.ClsFighter, 0
	c.Stats = [6]rules1e.Score{17, 17, 10, 10, 10, 10}
	c.Levels[rules1e.ClsFighter], c.Exp = 5, 40000
	e.state().Party = []*rules1e.Character{c}
	if got := rules1e.NewClasses(c); len(got) != 1 || got[0] != rules1e.ClsMage {
		t.Fatalf("new classes %v, want only the magic-user", got)
	}
	offers := e.classOffers(e.charID(c))
	if len(offers) != 1 || offers[0].Name != rules1e.ClassNames[rules1e.ClsMage] {
		t.Fatalf("offers %+v, want the magic-user", offers)
	}
	if _, err := e.ruleset().Resolve(core.Training{Party: e.partyIDs(), DualClass: true}, offers[0].Action, nil); err != nil {
		t.Fatal(err)
	}
	if c.Levels[rules1e.ClsFighter] != 0 || c.OldLevels[rules1e.ClsFighter] != 5 || c.Levels[rules1e.ClsMage] != 1 || c.Exp != 0 || rules1e.NewClasses(c) != nil {
		t.Fatalf("after the change: levels %v old %v exp %d", c.Levels, c.OldLevels, c.Exp)
	}
	for c.Levels[rules1e.ClsMage] < 6 {
		mask, _, _ := e.state().TrainCheck(c)
		e.state().TrainLevels(c, mask)
	}
	if c.Levels[rules1e.ClsFighter] != 5 || c.Levels[rules1e.ClsMage] != 6 {
		t.Errorf("at magic-user 6: levels %v", c.Levels)
	}
	if mask, _, _ := e.state().TrainCheck(c); mask&byte(rules1e.HpClassMask[rules1e.ClsFighter]) != 0 {
		t.Errorf("trains mask %x: the fighter trains again", mask)
	}
}
