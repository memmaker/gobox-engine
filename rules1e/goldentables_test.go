package rules1e

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gobox/core"
	"gobox/core/dnd"
)

// The golden tables: for each of the three imported games (sotsb, pod,
// cotab) testdata/<game>/ holds the game's rules.json and layout.json as its
// import wrote them, and tables.golden.json, every value rules1e's table
// accessors give over them. The values were captured (HX_GOLDEN_FROM=<folder>
// go test -run TestGoldenTables -update) after each game's own oracle tests
// (tables1e, slots1e, spells1e: its literals against these accessors) passed,
// so they are the game's values; the test needs no game code.
var updateGolden = flag.Bool("update", false, "rewrite testdata/<game>/tables.golden.json from HX_GOLDEN_FROM/<game>")

var goldenGames = []string{"sotsb", "pod", "cotab"}

var (
	goldenClasses = []core.Class{dnd.Cleric, dnd.Druid, dnd.Fighter, dnd.Paladin, dnd.Ranger, dnd.Mage, dnd.Thief, dnd.Monk}
	goldenRaces   = []core.Race{dnd.Monster, dnd.Dwarf, dnd.Elf, dnd.Gnome, dnd.HalfElf, dnd.Halfling, dnd.HalfOrc, dnd.Human}
	goldenSaves   = []core.Save{dnd.SaveParalysis, dnd.SavePetrify, dnd.SaveRod, dnd.SaveBreath, dnd.SaveSpell}
	goldenSkills  = thiefSkillOrder
)

const (
	goldenLevels  = 42 // class levels 0..41 (xpTop)
	goldenScores  = 19 // ability scores 0..18
	goldenCareers = 7  // single classes, without the monk (the age rows' columns)
)

// goldenTables is every value read through Tables' accessors.
type goldenTables struct {
	Thac0     map[core.Class][]int
	Saves     map[core.Class][][]int // by level, by save
	XP        map[core.Class][]int32 // levels 2..41
	HitDice   map[core.Class][2]int
	Cleric    [][]int // slots by class level, slot level 1..
	Mage      [][]int
	ThiefBase [][]int // by level, by skill
	ThiefRace map[core.Race][]int
	ThiefDex  [][]int
	Ability   map[core.Race][][2]int // strength, then abilities 1..5
	StrCap    map[core.Race]int
	Careers   map[core.Race][]Career
	Ages      map[core.Race][][3]int // by career
	AgeSteps  map[core.Race][]int
	CareerMin [][]int
	CareerAl  [][]AlignmentNumber
	ConHP     []int8
	Paladin   [][]int16
	Ranger    [][]int16
	MonkOld   [][]int16
	Spells    [][]SpellProp
	Stats     map[core.Class][][][3][]int // by level 1..SlotLevels, by score, cleric/druid/mage slots by level
}

func dumpTables(tb *Tables) goldenTables {
	g := goldenTables{Thac0: map[core.Class][]int{}, Saves: map[core.Class][][]int{}, XP: map[core.Class][]int32{},
		HitDice: map[core.Class][2]int{}, ThiefRace: map[core.Race][]int{}, Ability: map[core.Race][][2]int{}, StrCap: map[core.Race]int{},
		Careers: map[core.Race][]Career{}, Ages: map[core.Race][][3]int{}, AgeSteps: map[core.Race][]int{},
		Stats: map[core.Class][][][3][]int{}, ConHP: tb.ConHPByScore, Paladin: tb.PaladinSpells, Ranger: tb.RangerSpells,
		MonkOld: tb.MonkOldSaves, Spells: tb.Spells}
	for k, class := range goldenClasses {
		if k >= len(tb.Thac0) {
			continue // no monk row (SotSB, POD)
		}
		for lvl := range goldenLevels {
			g.Thac0[class] = append(g.Thac0[class], tb.Thac0At(class, lvl))
		}
		for lvl := range tb.SaveRows {
			var row []int
			for _, s := range goldenSaves {
				row = append(row, tb.SaveAt(class, lvl, s))
			}
			g.Saves[class] = append(g.Saves[class], row)
		}
		for lvl := 2; lvl < goldenLevels; lvl++ {
			g.XP[class] = append(g.XP[class], tb.XPForLevel(class, lvl))
		}
		n, s := tb.TrainingDie(class)
		g.HitDice[class] = [2]int{n, s}
	}
	for lvl := 2; lvl <= tb.SlotLevels; lvl++ {
		cl, mu := tb.ClericSlots(lvl), tb.MageSlots(lvl)
		g.Cleric, g.Mage = append(g.Cleric, cl[:]), append(g.Mage, mu[:])
	}
	for lvl := range tb.ThiefByLevel {
		var row []int
		for _, s := range goldenSkills {
			row = append(row, tb.ThiefBase(lvl, s))
		}
		g.ThiefBase = append(g.ThiefBase, row)
	}
	for dex := range tb.ThiefByDex {
		var row []int
		for _, s := range goldenSkills[:len(tb.ThiefByDex[dex])] {
			row = append(row, tb.ThiefDex(dex, s))
		}
		g.ThiefDex = append(g.ThiefDex, row)
	}
	races := goldenRaces
	if tb.Races != nil { // the game's own rows (SotSB, POD: no half-orc)
		races = nil
		for _, r := range tb.Races {
			if r != "" {
				races = append(races, r)
			}
		}
	}
	for _, race := range races {
		for _, s := range goldenSkills {
			g.ThiefRace[race] = append(g.ThiefRace[race], tb.ThiefRace(race, s))
		}
		for a := Strength; a <= Charisma; a++ {
			lo, hi := tb.AbilityRange(race, a)
			g.Ability[race] = append(g.Ability[race], [2]int{lo, hi})
		}
		g.StrCap[race] = tb.ExceptionalStrengthCap(race)
		if race != dnd.Monster { // its career row is no list
			g.Careers[race] = tb.RaceCareers(race)
		}
		for c := range goldenCareers {
			b, n, s := tb.StartingAge(race, Career(c))
			g.Ages[race] = append(g.Ages[race], [3]int{b, n, s})
		}
		for step := 1; step <= 5; step++ {
			g.AgeSteps[race] = append(g.AgeSteps[race], tb.AgeStep(race, step))
		}
	}
	for c := range tb.ClassMinStats {
		var row []int
		for a := range 6 {
			row = append(row, tb.CareerMinimum(Career(c), Ability(a)))
		}
		g.CareerMin = append(g.CareerMin, row)
		g.CareerAl = append(g.CareerAl, tb.CareerAlignments(Career(c)))
	}
	for _, class := range []core.Class{dnd.Cleric, dnd.Druid, dnd.Paladin, dnd.Ranger, dnd.Mage} {
		for lvl := 1; lvl <= tb.SlotLevels; lvl++ {
			var byScore [][3][]int
			for score := 3; score < goldenScores; score++ {
				s := tb.ClassStats(Who{Levels: map[core.Class]int{class: lvl}, Race: dnd.Human, Wis: score, Int: score})
				byScore = append(byScore, [3][]int{s.ClericSlots[:], s.DruidSlots[:], s.MageSlots[:]})
			}
			g.Stats[class] = append(g.Stats[class], byScore)
		}
	}
	return g
}

// goldenOwned is what rules1e itself owns of a game's rules.json (the
// importer writes it from these): the ids, variants, groups and rules. The
// golden's rules.json keeps what the game's executable holds; these keys are
// laid over it from rules1e as they are now, so a renamed handler or a
// renumbered id does not stale the golden.
var goldenOwned = map[string]map[string]any{
	"sotsb": {"affectIds": SotsbAffectIDs, "spellIds": SotsbSpellIDs, "itemSpellCode": SevenBitItemSpell, "spellClasses": SotsbSpellClasses,
		"affectVariants": SotsbAffectVariants, "affectGroups": SotsbAffectGroups, "giantRule": LaterGiants, "oldAge": SotsbOldAge,
		"casterLevel": SotsbCasterLevel, "itemPowers": SotsbItemPowers},
	"pod": {"affectIds": PodAffectIDs, "spellIds": PodSpellIDs, "itemSpellCode": SevenBitItemSpell, "spellClasses": PodSpellClasses,
		"affectVariants": PodAffectVariants, "affectGroups": PodAffectGroups, "giantRule": LaterGiants, "oldAge": PodOldAge,
		"casterLevel": PodCasterLevel, "affectRule": PodAffectRule, "itemPowers": PodItemPowers},
	"cotab": {"affectIds": CotabAffectIDs, "spellIds": CotabSpellIDs, "spellClasses": CotabSpellClasses,
		"affectVariants": CotabAffectVariants, "affectGroups": CotabAffectGroups, "giantRule": LaterGiants,
		"casterLevel": CotabCasterLevel, "affectRule": CotabAffectRule, "itemPowers": CotabItemPowers},
}

func loadGoldenTables(t *testing.T, dir string) Tables {
	t.Helper()
	l, err := ReadLayout(filepath.Join(dir, LayoutFile))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for k, v := range goldenOwned[filepath.Base(dir)] { // rules1e's, as they are now, over the golden's
		if m[k], err = json.Marshal(v); err != nil {
			t.Fatal(err)
		}
	}
	if b, err = json.Marshal(m); err != nil {
		t.Fatal(err)
	}
	tb, err := POR.WithRules(bytes.NewReader(b), l)
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

func TestGoldenTables(t *testing.T) {
	for _, game := range goldenGames {
		t.Run(game, func(t *testing.T) {
			dir := filepath.Join("testdata", game)
			if *updateGolden {
				src := filepath.Join(os.Getenv("HX_GOLDEN_FROM"), game, game, "data")
				for _, f := range []string{"rules.json", LayoutFile} {
					b, err := os.ReadFile(filepath.Join(src, f))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.MkdirAll(dir, 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, f), b, 0o644); err != nil {
						t.Fatal(err)
					}
				}
				tb := loadGoldenTables(t, dir)
				b, err := json.MarshalIndent(dumpTables(&tb), "", " ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "tables.golden.json"), b, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			b, err := os.ReadFile(filepath.Join(dir, "tables.golden.json"))
			if err != nil {
				t.Fatal(err)
			}
			var want goldenTables
			if err := json.Unmarshal(b, &want); err != nil {
				t.Fatal(err)
			}
			tb := loadGoldenTables(t, dir)
			got := dumpTables(&tb)
			gb, _ := json.Marshal(got)
			wb, _ := json.Marshal(want)
			var gv, wv goldenTables
			json.Unmarshal(gb, &gv)
			json.Unmarshal(wb, &wv)
			if !reflect.DeepEqual(gv, wv) {
				t.Errorf("tables differ from the golden (the value dumps %d and %d bytes)", len(gb), len(wb))
			}
		})
	}
}

// readGolden decodes the gzipped JSON golden file at path into v.
func readGolden(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	return json.NewDecoder(z).Decode(v)
}
