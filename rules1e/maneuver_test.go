package rules1e

import (
	"errors"
	"gobox/rules"
	"testing"
)

// moveRules has one maneuver of each aim and records the Use calls.
type moveRules struct {
	ClassicRules
	used *[]useCall
}

type useCall struct {
	id    int
	t     *Character
	aimed rules.BoardPos
}

const (
	mvWeapon = iota + 1
	mvArea
	mvFoe
	mvSelf
)

func (moveRules) Maneuvers(*RuleState, *Character) []Maneuver {
	return []Maneuver{{ID: mvWeapon, Aim: AimWeapon, Move: 2}, {ID: mvArea, Aim: AimArea, Range: 5, Area: 1},
		{ID: mvFoe, Aim: AimFoe, Range: 3}, {ID: mvSelf, Aim: AimSelf}}
}

func (m moveRules) Use(s *RuleState, _ *Character, id int, t *Character) bool {
	*m.used = append(*m.used, useCall{id, t, s.Aimed()})
	return id != mvWeapon
}

// TestManeuverAims: Resolve(Maneuver) for each aim flow the engine drives
// (combatrules_engine.go): the move aimed first, the cancel, the area, the target.
func TestManeuverAims(t *testing.T) {
	var used []useCall
	s := &RuleState{}
	s.SetRules(moveRules{used: &used})
	r := &Rules{S: s}
	c, foe, ally := &Character{}, &Character{}, &Character{}
	c.Side, foe.Side, ally.Side = SideParty, SideMonsters, SideParty
	f := rules.Fight{Actor: r.ID(c)}
	do := func(a rules.Maneuver) (rules.Outcome, error) { return r.Resolve(f, a, nil) }

	to := rules.BoardPos{X: 4, Y: 5}
	if o, err := do(rules.Maneuver{ID: mvWeapon, Ats: []rules.BoardPos{to}}); err != nil || o.TurnOver {
		t.Fatalf("move: %v %+v", err, o)
	}
	if len(used) != 1 || used[0].aimed != to || s.Fight.Armed != (Armed{C: c, ID: mvWeapon}) || s.Aimed() != (rules.BoardPos{X: -1, Y: -1}) {
		t.Fatalf("move: used %+v armed %+v aimed %v", used, s.Fight.Armed, s.Aimed())
	}
	if _, err := do(rules.Maneuver{ID: mvWeapon, Cancel: true}); err != nil || s.Fight.Armed != (Armed{}) || len(used) != 1 {
		t.Fatalf("cancel: %v armed %+v used %d", err, s.Fight.Armed, len(used))
	}

	at := rules.BoardPos{X: 7, Y: 2}
	if o, err := do(rules.Maneuver{ID: mvArea, Ats: []rules.BoardPos{at}}); err != nil || !o.TurnOver || used[1].aimed != at || used[1].t != nil {
		t.Fatalf("area: %v %+v %+v", err, o, used)
	}

	if _, err := do(rules.Maneuver{ID: mvFoe, Target: r.ID(ally)}); !errors.Is(err, ErrNotNow) || len(used) != 2 {
		t.Fatalf("ally as a foe: %v, used %d", err, len(used))
	}
	if o, err := do(rules.Maneuver{ID: mvFoe, Target: r.ID(foe)}); err != nil || !o.TurnOver || used[2].t != foe {
		t.Fatalf("foe: %v %+v %+v", err, o, used)
	}
	if _, err := do(rules.Maneuver{ID: 99}); !errors.Is(err, ErrNotNow) {
		t.Fatalf("unknown maneuver: %v", err)
	}
}
