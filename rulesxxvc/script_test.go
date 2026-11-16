package rulesxxvc

import (
	"strings"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

func TestSkillTitlesParallelSkills(t *testing.T) {
	if len(xxvc.SkillTitles) != len(xxvc.Skills) {
		t.Fatalf("%d titles, %d skills", len(xxvc.SkillTitles), len(xxvc.Skills))
	}
}

// EclSkillCheck's message over a script's skill number: the best character succeeds, the party or the one
// tried fails; the number is skillDefinitions' index, 1-based (13 = First Aid, a number outside the 84 fails badly).
func TestSkillRollScriptNumber(t *testing.T) {
	r := New()
	a, b := r.Add(member("Ann", 10, 0, 60, 0)), r.Add(member("Bob", 10, 0, 0, 0))
	got := r.Check(rules.SkillRoll{Whom: []rules.CharID{a, b}, Script: 13, All: true, Hardness: rules.VeryEasy}, nil)
	if got.Who != a || got.Grade != rules.MadeWell || got.Text != "Ann succeeds at First Aid" {
		t.Errorf("%+v", got)
	}
	got = r.Check(rules.SkillRoll{Whom: []rules.CharID{b}, Script: 13, Hardness: rules.Normal}, nil)
	if got.Who != b || got.Grade != rules.FailedBadly || got.Text != "Bob fails at First Aid" {
		t.Errorf("one: %+v", got)
	}
	got = r.Check(rules.SkillRoll{Whom: []rules.CharID{b}, Script: 13, All: true, Hardness: rules.Normal}, nil)
	if got.Text != "party fails at First Aid" {
		t.Errorf("party: %+v", got)
	}
	if got = r.Check(rules.SkillRoll{Whom: []rules.CharID{a}, Script: 85}, nil); got.Grade != rules.FailedBadly || got.Who != rules.NoChar {
		t.Errorf("85: %+v", got)
	}
}

// EclFindItem over an item type: ItemByNumber gives the kind (xxvc.ItemKinds), HasItem finds the carrier.
func TestHasItemByType(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: tables()}, rules.Options{})
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	r.tables.ItemTypes = append(r.tables.ItemTypes, make([]xxvc.ItemType, 9)...)
	r.tables.Items[8] = xxvc.ItemTemplate{Type: 9} // type 9 is the laser pistol's kind
	r.chars[b].Items = []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 7}}, {Slot: 2, Ref: rules.ItemRef{ID: 8}}}
	rope, err := r.Fact(rules.NoChar, rules.ItemByNumber{N: 29})
	if err != nil || rope != xxvc.ItemIDRope {
		t.Fatalf("%v %v", rope, err)
	}
	if _, err := r.Fact(rules.NoChar, rules.ItemByNumber{N: 0}); err == nil {
		t.Error("type 0 has no kind")
	}
	if got := r.Check(rules.HasItem{Whom: []rules.CharID{a, b}, Item: xxvc.ItemIDLaserPistol}, nil); !got.Yes || got.Who != b {
		t.Errorf("the pistol is b's: %+v", got)
	}
	if r.Check(rules.HasItem{Whom: []rules.CharID{a, b}, Item: xxvc.ItemIDRope}, nil).Yes {
		t.Error("nobody carries a rope")
	}
}

// AddCreatureToParty: the monster made is in no list until it joins; DressNPC changes nothing; the npc byte
// is the morale and the non-player flag; ApplyItems recalculates (EclAddNpc).
func TestNPCJoinRules(t *testing.T) {
	r := New()
	r.Begin(rules.Campaign{Tables: tables()}, rules.Options{})
	if !r.Check(rules.MonsterKnown{ID: 3}, nil).Yes || r.Check(rules.MonsterKnown{ID: 4}, nil).Yes {
		t.Error("MonsterKnown")
	}
	o, err := r.Resolve(nil, rules.MakeMonster{Of: 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	id := o.Effects[0].(rules.MonsterMade).Who
	if len(r.party) != 0 {
		t.Fatal("made monsters are in no list")
	}
	if o, err = r.Resolve(nil, rules.DressNPC{Who: id}, nil); err != nil || o.Effects[0].(rules.NPCDressed).Who != id {
		t.Fatal(err, o)
	}
	if _, err = r.Resolve(nil, rules.JoinParty{Who: id}, nil); err != nil {
		t.Fatal(err)
	}
	if err = r.SetFact(id, rules.FactControl, rules.Number(50|0x80)); err != nil {
		t.Fatal(err)
	}
	if c := r.chars[id]; c.Morale != 50 || !c.NonPlayer {
		t.Errorf("%+v", c)
	}
	if v, _ := r.Fact(id, rules.FactControl); v != rules.Number(50|0x80) {
		t.Errorf("control %v", v)
	}
	if _, err = r.Resolve(nil, rules.ApplyItems{Who: id}, nil); err != nil || !strings.Contains(r.chars[id].Name, "trooper") {
		t.Error(err)
	}
}

// itemRig is two members: a carries a rope stack of 3 (joinable) and a laser pistol; b a rope.
func itemRig() (*Rules, rules.CharID, rules.CharID) {
	r := New()
	r.Begin(rules.Campaign{Tables: xxvc.Tables{
		ItemTypes: make([]xxvc.ItemType, 30),
		Items: map[rules.ItemID]xxvc.ItemTemplate{
			1: {Type: 29, Quantity: 1}, 2: {Type: 9, Quantity: 1}, 3: {Type: 29, Quantity: 1, NotJoinable: true},
		}}}, rules.Options{})
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	r.chars[a].Items = []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 3}}, {Slot: 2, Ref: rules.ItemRef{ID: 2, Count: 1}}}
	r.chars[b].Items = []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 3, Count: 2}}}
	return r, a, b
}

// EclDestroyItems: all of the kind go; with the high bit one each: a joinable stack loses one, anything else goes whole.
func TestTakeItems(t *testing.T) {
	r, a, b := itemRig()
	if _, err := r.Resolve(nil, rules.TakeItems{Whom: rules.Party{a, b}, Item: xxvc.ItemIDRope, OneEach: true}, nil); err != nil {
		t.Fatal(err)
	}
	if ca := r.chars[a]; len(ca.Items) != 2 || ca.Items[0].Ref.Count != 2 {
		t.Errorf("a's stack of 3 loses one: %+v", ca.Items)
	}
	if len(r.chars[b].Items) != 0 {
		t.Errorf("b's unjoinable rope goes whole: %+v", r.chars[b].Items)
	}
	if _, err := r.Resolve(nil, rules.TakeItems{Whom: rules.Party{a}, Item: xxvc.ItemIDRope}, nil); err != nil {
		t.Fatal(err)
	}
	if ca := r.chars[a]; len(ca.Items) != 1 || ca.Items[0].Ref.ID != 2 {
		t.Errorf("all the rope goes, the pistol stays: %+v", ca.Items)
	}
}

// EclTransferCharacterLists: the source list goes in front of the destination's.
func TestStashItems(t *testing.T) {
	r, a, _ := itemRig()
	c := r.chars[a]
	c.Stashed = []Carried{{Slot: 9, Ref: rules.ItemRef{ID: 3}}}
	if _, err := r.Resolve(nil, rules.StashItems{Whom: rules.Party{a}, Away: true}, nil); err != nil {
		t.Fatal(err)
	}
	if len(c.Items) != 0 || len(c.Stashed) != 3 || c.Stashed[0].Ref.ID != 1 || c.Stashed[2].Ref.ID != 3 {
		t.Fatalf("away: %+v / %+v", c.Items, c.Stashed)
	}
	c.Items = []Carried{{Slot: 5, Ref: rules.ItemRef{ID: 2}}}
	if _, err := r.Resolve(nil, rules.StashItems{Whom: rules.Party{a}}, nil); err != nil {
		t.Fatal(err)
	}
	if len(c.Stashed) != 0 || len(c.Items) != 4 || c.Items[0].Ref.ID != 1 || c.Items[3].Ref.ID != 2 {
		t.Fatalf("back: %+v / %+v", c.Items, c.Stashed)
	}
}

// EclSkillDamage: who fails takes the damage (a Byte) and is reported; "dies" when hit points + 10 do not cover
// it; a member who is not okay is passed over; who makes the roll is not hurt.
func TestSkillDamage(t *testing.T) {
	r := New()
	weak := member("weak", 10, 0, 0, 0)
	weak.HP = 3
	hurt, skilled, down := r.Add(member("hurt", 10, 0, 0, 0)), r.Add(member("skilled", 10, 0, 60, 0)), r.Add(weak)
	r.chars[hurt].HP, r.chars[skilled].HP = 20, 20
	r.chars[down].Status = StDying
	o, err := r.Resolve(nil, rules.SkillDamage{Whom: rules.Party{hurt, skilled, down}, Skill: 13, Hardness: rules.VeryEasy, Dice: 1, Sides: 1, Bonus: 4}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Effects) != 1 || o.Effects[0] != (rules.Hurt{Unit: hurt, Damage: 5}) || r.chars[hurt].HP != 15 || r.chars[skilled].HP != 20 {
		t.Fatalf("%+v hurt %d skilled %d", o.Effects, r.chars[hurt].HP, r.chars[skilled].HP)
	}
	r.chars[down].Status, r.chars[down].Enabled = StOkay, true
	o, _ = r.Resolve(nil, rules.SkillDamage{Whom: rules.Party{down}, Skill: 13, Hardness: rules.VeryEasy, Dice: 1, Sides: 1, Bonus: 14}, nil)
	if len(o.Effects) != 1 || !o.Effects[0].(rules.Hurt).Dies || r.chars[down].Status != StComatose {
		t.Errorf("3 + 10 < 15 (overkill 12: comatose): %+v %v", o.Effects, r.chars[down].Status)
	}
}

// soloRig is a team of two, the second the current character, and one downed monster worth 30.
func soloRig() (r *Rules, other, duelist, monster rules.CharID) {
	r = New()
	other, duelist = r.Add(member("other", 10, 0, 0, 0)), r.Add(member("duelist", 10, 0, 0, 0))
	foe := member("foe", 10, 0, 0, 0)
	foe.Side, foe.Status, foe.Enabled, foe.XPAward = rules.SideMonsters, StDead, false, 30
	monster = r.Add(foe)
	r.party, r.memberCount, r.acting = []rules.CharID{other, duelist, monster}, 2, 1
	return
}

// EclCombat's solo combat: the others sit out (disabled, so placed on no board); a win gives the duelist the
// whole xpAward and no one else, and the duelist is okay and enabled again; the monsters leave.
func TestSoloCombatWon(t *testing.T) {
	r, other, duelist, monster := soloRig()
	f := rules.Fight{Terms: rules.FightTerms{Duel: true}}
	r.BeginFight(f, nil)
	if r.chars[other].Enabled || !r.chars[duelist].Enabled || r.duelist != duelist {
		t.Fatalf("other %v duelist %v (%d)", r.chars[other].Enabled, r.chars[duelist].Enabled, r.duelist)
	}
	res := r.EndFight(f, nil)
	if res.Verdict != rules.PartyWon || res.Exp != 30 || res.Beaten != 1 || r.chars[duelist].Exp != 30 || r.chars[other].Exp != 0 {
		t.Errorf("%+v exp %d / %d", res, r.chars[duelist].Exp, r.chars[other].Exp)
	}
	if !r.chars[other].Enabled || r.chars[monster] != nil || len(r.party) != 2 {
		t.Errorf("sat-out member enabled %v, monster %v, party %v", r.chars[other].Enabled, r.chars[monster], r.party)
	}
}

// A beaten duelist: lost, the dying fall unconscious, no experience, no one leaves the party.
func TestSoloCombatLost(t *testing.T) {
	r, other, duelist, _ := soloRig()
	f := rules.Fight{Terms: rules.FightTerms{Duel: true}}
	r.BeginFight(f, nil)
	r.chars[duelist].Status, r.chars[duelist].Enabled = StDying, false
	res := r.EndFight(f, nil)
	if res.Verdict != rules.PartyLost || res.Exp != 0 || r.chars[duelist].Status != StUnconscious || len(r.party) != 2 || !r.chars[other].Enabled {
		t.Errorf("%+v status %v party %v", res, r.chars[duelist].Status, r.party)
	}
}

// eclvm.pas EclDamage: the whole party without a save (flags 0x40|0x20) takes Dice d Sides + Bonus each, a member
// not okay is passed over, "dies" when the hit points and ten do not cover it; the selected member alone (BySelected,
// no save) takes it once. 1d1 + 4 is always 5.
func TestScriptDamage(t *testing.T) {
	r := New()
	ann, bob, cy := member("Ann", 10, 0, 0, 0), member("Bob", 10, 0, 0, 0), member("Cy", 10, 0, 0, 0)
	ann.HP, bob.HP, cy.HP, cy.Status = 20, 3, 20, StDead
	a, b, c := r.Add(ann), r.Add(bob), r.Add(cy)
	r.party, r.memberCount = []rules.CharID{a, b, c}, 3
	o, err := r.Resolve(rules.Scripted{Party: r.party}, rules.ScriptDamage{Selected: a, Dice: 1, Sides: 1, Bonus: 4, All: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Effects) != 2 || r.chars[a].HP != 15 || r.chars[b].HP != 0 || r.chars[c].HP != 20 {
		t.Errorf("party: effects %v, hit points %d %d %d, want 15 0 20 and two Hurt", o.Effects, r.chars[a].HP, r.chars[b].HP, r.chars[c].HP)
	}
	if h := o.Effects[1].(rules.Hurt); h.Unit != b || h.Damage != 5 || h.Dies {
		t.Errorf("Bob: %+v (3 hit points and 10 cover 5)", h)
	}
	o, err = r.Resolve(rules.Scripted{Party: r.party}, rules.ScriptDamage{Selected: a, Dice: 1, Sides: 1, Bonus: 30, BySelected: true}, nil)
	if err != nil || len(o.Effects) != 1 || !o.Effects[0].(rules.Hurt).Dies || r.chars[a].HP != 0 {
		t.Errorf("selected: %v %v, hit points %d", o.Effects, err, r.chars[a].HP)
	}
}

// The party killed by a script ends the session (the engine's unwindReset): XXVc has nothing open to close.
func TestEndSession(t *testing.T) {
	if _, err := New().Resolve(nil, rules.EndSession{}, nil); err != nil {
		t.Error(err)
	}
}
