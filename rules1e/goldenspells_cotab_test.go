package rules1e

import (
	"path/filepath"
	"testing"
)

// TestGoldenSpellsCotab: rules1e's effects of the spells cotab has first
// (66, 68, 72, 95, 96, 97, 100) and spell 98's cone over a bare combat map
// leave what cotab's own spell effects left (captured before its rules/
// went), over the 48 scenes of the affect goldens: rule globals, the three
// characters, affects, messages and dice. Numbers in the file are cotab's.

type goldenCotabSpellCase struct {
	goldenCotabCase
	Kind     string // "first": one foe; "cone98": the cone on a bare map
	Spell    byte
	NTargets int
}

func TestGoldenSpellsCotab(t *testing.T) {
	dir := filepath.Join("testdata", "cotab")
	var cases []goldenCotabSpellCase
	if err := readGolden(filepath.Join(dir, "spells.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for i, k := range cases {
		r, rc := goldenCotabSetup(&tb, k.Scene, 0)
		id := CotabSpellIDs[k.Spell]
		switch k.Kind {
		case "first":
			rc[0].Combat = &CombatInfo{Target: rc[2], Moves: 6, CanCast: 1, CanUse: 1}
			rc[2].HP, rc[2].MaxHP = 120, 120 // no kill: no combat map here
			r.SetTgt(1, rc[2])
			r.NSpellTargets = 1
		case "cone98":
			r.Cmap = &CombatMap{}
			for y := range r.Cmap.Ground {
				for x := range r.Cmap.Ground[y] {
					r.Cmap.Ground[y][x] = PorGroundIDs.Floor
				}
			}
			places := [][2]int{{10, 10}, {12, 10}, {13, 10 + k.Scene.Levels[ClsMage]%2}} // cotabScenes' index parity: its mage level is j, the index mod 6
			r.CombatantSlots, r.Ncomb = r.Party, len(r.Party)
			for j, p := range places {
				r.Cpos[j+1] = CombatPos{X: byte(p[0]), Y: byte(p[1]), Size: 1}
			}
			r.CombatRedrawMap()
			for _, c := range rc[1:] {
				c.HP, c.MaxHP = 120, 120
			}
			r.TgtX, r.TgtY = 11, 10
		default:
			t.Fatalf("case %d: kind %q", i, k.Kind)
		}
		r.SpellEffect(id)
		if !goldenCotabSame(goldenCotabOutcomeOf(r, rc), k.goldenCotabCase) || k.Kind == "cone98" && (r.NSpellTargets != k.NTargets || k.NTargets == 0) {
			t.Errorf("case %d (%s spell %d) differs from the golden (%d targets, golden %d)", i, k.Kind, k.Spell, r.NSpellTargets, k.NTargets)
		}
	}
	t.Logf("%d cases", len(cases))
}
