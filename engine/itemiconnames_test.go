package engine

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"gobox/asset"
	"gobox/rules"
)

// TestNamedItemsHaveIconRecords: the valuables of POR's treasure 39 (Gold
// Banded Wand, Jewel Studded Bowl, Jewelled Dragon Statuette) are published
// identified, with their kind and name words, and the pool's items.rec has
// a record of its own for each ("<kind>/<name>"): the remapper maps each alone.
func TestNamedItemsHaveIconRecords(t *testing.T) {
	e := quickStarted(t)
	e.set = 4 // POR's ITEM4 holds block 39
	e.do(rules.AddTreasureBlock{Block: 39})
	_, rec, err := asset.ItemIcons("pool")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, h := range e.goods() {
		it := gameItem(h, e.itemCard(rules.NoChar, h))
		var w []string
		for _, s := range slices.Backward(it.Words) {
			if s = strings.TrimSpace(s); s != "" {
				w = append(w, s)
			}
		}
		name := strings.Join(w, " ")
		if !slices.Contains([]string{"Gold Banded Wand", "Jewel Studded Bowl", "Jewelled Dragon Statuette"}, name) {
			continue
		}
		found++
		if !it.Identified {
			t.Errorf("%s is not identified: it shows its kind's icon", name)
		}
		if key := fmt.Sprintf("internal_name: %d/%s\n", it.Base, name); !strings.Contains(rec, key) {
			t.Errorf("items.rec has no record %q", strings.TrimSpace(key))
		}
	}
	if found != 3 {
		t.Fatalf("%d of the 3 valuables in treasure 39: %+v", found, e.goods())
	}
}
