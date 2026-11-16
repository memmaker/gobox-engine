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
