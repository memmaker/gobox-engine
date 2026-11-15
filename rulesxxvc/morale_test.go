package rulesxxvc

import (
	"image"
	"testing"

	"gobox/rules"
)

// CheckMorale: a TEAM companion at 1 of 10 hit points fails both tests; as
// fast as its foe (halved) it flees, slower with INT 9 it surrenders.
func TestCheckMorale(t *testing.T) {
	r := New()
	a := member("a", 10, 0, 0, 0)
	a.Companion, a.Morale, a.HP, a.MaxHP, a.BaseMovement = true, 40, 1, 10, 12
	a.Scores[rules.Intelligence] = 9
	aid := r.Add(a)
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 12})
	f := rules.Fight{Party: rules.Party{aid}, Board: rules.Board{Size: image.Pt(5, 5), Units: []rules.BoardUnit{{ID: aid}, {ID: m}}}}
	r.BeginFight(f, nil)
	if r.hpPercent != 100 {
		t.Fatalf("hpPercent %d", r.hpPercent)
	}
	s := &combatState{}
	if r.checkMorale(f, r.chars[aid], s) || !s.fleeing {
		t.Errorf("equal speed: fleeing %v", s.fleeing)
	}
	r.chars[m].BaseMovement = 24
	if !r.checkMorale(f, r.chars[aid], s) || r.chars[aid].Status != StFleeing || r.chars[aid].Enabled {
		t.Errorf("slower: %+v", r.chars[aid])
	}
	if fleeDirection(0, false) != 7 || fleeDirection(0, true) != 3 || fleeDirection(2, false) != 2 {
		t.Error("fleeDirection")
	}
}

// A fleeing monster with no enemy left runs (AiMoveAttack's flee loop comes
// before the target), not FinishRound.
func TestFleeWithoutTarget(t *testing.T) {
	r := New()
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 12})
	f := rules.Fight{Board: rules.Board{Size: image.Pt(5, 5), Units: []rules.BoardUnit{{ID: m, At: rules.BoardPos{X: 2, Y: 2}}}}}
	r.fight = map[rules.CharID]*combatState{m: {aiBegun: true, fleeing: true, initiative: 5, movement: 12, approach: 1}}
	switch a := r.Decide(f, m, &rules.RNG{State: 1}).(type) {
	case rules.Move, rules.Flee:
	default:
		t.Fatalf("fleeing without a target: %T %+v", a, a)
	}
}
