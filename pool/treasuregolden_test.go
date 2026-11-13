package pool

import (
	"fmt"
	"os"
	"gobox/rules1e"
	"strings"
	"testing"
)

// porTreasureLines are POR's random items on fixed seeds: 4000 of the type
// roll and the item made for it, and every item type 8 times, each with the
// RNG state after.
func porTreasureLines(t *testing.T) []string {
	e := testEngine(t)
	s := e.state()
	line := func(kind string, seed uint32, it *rules1e.Item) string {
		return fmt.Sprintf("%s %d: type %d words %v plus %d/%d hidden %d weight %d count %d price %d spells %v seed %d",
			kind, seed, it.Type, it.Words, it.Plus, it.Plus2, it.Hidden, it.Weight, it.Count, it.Price, it.Spells, s.RandSeed)
	}
	var out []string
	for seed := uint32(1); seed <= 4000; seed++ {
		s.RandSeed, s.Treasure = seed, nil
		if err := s.AddRandomTreasure(1); err != nil {
			t.Fatal(err)
		}
		out = append(out, line("roll", seed, s.Treasure[0]))
	}
	for typ := range 256 {
		for k := uint32(1); k <= 8; k++ {
			s.RandSeed = uint32(typ)*8 + k
			out = append(out, line(fmt.Sprint("type", typ), s.RandSeed, s.RandomItem(rules1e.ItemType(typ))))
		}
	}
	return out
}

// TestPorTreasureGolden: POR's random items equal testdata/treasure-por.golden,
// the outcomes of the original RandomItemType and RandomItem code (CT_GOLDEN=1
// rewrites it).
func TestPorTreasureGolden(t *testing.T) {
	got := strings.Join(porTreasureLines(t), "\n") + "\n"
	const file = "../rules1e/testdata/treasure-por.golden"
	if os.Getenv("CT_GOLDEN") != "" {
		if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	g, w := strings.Split(got, "\n"), strings.Split(string(want), "\n")
	for i := range w {
		if i >= len(g) || g[i] != w[i] {
			t.Fatalf("line %d:\n got  %s\n want %s", i+1, g[min(i, len(g)-1)], w[i])
		}
	}
	t.Logf("%d POR treasure outcomes compared", len(w)-1)
}
