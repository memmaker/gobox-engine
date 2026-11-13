package rules1e

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobox/core"
)

// TestGoldenSpellsSotsb: rules1e's SpellEffect on the spells sotsb has first
// (harm, cure light wounds, disintegrate, flesh to stone, power word stun,
// death spell, mass invisibility, globe of invulnerability ...) leaves what
// sotsb's own spellFx left (captured before its rules/ went): characters,
// affects, messages, dice and the caster's age. Spell and affect numbers in the
// file are sotsb's own.

const goldenSotsbCasterLevel = 10 // mage and cleric level of the golden caster

type goldenSpellChar struct {
	HP, Side, Status, Okay int
	Affects                []goldenAffAffect
}

type goldenSpellSnap struct {
	Seed  uint32
	Age   uint16
	Chars []goldenSpellChar
	Msgs  []string
}

type goldenSpellCase struct {
	Name    string
	Spell   byte
	Caster  goldenAffWho
	Targets []goldenAffWho
	Seed    uint32
	Want    goldenSpellSnap
}

func TestGoldenSpellsSotsb(t *testing.T) {
	dir := filepath.Join("testdata", "sotsb")
	var cases []goldenSpellCase
	if err := readGolden(filepath.Join(dir, "spells.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for _, k := range cases {
		e := &RuleState{Tables: &tb}
		e.RandSeed = k.Seed
		mk := func(w goldenAffWho) *Character {
			c := &Character{}
			c.HP, c.MaxHP, c.Side, c.MaxLevel, c.Race = w.HP, w.MaxHP, w.Side, w.MaxLevel, RaceID(w.Race)
			c.AC = [2]ArmourClass{ArmourClass(w.AC[0]), ArmourClass(w.AC[1])}
			c.CreatureFlags, c.MonsterFlags = CreatureFlags(w.Creature), MonsterFlags(w.Monster)
			c.Levels[ClsMage], c.Levels[ClsCleric] = goldenSotsbCasterLevel, goldenSotsbCasterLevel
			c.Combat = &CombatInfo{}
			c.Age = 30
			for _, a := range w.Affects {
				c.Affects = append(c.Affects, &Affect{AffectRecord: AffectRecord{Type: sotsbAffect(a.Type), Time: a.Time, Data: a.Data, Call: a.Call}})
			}
			return c
		}
		caster := mk(k.Caster)
		cs := []*Character{caster}
		for i, w := range k.Targets {
			c := mk(w)
			cs = append(cs, c)
			e.SetTgt(i+1, c)
		}
		e.NSpellTargets = len(k.Targets)
		e.Party, e.Who = cs, 0
		id := SotsbSpellIDs[k.Spell]
		e.CurSpell = id
		e.SpellEffect(id)
		got := goldenSpellSnap{Seed: e.RandSeed, Age: uint16(caster.Age)}
		for _, c := range cs {
			cc := goldenSpellChar{HP: int(c.HP), Side: int(c.Side), Status: int(c.Status), Okay: int(c.Okay)}
			for _, x := range c.Affects {
				cc.Affects = append(cc.Affects, goldenAffAffect{x.Type, x.Time, x.Data, x.Call})
			}
			got.Chars = append(got.Chars, cc)
		}
		for _, f := range e.Fx {
			if n, ok := f.(core.Narrated); ok {
				got.Msgs = append(got.Msgs, strings.TrimSpace(n.Text))
			}
		}
		want := k.Want
		for i := range want.Chars {
			for j := range want.Chars[i].Affects {
				want.Chars[i].Affects[j].Type = sotsbAffect(want.Chars[i].Affects[j].Type)
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s:\n rules1e %+v\n golden  %+v", k.Name, got, want)
		}
	}
	t.Logf("%d cases", len(cases))
}
