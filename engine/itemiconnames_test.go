package engine

import (
	"fmt"
	"strings"
	"testing"

	"gobox/asset"
	"gobox/rules"
	"gobox/rules1e"
)

// TestNamedItemsHaveIconRecords: the valuables of POR's treasure 39 (Gold
// Banded Wand, Jewel Studded Bowl, Jewelled Dragon Statuette) are published
// with their own ids (rules1e's named items), not their shared kind's, and
// the pool's items.rec has a record for each: the remapper maps each alone.
// Unidentified, an item shows its kind's.
func TestNamedItemsHaveIconRecords(t *testing.T) {
	e := quickStarted(t)
	e.set = 4 // POR's ITEM4 holds block 39
	e.do(rules.AddTreasureBlock{Block: 39})
	_, rec, err := asset.ItemIcons("pool")
	if err != nil {
		t.Fatal(err)
	}
	named := map[rules.ItemID]bool{}
	for _, n := range rules1e.PorNamedItems {
		named[n.ID] = true
	}
	found := 0
	for _, h := range e.goods() {
		it := gameItem(h, e.itemCard(rules.NoChar, h))
		if !named[it.Base] {
			continue
		}
		found++
		if !strings.Contains(rec, fmt.Sprintf("internal_name: %d\n", it.Base)) {
			t.Errorf("items.rec has no record for %s (id %d)", it.Name, it.Base)
		}
		held := e.heldOf(h)
		kind := e.state().Kinds()[held.Type]
		if it.Base == kind {
			t.Errorf("%s shows its kind's id %d, not its own", it.Name, kind)
		}
	}
	if found != 3 {
		t.Fatalf("%d of the 3 valuables of treasure 39 have an id of their own", found)
	}
}

// TestNamedItemTables: every game's named items have ids from 1001, no
// words twice in a game, and each id has its record in the pool's items.rec.
func TestNamedItemTables(t *testing.T) {
	_, rec, err := asset.ItemIcons("pool")
	if err != nil {
		t.Fatal(err)
	}
	for game, l := range map[string][]rules1e.NamedItem{"por": rules1e.PorNamedItems, "cotab": rules1e.CotabNamedItems, "sotsb": rules1e.SotsbNamedItems, "pod": rules1e.PodNamedItems} {
		seen := map[[3]rules1e.WordID]bool{}
		for _, n := range l {
			if n.ID < 1001 || seen[n.Words] {
				t.Errorf("%s: %v id %d: below 1001 or its words twice", game, n.Words, n.ID)
			}
			seen[n.Words] = true
			if !strings.Contains(rec, fmt.Sprintf("internal_name: %d\n", n.ID)) {
				t.Errorf("%s: items.rec has no record for id %d", game, n.ID)
			}
		}
		if len(l) == 0 {
			t.Errorf("%s: no named items", game)
		}
	}
}

// TestUnidentifiedShowsKind: a named item not yet identified is published
// with its kind's id: its icon never shows what it is.
func TestUnidentifiedShowsKind(t *testing.T) {
	e := quickStarted(t)
	e.set = 4
	e.do(rules.AddTreasureBlock{Block: 39})
	for _, h := range e.goods() {
		held := e.heldOf(h)
		if e.itemCard(rules.NoChar, h).Kind == e.state().Kinds()[held.Type] {
			continue // no named item
		}
		held.Hidden = 7 // every word hidden
		if c := e.itemCard(rules.NoChar, h); c.Identified || c.Kind != e.state().Kinds()[held.Type] {
			t.Fatalf("%s unidentified: identified %v, id %d, want its kind's %d", c.Name, c.Identified, c.Kind, e.state().Kinds()[held.Type])
		}
		return
	}
	t.Fatal("no named item in treasure 39")
}
