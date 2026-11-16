package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// combat.pas CombatMain ends a fight once combatRound >= combatEndRound; CombatStart sets combatEndRound to IDLE_ROUNDS (15) and
// MakeAttack to combatRound + 15, EndOfRound counts the round. With both sides up and no attack, the fight is over after 15
// rounds; EndFight then follows postcbt.pas DetermineCombatResults: won when a member is okay, the beaten's experience shared.
func TestIdleRoundsEndFight(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Status: StOkay, Enabled: true, XPAward: 100, MaxHP: 5, HP: 5})
	dead := r.Add(Character{Side: rules.SideMonsters, Status: StDead, XPAward: 40})
	f := rules.Fight{Party: rules.Party{a}, Board: rules.Board{Units: []rules.BoardUnit{{ID: m}, {ID: dead}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	for i := range idleRounds {
		if v := r.Verdict(f); v != rules.FightOn {
			t.Fatalf("after %d rounds the fight is %v", i, v)
		}
		r.EndRound(f, &rng)
	}
	if v := r.Verdict(f); v != rules.FightIdle {
		t.Fatalf("after %d idle rounds: %v", idleRounds, v)
	}
	r.combatEndRound = r.combatRound + idleRounds // an attack (MakeAttack) puts the end 15 rounds on
	if v := r.Verdict(f); v != rules.FightOn {
		t.Fatalf("after an attack: %v", v)
	}
	r.combatEndRound = r.combatRound
	res := r.EndFight(f, nil)
	if res.Verdict != rules.PartyWon || res.Exp != 40 || res.Beaten != 1 {
		t.Errorf("result %+v", res)
	}
}
