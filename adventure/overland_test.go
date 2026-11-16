package adventure

import "testing"

func TestOverlandAndNoRemains(t *testing.T) {
	g := GameTables{Overland: []OverlandBlock{{Block: 26, ViewMode: 3, Finale: true}}, NoRemainsBlocks: []int{10}}
	if o := g.OverlandBlock(26); o == nil || o.ViewMode != 3 || !o.Finale {
		t.Fatalf("block 26: %v", o)
	}
	if g.OverlandBlock(25) != nil || (GameTables{}).OverlandBlock(26) != nil {
		t.Fatal("an unlisted block is not overland")
	}
	if !g.NoRemains(10) || g.NoRemains(11) {
		t.Fatal("NoRemains")
	}
}
