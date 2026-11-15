package rules1e

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

// TestGoldenAffectsPod: POD's own affect handlers (captured before its
// rules/ went) did to the rule globals, the bearer's strength, affects, armour
// class and combat block what rules1e's POD handlers do over the golden
// rules.json, in three kinds of case: "damage" (the attack taken, by an
// attacker with a weapon), "char" (the character's own rolls and strength)
// and "defence" (the defender's change of the attack). Affect numbers are
// POD's own; the test maps them through PodAffectIDs.

type goldenPodCase struct {
	Kind                  string
	Affect                byte
	Flag                  byte
	Damage                int16
	DmgType               byte
	AffVal                byte
	Weapon                bool
	Plus                  int8
	Material              byte
	Monster               bool
	MaxLevel              byte
	Delay, Moves          byte
	Data, Side, Str, Cure byte
	Sees                  bool
	Want                  struct {
		Damage       int
		AffVal       byte
		Delay, Moves byte
		HitRoll      int8
		Save         byte
		Str          byte
		Affects      json.RawMessage // char: the records; defence: their number
		Data         byte
		AC           [2]int
	}
}

const (
	goldenPodSceneWeapon = 5  // the base item the scenes' weapon is
	goldenPodDetectInvis = 24 // detect invisibility, POD's number
)

func TestGoldenAffectsPod(t *testing.T) {
	dir := filepath.Join("testdata", "pod")
	var cases []goldenPodCase
	if err := readGolden(filepath.Join(dir, "affects.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for i, k := range cases {
		var bad string
		switch k.Kind {
		case "damage":
			bad = runGoldenPodDamage(&tb, k)
		case "char":
			bad = runGoldenPodChar(&tb, k)
		case "defence":
			bad = runGoldenPodDefence(&tb, k)
		default:
			t.Fatalf("case %d: kind %q", i, k.Kind)
		}
		if bad != "" {
			t.Fatalf("case %d %+v: %s", i, k, bad)
		}
	}
	t.Logf("%d cases", len(cases))
}

// podAffect is POD's affect number n as the unified id (0: none).
func podAffect(n byte) AffectType {
	if n == 0 {
		return 0
	}
	return PodAffectIDs[n]
}

func runGoldenPodDamage(tb *Tables, k goldenPodCase) string {
	r := &RuleState{Tables: tb, Items: make([]ItemBase, goldenPodSceneWeapon+1)}
	r.Items[goldenPodSceneWeapon].Material = k.Material
	atk := &Character{}
	atk.MaxLevel, atk.Race = k.MaxLevel, RaceHuman
	if k.Monster {
		atk.Race = RaceMonster
	}
	if k.Weapon {
		w := &Item{}
		w.Type, w.Plus = goldenPodSceneWeapon, byte(k.Plus)
		atk.Slots[SlotWeapon] = w
	}
	c := &Character{Combat: &CombatInfo{Delay: k.Delay, Moves: k.Moves}}
	r.Party = []*Character{atk, c}
	r.Rv.Damage, r.Rv.DmgType, r.Rv.AffVal = k.Damage, k.DmgType, podAffect(k.AffVal)
	id := podAffect(k.Affect)
	r.CallAffect(id, c, &Affect{AffectRecord: AffectRecord{Type: id}}, k.Flag)
	if int(r.Rv.Damage) != k.Want.Damage || r.Rv.AffVal != podAffect(k.Want.AffVal) || c.Combat.Delay != k.Want.Delay || c.Combat.Moves != k.Want.Moves {
		return "rules1e damage/affect/delay/moves differ from the golden"
	}
	return ""
}

func runGoldenPodChar(tb *Tables, k goldenPodCase) string {
	r := &RuleState{Tables: tb}
	atk, c := &Character{}, &Character{}
	if k.Sees {
		atk.Affects = []*Affect{{AffectRecord: AffectRecord{Type: podAffect(goldenPodDetectInvis)}}}
	}
	c.Side, c.Stats[StatStr] = k.Side, Score(k.Str)
	r.Party = []*Character{atk, c}
	r.Rv.Cure = k.Cure
	id := podAffect(k.Affect)
	r.CallAffect(id, c, &Affect{AffectRecord: AffectRecord{Type: id, Data: k.Data}}, k.Flag)
	var got []AffectRecord
	for _, a := range c.Affects {
		got = append(got, a.AffectRecord)
		if a.Type == affWeakened { // Felix's ruling: its data counts the Strength it took; the original passes the affect's own on
			got[len(got)-1].Data = k.Data
		}
	}
	var want []AffectRecord
	if err := json.Unmarshal(k.Want.Affects, &want); err != nil {
		return err.Error()
	}
	for i := range want {
		want[i].Type = podAffect(byte(want[i].Type))
	}
	if r.Rv.HitRoll != k.Want.HitRoll || r.Rv.Save != k.Want.Save || byte(c.Stats[StatStr]) != k.Want.Str || !reflect.DeepEqual(got, want) {
		return "rules1e hit roll/save/strength/affects differ from the golden"
	}
	return ""
}

func runGoldenPodDefence(tb *Tables, k goldenPodCase) string {
	r := &RuleState{Tables: tb}
	c := &Character{}
	c.AC = [2]ArmourClass{50, 52}
	id := podAffect(k.Affect)
	a := &Affect{AffectRecord: AffectRecord{Type: id, Data: k.Data}}
	c.Affects = []*Affect{a}
	r.Party = []*Character{{}, c}
	r.Rv.Damage, r.Rv.DmgType, r.Rv.AffVal = k.Damage, k.DmgType, podAffect(k.AffVal)
	r.CallAffect(id, c, a, 0)
	var n int
	json.Unmarshal(k.Want.Affects, &n)
	if int(r.Rv.Damage) != k.Want.Damage || r.Rv.AffVal != podAffect(k.Want.AffVal) || r.Rv.HitRoll != k.Want.HitRoll || r.Rv.Save != k.Want.Save ||
		a.Data != k.Want.Data || len(c.Affects) != n || int(c.AC[0]) != k.Want.AC[0] || int(c.AC[1]) != k.Want.AC[1] {
		return "rules1e damage/affect/rolls/data/AC differ from the golden"
	}
	return ""
}
