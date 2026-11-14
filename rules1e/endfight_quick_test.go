package rules1e

import (
	"testing"

	"gobox/rules"
)

// TestEndFightEndsQuick: a member the computer fought for (Quick) is the
// player's again once the fight is over; a hired NPC keeps its Auto.
func TestEndFightEndsQuick(t *testing.T) {
	r := &Rules{S: &RuleState{}}
	m := &Character{Combat: &CombatInfo{}}
	m.Okay, m.Auto = 1, 1
	npc := &Character{Combat: &CombatInfo{}}
	npc.Okay, npc.Auto, npc.Control = 1, 1, 0x80
	r.S.Party = []*Character{m, npc}
	r.EndFight(rules.Fight{}, nil)
	if m.Auto != 0 || npc.Auto != 1 {
		t.Errorf("Auto after the fight: member %d (want 0), NPC %d (want 1)", m.Auto, npc.Auto)
	}
}
