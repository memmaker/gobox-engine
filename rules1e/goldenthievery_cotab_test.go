package rules1e

import (
	"path/filepath"
	"testing"
)

// TestGoldenThieveryCotab: readying an item of thievery (alone, or beside
// gauntlets of free action) gives a thief of each level and dexterity the
// thief skills cotab's own item code gave (captured before its rules/ went).

const (
	goldenCotabThievery   = 128 + 11 // cotab's item power code: thievery
	goldenCotabFreeAction = 128 + 2  // free action
)

type goldenThieveryCase struct {
	Other, Level, Dex int
	Thief             []int
}

func TestGoldenThieveryCotab(t *testing.T) {
	dir := filepath.Join("testdata", "cotab")
	var cases []goldenThieveryCase
	if err := readGolden(filepath.Join(dir, "thievery.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	for _, k := range cases {
		var levels [8]int
		levels[ClsThief] = k.Level
		c := goldenCotabChar("THIEF", 0, 1, 0, levels)
		c.Stats[StatDex] = Score(k.Dex)
		r := &RuleState{Tables: &tb, Party: []*Character{c}}
		item := func(code byte) *Item {
			x := &Item{Readied: 1}
			x.Spells[2] = CotabItemPowers[code]
			return x
		}
		it := item(goldenCotabThievery)
		c.Items = []*Item{it}
		if k.Other != 0 {
			c.Items = []*Item{item(byte(k.Other)), it}
		}
		r.CallAffect(AffectType(it.Spells[2]), c, &Affect{Item: it}, 0)
		for i, want := range k.Thief {
			if int(c.Thief[i]) != want {
				t.Fatalf("other %d level %d dex %d: skill %d is %d, golden %d", k.Other, k.Level, k.Dex, i, c.Thief[i], want)
			}
		}
	}
	t.Logf("%d cases", len(cases))
}
