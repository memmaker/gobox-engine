package rules1e

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"gobox/rules"
)

// TestGoldenAffectsCotab: each affect handler of cotab's own (captured before
// its rules/ went) run from one of the 48 scenes gave the rule globals, the
// attacker's, bearer's and target's stats, combat block, affects (as unified
// ids), messages and dice the golden holds; rules1e's CotAB handlers over
// the golden rules.json give the same.

type goldenCotabScene struct {
	DmgType, Damage   byte
	AffVal            byte
	Data, Flag, Save  byte
	HitRoll           int8
	SaveMade, Combat  bool
	SeesInvisible     bool
	CurSpell          byte
	SelKind, SelSize  byte
	SelFlags          byte
	TgtKind, TgtFlags byte
	Delay             byte
	Levels            [8]int
	Seed              uint32
}

type goldenCotabOutcome struct {
	DmgType, Damage, Save     byte
	AffVal                    AffectType
	HitRoll                   int8
	Veto                      byte
	Seed                      uint32
	HP, AC0, AC1, Int, Wis    [3]int
	Side, Control, Auto, Stat [3]int
	Delay, Moves, Turned, Can int
	Affects                   [3]string
	Text                      []string
}

// goldenCotabCase: the affect, scene and spell are cotab's own numbers; Want
// holds the outcome but for the affects, which WantAffVal and WantAffects
// give as cotab's numbers too ([type, time, data] of the three characters').
type goldenCotabCase struct {
	Affect      byte
	Scene       goldenCotabScene
	Want        goldenCotabOutcome
	WantAffVal  byte
	WantAffects [3][][3]int
}

// goldenCotabChar is a character as the scenes built them: human, 20 of 30
// hit points, AC 50/57, one attack each, every score 12.
func goldenCotabChar(name string, kind, size, flags byte, levels [8]int) *Character {
	r := &Character{}
	r.CharName = name
	r.Kind, r.Size, r.CreatureFlags = kind, size, CreatureFlags(flags)
	for i, l := range levels {
		r.Levels[i] = Level(l)
	}
	r.HP, r.MaxHP, r.Okay, r.Race = 20, 30, 1, RaceHuman
	r.AC = [2]ArmourClass{50, 57}
	r.Attacks = [2]Count{1, 1}
	for i := range r.Stats {
		r.Stats[i] = 12
	}
	return r
}

// goldenCotabSetup is the scene s on rules1e: the three characters and the
// affect (cotab's number, 0: none) on the bearer.
func goldenCotabSetup(tb *Tables, s goldenCotabScene, affect byte) (*RuleState, [3]*Character) {
	r := &RuleState{Tables: tb}
	rc := [3]*Character{goldenCotabChar("SEL", s.SelKind, s.SelSize, s.SelFlags, s.Levels), goldenCotabChar("HERO", 0, 1, 0, s.Levels),
		goldenCotabChar("FOE", s.TgtKind, 1, s.TgtFlags, s.Levels)}
	r.Party = rc[:]
	rc[1].Combat = &CombatInfo{Target: rc[2], Delay: s.Delay, Moves: 6, CanCast: 1, CanUse: 1}
	if s.SeesInvisible {
		r.AddAffect(rc[0], CotabAffectIDs[goldenCotabDetectInvisibility], 0, 0, 0)
	}
	r.AddAffect(rc[1], CotabAffectIDs[affect], 0, s.Data, 0)
	r.Rv.DmgType, r.Rv.Damage, r.Rv.AffVal, r.Rv.Save, r.Rv.HitRoll, r.Rv.SaveMade = s.DmgType, int16(s.Damage), CotabAffectIDs[s.AffVal], s.Save, s.HitRoll, s.SaveMade
	r.CurSpell, r.RandSeed = CotabSpellIDs[s.CurSpell], s.Seed
	if s.Combat {
		r.BeginFlag = 5
	}
	return r, rc
}

// goldenCotabOutcomeOf is what the handler or spell left on r.
func goldenCotabOutcomeOf(r *RuleState, rc [3]*Character) goldenCotabOutcome {
	o := goldenCotabOutcome{DmgType: r.Rv.DmgType, Damage: byte(r.Rv.Damage), Save: r.Rv.Save, AffVal: r.Rv.AffVal,
		HitRoll: r.Rv.HitRoll, Veto: r.AffectVeto, Seed: r.RandSeed}
	for i, c := range rc {
		o.HP[i], o.AC0[i], o.AC1[i], o.Int[i], o.Wis[i] = int(c.HP), int(c.AC[0]), int(c.AC[1]), int(c.Stats[StatInt]), int(c.Stats[StatWis])
		o.Side[i], o.Control[i], o.Auto[i], o.Stat[i] = int(c.Side), int(c.Control), int(c.Auto), int(c.Status)
		for _, a := range c.Affects {
			o.Affects[i] += fmt.Sprintf("%#x/%d/%d ", a.Type, a.Time, a.Data)
		}
	}
	ci := rc[1].Combat
	o.Delay, o.Moves, o.Turned, o.Can = int(ci.Delay), int(ci.Moves), int(ci.Turned), int(ci.CanCast)
	for _, f := range r.Fx {
		if n, ok := f.(rules.Narrated); ok {
			o.Text = append(o.Text, n.Text)
		}
	}
	return o
}

// goldenCotabSame tells whether got, through JSON as the golden is, is the
// golden's outcome with its affect numbers (cotab's) mapped to the ids.
func goldenCotabSame(got goldenCotabOutcome, k goldenCotabCase) bool {
	gb, _ := json.Marshal(got)
	var gv goldenCotabOutcome
	json.Unmarshal(gb, &gv)
	want := k.Want
	want.AffVal = CotabAffectIDs[k.WantAffVal]
	for i, list := range k.WantAffects {
		want.Affects[i] = ""
		for _, a := range list {
			want.Affects[i] += fmt.Sprintf("%#x/%d/%d ", CotabAffectIDs[byte(a[0])], a[1], a[2])
		}
	}
	return reflect.DeepEqual(gv, want)
}

func runGoldenCotab(tb *Tables, k goldenCotabCase) goldenCotabOutcome {
	r, rc := goldenCotabSetup(tb, k.Scene, k.Affect)
	r.CallAffect(CotabAffectIDs[k.Affect], rc[1], rc[1].Affects[len(rc[1].Affects)-1], k.Scene.Flag)
	return goldenCotabOutcomeOf(r, rc)
}

// goldenCotabDetectInvisibility is cotab's number of detect invisibility,
// which the attacker may carry (the scenes' selSeesInvisible).
const goldenCotabDetectInvisibility = 24

func TestGoldenAffectsCotab(t *testing.T) {
	dir := filepath.Join("testdata", "cotab")
	var cases []goldenCotabCase
	if err := readGolden(filepath.Join(dir, "affects.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	_ = CoreRaceHuman
	for i, k := range cases {
		if !goldenCotabSame(runGoldenCotab(&tb, k), k) {
			t.Errorf("case %d affect %d differs from the golden", i, k.Affect)
		}
	}
	t.Logf("%d cases", len(cases))
}
