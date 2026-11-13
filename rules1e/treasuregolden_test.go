package rules1e

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// treasureLines are a game's random items as the game's file holds them, on
// fixed seeds: 4000 of the type roll and the item made for it, and every
// item type 8 times, each with the RNG state after.
func treasureLines(roll func(e *RuleState) (ItemType, Item), direct func(e *RuleState, typ ItemType) Item) []string {
	e := &RuleState{}
	line := func(kind string, seed uint32, it Item) string {
		return fmt.Sprintf("%s %d: type %d words %v plus %d/%d hidden %d weight %d count %d price %d spells %v seed %d",
			kind, seed, it.Type, it.Words, it.Plus, it.Plus2, it.Hidden, it.Weight, it.Count, it.Price, it.Spells, e.RandSeed)
	}
	var out []string
	for seed := uint32(1); seed <= 4000; seed++ {
		e.RandSeed = seed
		_, it := roll(e)
		out = append(out, line("roll", seed, it))
	}
	for typ := range 256 {
		for k := uint32(1); k <= 8; k++ {
			e.RandSeed = uint32(typ)*8 + k
			out = append(out, line(fmt.Sprint("type", typ), e.RandSeed, direct(e, ItemType(typ))))
		}
	}
	return out
}

// TestTreasureGolden: COTAB's and SotSB's (POR's: pool TestPorTreasureGolden, over the imported data) random items, made by the tables
// in treasure/*.json, equal testdata/treasure-<game>.golden, the outcomes of
// the code they replaced (cotab-re's Pascal as transcribed, SotsbTreasure's
// Go builder), frozen when that code went.
func TestTreasureGolden(t *testing.T) {
	games := []struct {
		name string
		data func() *TreasureRolls
	}{{"cotab", CotabTreasure}, {"sotsb", SotsbTreasure}}
	for _, g := range games {
		tr := g.data()
		got := strings.Join(treasureLines(
			func(e *RuleState) (ItemType, Item) { ty := rollOf(e, tr.Type); return ty, e.treasureRecord(tr, ty) },
			func(e *RuleState, typ ItemType) Item { return e.treasureRecord(tr, typ) }), "\n") + "\n"
		file := "testdata/treasure-" + g.name + ".golden"
		if os.Getenv("TREASURE_GOLDEN_WRITE") != "" { // regenerate after a deliberate table change
			f, err := os.Create(file)
			if err == nil {
				_, err = f.WriteString(got)
				err = errors.Join(err, f.Close())
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		if len(gl) != len(wl) {
			t.Fatalf("%s: %d lines, golden %d", g.name, len(gl), len(wl))
		}
		for i := range wl {
			if gl[i] != wl[i] {
				t.Fatalf("%s line %d:\n got  %s\n want %s", g.name, i+1, gl[i], wl[i])
			}
		}
		t.Logf("%s: %d treasure outcomes compared", g.name, len(wl)-1)
	}
}
