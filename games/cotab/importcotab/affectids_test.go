package importcotab

import (
	"fmt"
	"path/filepath"
	"testing"

	"gobox/dax"
	"rules1e"
)

// TestAffectIDsCoverData: every affect number the monsters' affects (MON
// SPC) use is in the game's affectIds. Items' +0x3E is an item property
// that also indexes the affect handlers: not covered yet (fold T4b note).
func TestAffectIDsCoverData(t *testing.T) {
	l, err := dax.NewLoader(filepath.Join("..", "..", "..", "data", "cotab", "sources", "dos"))
	if err != nil {
		t.Skip("no originals:", err)
	}
	var affects []rules1e.AffectFile
	for set := range 10 {
		for id := range 256 {
			affects = append(affects, dax.Records[rules1e.AffectFile](l.BlockShared(fmt.Sprintf("MON%dSPC", set), id))...)
		}
	}
	if len(affects) == 0 {
		t.Fatal("read no affects")
	}
	if miss := rules1e.CotabAffectIDs.Missing(affects); len(miss) > 0 {
		t.Fatalf("affect numbers not in affectIds: %v", miss)
	}
}

// TestItemPowersCoverData: every item power code (+0x3E above 0x7F) of the
// game's items (ITEM0..9, MON ITM) is in its itemPowers.
func TestItemPowersCoverData(t *testing.T) {
	l, err := dax.NewLoader(filepath.Join("..", "..", "..", "data", "cotab", "sources", "dos"))
	if err != nil {
		t.Skip("no originals:", err)
	}
	var items []rules1e.ItemFile
	for set := range 10 {
		for id := range 256 {
			items = append(items, dax.Records[rules1e.ItemFile](l.BlockShared(fmt.Sprintf("MON%dITM", set), id))...)
		}
		for _, its := range dax.ItemBlocks[rules1e.ItemFile](l, fmt.Sprintf("ITEM%d", set)) {
			items = append(items, its...)
		}
	}
	if len(items) == 0 {
		t.Fatal("read no items")
	}
	if miss := rules1e.CotabItemPowers.Missing(items); len(miss) > 0 {
		t.Fatalf("item power codes not in itemPowers: %v", miss)
	}
}
