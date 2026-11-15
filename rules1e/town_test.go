package rules1e

import (
	"encoding/json"
	"gobox/rules"
	"strings"
	"testing"
)

// TestTownData: rules.json's town reaches Tables.Town; a free hall asks no fee of
// an empty purse (POR's asks 1000), the temple lists the game's services, and a
// repair mends every member.
func TestTownData(t *testing.T) {
	town := TownData{Temple: []TempleService{{Name: "Heal", Kind: TempleHeal}, {Name: "Repair", Kind: TempleRepair}}}
	b, err := json.Marshal(map[string]any{"thac0Table": POR.Thac0, "town": town})
	if err != nil {
		t.Fatal(err)
	}
	tab, err := POR.WithRules(strings.NewReader(string(b)), Layout{SlotStride: saveCount})
	if err != nil {
		t.Fatal(err)
	}
	free := &RuleState{Tables: &tab}
	c := free.Fresh(Kind{Race: RaceHuman, Class: ClsFighter}, nil)
	c.Status = StatusOkay
	if _, why, _ := free.TrainCheck(c, rules.HallTerms{}); strings.HasPrefix(why, "Training costs") {
		t.Errorf("a free hall asks: %q", why)
	}
	por := &RuleState{}
	if _, why, _ := por.TrainCheck(c, rules.HallTerms{}); why != "Training costs 1000 gp." {
		t.Errorf("POR's hall: %q", why)
	}
	if sv, ok := free.templeService(1); !ok || sv.Kind != TempleRepair || sv.Price != 0 {
		t.Errorf("service 1 %+v", sv)
	}
	if _, ok := free.templeService(2); ok {
		t.Error("a third service")
	}
	if sv, _ := por.templeService(7); sv.Price != 3500 || sv.Not != "is not cursed." {
		t.Errorf("POR's Remove Curse %+v", sv)
	}
	a, d := &Character{}, &Character{}
	a.HP, a.MaxHP, a.Status = 3, 30, StatusOkay
	d.HP, d.MaxHP, d.Status = 0, 20, StatusDead
	free.Party = []*Character{a, d}
	free.TempleCure(a, 1)
	if a.HP < 20 || d.Status != StatusOkay {
		t.Errorf("repair: hp %d, dead's status %d", a.HP, d.Status)
	}
}

// TestTempleWound: a wound cure helps only a hurt member who is alive and here, in every
// game, and refuses another (Felix 2026-11-12); Heal also removes feeblemind (Int and Wis
// again) and is needed for it. Done and Declines reach the core fact.
func TestTempleWound(t *testing.T) {
	town := TownData{Temple: []TempleService{
		{Name: "Light", Kind: TempleLight, Done: "is healed", Declines: true},
		{Name: "Heal", Kind: TempleHeal},
	}}
	b, err := json.Marshal(map[string]any{"thac0Table": POR.Thac0, "town": town})
	if err != nil {
		t.Fatal(err)
	}
	tab, err := POR.WithRules(strings.NewReader(string(b)), Layout{SlotStride: saveCount})
	if err != nil {
		t.Fatal(err)
	}
	s := &RuleState{Tables: &tab}
	c := &Character{}
	c.HP, c.MaxHP, c.Status = 20, 20, StatusOkay
	if s.CureNeeded(c, 0) || s.CureNeeded(c, 1) {
		t.Error("an unhurt member needs no wound cure")
	}
	r := &Rules{S: s, chars: map[rules.CharID]*Character{1: c}}
	if _, err := r.Resolve(nil, rules.GiveTempleCure{Who: 1, Cure: 0}, nil); err == nil {
		t.Error("a wound cure was given to an unhurt member")
	}
	c.Base[StatInt], c.Base[StatWis] = 15, 14 // the own scores (statAffects)
	s.AddAffect(c, affFeebleminded, 0, 0, 0)
	c.Stats[StatInt], c.Stats[StatWis] = lowScore, lowScore
	if !s.CureNeeded(c, 1) {
		t.Fatal("Heal is not needed for feeblemind")
	}
	s.TempleCure(c, 1)
	if s.FindAffect(c, affFeebleminded) != nil || c.Stats[StatInt] != 15 || c.Stats[StatWis] != 14 {
		t.Errorf("Heal left feeblemind: int %d wis %d", c.Stats[StatInt], c.Stats[StatWis])
	}
	c.HP = 5
	if !s.CureNeeded(c, 0) {
		t.Error("a hurt member")
	}
	c.Status = StatusDead
	if s.CureNeeded(c, 0) {
		t.Error("a dead member is not healed by a wound service")
	}
	if sv, _ := s.templeService(0); sv.Done != "is healed" || !sv.Declines {
		t.Errorf("service %+v", sv)
	}
}

// TestTempleDiseaseRestoresStats: a disease cure of a service with RestoresStats (pod's) puts the
// weakened strength back from the base score; without the flag (POR's) the score stays.
func TestTempleDiseaseRestoresStats(t *testing.T) {
	town := TownData{Temple: []TempleService{{Name: "Disease", Kind: TempleDisease, RestoresStats: true}}}
	plain := TownData{Temple: []TempleService{{Name: "Disease", Kind: TempleDisease}}}
	for _, k := range []struct {
		town TownData
		want Score
	}{{town, 16}, {plain, 15}} {
		b, err := json.Marshal(map[string]any{"thac0Table": POR.Thac0, "town": k.town})
		if err != nil {
			t.Fatal(err)
		}
		tab, err := POR.WithRules(strings.NewReader(string(b)), Layout{SlotStride: saveCount})
		if err != nil {
			t.Fatal(err)
		}
		s := &RuleState{Tables: &tab}
		c := &Character{}
		c.Base[StatStr], c.Stats[StatStr] = 16, 15 // one point weakened
		s.cure(c, TempleDisease)
		if c.Stats[StatStr] != k.want {
			t.Errorf("RestoresStats %v: strength %d, want %d", k.town.Temple[0].RestoresStats, c.Stats[StatStr], k.want)
		}
		c.Stats[StatStr] = 15
		s.cure(c, TempleHeal) // pod's Heal (O04C:0199) never reruns the strength
		if c.Stats[StatStr] != 15 {
			t.Errorf("Heal with RestoresStats %v: strength %d, want 15", k.town.Temple[0].RestoresStats, c.Stats[StatStr])
		}
	}
}
