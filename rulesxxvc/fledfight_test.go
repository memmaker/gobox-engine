package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
)

// TestEndFightFled: postcbt.pas TEAM_FLED: the fled come back okay, every other member (a downed one, a companion
// too) is removed with what it carried (CharGen.RemoveCreature(False, True)), the party count goes down and the
// acting member is the one before the removed; no experience, no credits and no items (PartyWon only). An okay
// member that did not flee cannot be: it is enabled, so the fight is not a flight (hasPartyFled is cleared by it).
func TestEndFightFled(t *testing.T) {
	r := New()
	first, fled, down, companion := r.Add(member("first", 10, 0, 0, 10)), r.Add(member("fled", 10, 0, 0, 0)),
		r.Add(member("down", 10, 0, 0, 30)), r.Add(member("comp", 10, 0, 0, 0))
	mon := r.Add(Character{Side: rules.SideMonsters, Status: StOkay, Enabled: true, XPAward: 100, Coins: 50})
	r.chars[companion].Companion = true
	r.party, r.memberCount, r.acting = []rules.CharID{first, fled, down, companion, mon}, 4, 3
	r.chars[first].Status, r.chars[first].Enabled = StDead, false
	r.chars[fled].Status, r.chars[fled].Enabled = StFleeing, false
	r.chars[down].Status, r.chars[down].Enabled = StUnconscious, false
	r.chars[companion].Status, r.chars[companion].Enabled = StDying, false
	f := rules.Fight{Party: rules.Party{first, fled, down, companion}, Board: rules.Board{Units: []rules.BoardUnit{{ID: mon}}}}
	if v := r.Verdict(f); v != rules.PartyFled {
		t.Fatal(v)
	}
	res := r.EndFight(f, nil)
	if res.Exp != 0 || res.Money.Copper != 0 || len(res.Items) != 0 || res.Beaten != 0 {
		t.Errorf("a flight paid: %+v", res)
	}
	if !slices.Equal(r.party, []rules.CharID{fled}) || r.memberCount != 1 || r.acting != 0 {
		t.Errorf("party %v count %d acting %d, want [%d] 1 0", r.party, r.memberCount, r.acting, fled)
	}
	if c := r.chars[fled]; c == nil || c.Status != StOkay || !c.Enabled || c.Exp != 0 {
		t.Errorf("the fled member: %+v", c)
	}
	for _, id := range []rules.CharID{first, down, companion, mon} {
		if r.chars[id] != nil {
			t.Errorf("%d is still a character", id)
		}
	}
}
