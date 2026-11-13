package rules1e

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobox/rules"
)

// TestGoldenAffectsSotsb: every case of sotsb's affect-handler parity
// (what its own handlers did to the rule globals, the characters, their
// affects, the messages and the dice; captured before its rules/ went)
// gives the same result from rules1e's SotSB handlers over the golden rules.json (with rules1e's own
// tables for the ids, variants and groups). The numbers in the file are
// sotsb's own (affects, spells); the test maps them through SotsbAffectIDs.

type goldenAffAffect struct {
	Type       AffectType
	Time       uint16
	Data, Call byte
}

type goldenAffWho struct {
	HP, MaxHP, Con, Side, Control, Auto, MaxLevel, Ranger, Race byte
	AC, Attacks                                                 [2]int
	Creature, Monster, Delay, Spell, CanCast                    byte
	Affects                                                     []goldenAffAffect
}

type goldenAffChar struct {
	HP, Side, Control, Auto, Status, Okay, Int, Wis byte
	AC                                              [2]int
	Combat                                          [6]int // Delay Spell Guarding Moves CanCast Turned
	Target                                          int
	Affects                                         []goldenAffAffect
}

type goldenAffSnap struct {
	Save, Damage, DmgType, SaveType byte
	HitRoll                         int8
	AffVal                          AffectType
	Veto                            byte
	Seed                            uint32
	Chars                           []goldenAffChar
	Msgs                            []string
}

type goldenAffCase struct {
	Name            string
	Typ             AffectType
	Data, Flag      byte
	Who             []goldenAffWho
	Acting          int
	DmgType, Damage byte
	HitRoll         int8
	Save, SaveTyp   byte
	Made            bool
	AffVal          AffectType
	Seed            uint32
	Want            goldenAffSnap
}

func sotsbSpellNumber(id SpellID) byte {
	for n, v := range SotsbSpellIDs {
		if v == id {
			return n
		}
	}
	return 0
}

// sotsbAffect is sotsb's affect number n as the unified id (0: none).
func sotsbAffect(n AffectType) AffectType {
	if n == 0 {
		return 0
	}
	id, ok := SotsbAffectIDs[byte(n)]
	if !ok {
		panic(fmt.Sprintf("sotsb affect %d has no unified id", n))
	}
	return id
}

func runGoldenAffect(tb *Tables, k goldenAffCase) goldenAffSnap {
	e := &RuleState{Tables: tb}
	e.RandSeed = k.Seed
	var cs []*Character
	for _, w := range k.Who {
		c := &Character{}
		c.HP, c.MaxHP, c.Side, c.Control, c.Auto, c.MaxLevel, c.Race = w.HP, w.MaxHP, w.Side, w.Control, w.Auto, w.MaxLevel, RaceID(w.Race)
		c.Stats[StatCon] = Score(w.Con)
		c.Levels[ClsRanger] = Level(w.Ranger)
		c.AC = [2]ArmourClass{ArmourClass(w.AC[0]), ArmourClass(w.AC[1])}
		c.Attacks = [2]Count{Count(w.Attacks[0]), Count(w.Attacks[1])}
		c.CreatureFlags, c.MonsterFlags = CreatureFlags(w.Creature), MonsterFlags(w.Monster)
		c.Combat = &CombatInfo{Delay: w.Delay, Spell: SotsbSpellIDs[w.Spell], CanCast: w.CanCast}
		for _, a := range w.Affects {
			c.Affects = append(c.Affects, &Affect{AffectRecord: AffectRecord{Type: sotsbAffect(a.Type), Time: a.Time, Data: a.Data, Call: a.Call}})
		}
		cs = append(cs, c)
	}
	if len(cs) > 1 {
		cs[0].Combat.Target = cs[1]
	}
	e.Party, e.Who = cs, k.Acting
	e.Rv.DmgType, e.Rv.Damage, e.Rv.HitRoll, e.Rv.Save, e.Rv.SaveType, e.Rv.SaveMade = k.DmgType, int16(k.Damage), k.HitRoll, k.Save, k.SaveTyp, k.Made
	e.Rv.AffVal = sotsbAffect(k.AffVal)
	id := sotsbAffect(k.Typ)
	a := &Affect{AffectRecord: AffectRecord{Type: id, Data: k.Data}}
	cs[0].Affects = append(cs[0].Affects, a)
	e.CallAffect(id, cs[0], a, k.Flag)
	s := goldenAffSnap{Save: e.Rv.Save, Damage: byte(e.Rv.Damage), DmgType: e.Rv.DmgType, SaveType: e.Rv.SaveType, HitRoll: e.Rv.HitRoll,
		AffVal: e.Rv.AffVal, Veto: e.AffectVeto, Seed: e.RandSeed}
	for _, c := range cs {
		cc := goldenAffChar{HP: c.HP, Side: c.Side, Control: c.Control, Auto: c.Auto, Status: c.Status, Okay: c.Okay,
			Int: byte(c.Stats[StatInt]), Wis: byte(c.Stats[StatWis]), AC: [2]int{int(c.AC[0]), int(c.AC[1])}, Target: -1}
		ci := c.Combat
		cc.Combat = [6]int{int(ci.Delay), int(sotsbSpellNumber(ci.Spell)), int(ci.Guarding), int(ci.Moves), int(ci.CanCast), int(ci.Turned)}
		for i, p := range cs {
			if ci.Target == p {
				cc.Target = i
			}
		}
		for _, x := range c.Affects {
			cc.Affects = append(cc.Affects, goldenAffAffect{x.Type, x.Time, x.Data, x.Call})
		}
		s.Chars = append(s.Chars, cc)
	}
	for _, f := range e.Fx {
		if n, ok := f.(rules.Narrated); ok {
			s.Msgs = append(s.Msgs, strings.ToLower(strings.TrimSpace(n.Text)))
		}
	}
	return s
}

func TestGoldenAffectsSotsb(t *testing.T) {
	dir := filepath.Join("testdata", "sotsb")
	var cases []goldenAffCase
	if err := readGolden(filepath.Join(dir, "affects.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for _, k := range cases {
		want := k.Want
		want.AffVal = sotsbAffect(want.AffVal)
		for i := range want.Chars {
			for j := range want.Chars[i].Affects {
				want.Chars[i].Affects[j].Type = sotsbAffect(want.Chars[i].Affects[j].Type)
			}
		}
		if got := runGoldenAffect(&tb, k); !reflect.DeepEqual(got, want) {
			t.Errorf("%s:\n rules1e %+v\n golden  %+v", k.Name, got, k.Want)
		}
	}
	t.Logf("%d cases", len(cases))
}
