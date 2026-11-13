package rules1e

import (
	"path/filepath"
	"reflect"
	"testing"

	"gobox/rules"
)

// TestGoldenSpellsPod: rules1e's ports of POD's own spells (128..145 and the
// shared ones) leave what POD's spell effects left (captured before its
// rules/ went): messages, each target's hit points, status and affects, and
// the dice. Spell and affect numbers in the file are POD's own.

type goldenPodSpellTarget struct {
	HP, MaxHP, MaxLevel, Size, Side, Status, Race, Con, MonsterFlags, Affect byte
}

type goldenPodSpellCase struct {
	Spell, Level byte
	Scene        int
	Seed         uint32
	WantSeed     uint32
	Said         []string
	Targets      []struct {
		HP, Status, Okay byte
		Affects          [][4]int // type (POD's number), time, data, call
	}
}

type goldenPodSpells struct {
	Scenes [][]goldenPodSpellTarget
	Cases  []goldenPodSpellCase
}

func TestGoldenSpellsPod(t *testing.T) {
	dir := filepath.Join("testdata", "pod")
	var g goldenPodSpells
	if err := readGolden(filepath.Join(dir, "spells.golden.json.gz"), &g); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for _, k := range g.Cases {
		scene := g.Scenes[k.Scene]
		id := PodSpellIDs[k.Spell]
		r := &RuleState{Tables: &tb}
		caster := &Character{Combat: &CombatInfo{}}
		caster.Levels[ClsCleric], caster.Levels[ClsMage], caster.Okay = Level(k.Level), Level(k.Level), 1
		r.Party = []*Character{caster}
		r.CurSpell, r.RandSeed = id, k.Seed
		var ts []*Character
		for i, s := range scene {
			c := &Character{Combat: &CombatInfo{}}
			c.HP, c.MaxHP, c.MaxLevel, c.Size, c.Side, c.Status, c.Race = s.HP, s.MaxHP, s.MaxLevel, s.Size, s.Side, s.Status, RaceID(s.Race)
			c.Stats[StatCon] = Score(s.Con)
			c.MonsterFlags = MonsterFlags(s.MonsterFlags)
			c.AC = [2]ArmourClass{50, 50}
			if s.Status == goldenPodStatusOkay {
				c.Okay = 1
			}
			if s.Affect != 0 {
				c.Affects = []*Affect{{AffectRecord: AffectRecord{Type: podAffect(s.Affect), Time: 5}}}
			}
			ts = append(ts, c)
			r.SetTgt(i+1, c)
		}
		r.NSpellTargets = len(scene)
		r.SpellEffect(id)
		var said []string
		for _, f := range r.Fx {
			if n, ok := f.(rules.Narrated); ok {
				said = append(said, n.Text)
			}
		}
		fail := r.RandSeed != k.WantSeed || !reflect.DeepEqual(said, k.Said)
		for i, c := range ts {
			w := k.Targets[i]
			var ga, wa []AffectRecord
			for _, a := range c.Affects {
				ga = append(ga, a.AffectRecord)
			}
			for _, a := range w.Affects {
				wa = append(wa, AffectRecord{Type: podAffect(byte(a[0])), Time: uint16(a[1]), Data: byte(a[2]), Call: byte(a[3])})
			}
			if c.HP != w.HP || byte(c.Status) != w.Status || c.Okay != w.Okay || !reflect.DeepEqual(ga, wa) {
				fail = true
			}
		}
		if fail {
			t.Fatalf("spell %d scene %d level %d seed %d differs from the golden", k.Spell, k.Scene, k.Level, k.Seed)
		}
	}
	t.Logf("%d cases", len(g.Cases))
}

const goldenPodStatusOkay = 0 // a character with no status
