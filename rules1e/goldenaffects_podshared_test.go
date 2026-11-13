package rules1e

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobox/core"
)

// TestGoldenAffectsPodShared: POD's affects that rules1e plays with a
// handler shared with SotSB (bane weapon, berserk, confused, vorpal weapon,
// weapon sequence, haste, sticks to snakes ...) do what POD's own handlers did
// (captured before its rules/ went): rule globals, characters, affects and
// messages in lower case. Affect numbers in the file are POD's own.

type goldenPodWho struct {
	HP, MaxHP, Side, Control, Auto, Creature, Monster byte
}

type goldenPodSnap struct {
	Damage           int
	HitRoll          int8
	Data, Attacks1   byte
	CanUse, CanCast  byte
	Save             byte
	Seed             uint32
	HP, Side, Status []int // bytes as numbers: a JSON byte slice is a string
	Ctl, Auto        []int
	Target           []int
	Affects          [][]AffectType
	Msgs             []string
}

type goldenPodSharedCase struct {
	Name             string
	Typ, Flag, Data  byte
	Who              []goldenPodWho
	Acting           int
	DmgType, Damage  byte
	Seed             uint32
	Combat, Critical bool
	Blow             byte
	Want             goldenPodSnap
}

const goldenPodBeginFlagCombat = 5 // BeginFlag in a fight

func TestGoldenAffectsPodShared(t *testing.T) {
	dir := filepath.Join("testdata", "pod")
	var cases []goldenPodSharedCase
	if err := readGolden(filepath.Join(dir, "affects_shared.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for _, k := range cases {
		want := k.Want
		for i := range want.Affects {
			for j, n := range want.Affects[i] {
				want.Affects[i][j] = podAffect(byte(n))
			}
		}
		if got := runGoldenPodShared(&tb, k); !reflect.DeepEqual(got, want) {
			t.Errorf("%s:\n rules1e %+v\n golden  %+v", k.Name, got, want)
		}
	}
	t.Logf("%d cases", len(cases))
}

func runGoldenPodShared(tb *Tables, k goldenPodSharedCase) goldenPodSnap {
	id := podAffect(k.Typ)
	r := &RuleState{Tables: tb}
	r.RandSeed = k.Seed
	var cs []*Character
	for _, w := range k.Who {
		c := &Character{}
		c.HP, c.MaxHP, c.Side, c.Control, c.Auto = w.HP, w.MaxHP, w.Side, w.Control, w.Auto
		c.CreatureFlags, c.MonsterFlags = CreatureFlags(w.Creature), MonsterFlags(w.Monster)
		c.Combat = &CombatInfo{CanUse: 1, CanCast: 1}
		cs = append(cs, c)
	}
	cs[0].Combat.Target = cs[1]
	r.Party, r.Who = cs, k.Acting
	r.Rv.Damage, r.Rv.DmgType = int16(k.Damage), k.DmgType
	if k.Combat {
		r.BeginFlag, r.Cmap = goldenPodBeginFlagCombat, &CombatMap{}
	}
	r.Rv.Critical, r.SlotShots[1] = k.Critical, Count(k.Blow)
	a := &Affect{AffectRecord: AffectRecord{Type: id, Data: k.Data}}
	cs[0].Affects = append(cs[0].Affects, a)
	r.CallAffect(id, cs[0], a, k.Flag)
	var s goldenPodSnap
	for i, c := range cs {
		s.HP, s.Side, s.Status = append(s.HP, int(c.HP)), append(s.Side, int(c.Side)), append(s.Status, int(c.Status))
		s.Ctl, s.Auto = append(s.Ctl, int(c.Control)), append(s.Auto, int(c.Auto))
		target := -1
		for j, p := range cs {
			if cs[i].Combat.Target == p {
				target = j
			}
		}
		s.Target = append(s.Target, target)
		var ids []AffectType
		for _, x := range c.Affects {
			ids = append(ids, x.Type)
		}
		s.Affects = append(s.Affects, ids)
	}
	s.Damage, s.HitRoll, s.Save, s.Seed = int(r.Rv.Damage), r.Rv.HitRoll, r.Rv.Save, r.RandSeed
	s.Data, s.Attacks1, s.CanUse, s.CanCast = a.Data, byte(cs[0].Attacks[1]), cs[0].Combat.CanUse, cs[0].Combat.CanCast
	for _, f := range r.Fx {
		if n, ok := f.(core.Narrated); ok {
			s.Msgs = append(s.Msgs, strings.ToLower(strings.TrimSpace(n.Text)))
		}
	}
	return s
}

var _ = fmt.Sprint
