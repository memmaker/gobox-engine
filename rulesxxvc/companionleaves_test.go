package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
)

// postcbt.pas PostCombatCreatureCleanup removes every creature out of play, and combatst.pas InitCreatureCombatData puts
// every one past the party's count out of play: a friendly companion loaded for the fight (EclLoadMonster, BR1's NEO
// WARRIORs) leaves with the hostiles, the members stay. Detects the class: a fight's loaded creature kept in the party.
func TestCompanionLeavesAfterFight(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	r.party, r.memberCount = []rules.CharID{a}, 1
	ally := r.Add(Character{Side: rules.SideAuto, Companion: true, Status: StOkay, Enabled: true, MaxHP: 5, HP: 5})
	foe := r.Add(Character{Side: rules.SideMonsters, Status: StDead, XPAward: 40})
	r.party = append(r.party, ally, foe) // EclLoadMonster appends them; the count stays
	f := rules.Fight{Party: rules.Party{a, ally}, Board: rules.Board{Units: []rules.BoardUnit{{ID: foe}}}}
	r.BeginFight(f, &rules.RNG{State: 1})
	r.EndFight(f, nil)
	if !slices.Equal(r.party, []rules.CharID{a}) {
		t.Fatalf("party after the fight: %v, want only the member %v", r.party, a)
	}
	if _, ok := r.chars[ally]; ok {
		t.Fatal("the companion is still one of the ruleset's characters")
	}
}
