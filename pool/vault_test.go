package pool

import (
	"encoding/binary"
	"gobox/char"
	"gobox/core"
	"gobox/game/play"
	"os"
	"rules1e"
	"slices"
	"testing"
)

// TestVault: the vault opens as the treasure screen over vault.dat: its
// platinum is the party's, its items are taken one by one, and what is left is
// written back when the party leaves (no "treasure left" question).
func TestVault(t *testing.T) {
	e := quickStarted(t)
	e.selectMember(0)
	st := e.state()
	items := []char.Item{*st.RandomItem(rules1e.ItemJavelin), *st.RandomItem(rules1e.ItemArrow)}
	head := make([]byte, vaultCoins*4)
	binary.LittleEndian.PutUint32(head, 3) // platinum
	path := e.savePath(vaultFile)
	if err := os.WriteFile(path, slices.Concat(head, char.WriteItems(items)), 0o644); err != nil {
		t.Fatal(err)
	}
	gold := st.Purse.Gold
	e.push(cont{kind: kNop})
	e.Start(e.Vault)
	if l, ok := e.State().(play.Looting); !ok || len(l.Items) != 2 {
		t.Fatalf("state %T: want the treasure screen with the vault's 2 items", e.State())
	}
	if st.Purse.Gold <= gold {
		t.Errorf("the vault's platinum was not banked: gold %d, was %d", st.Purse.Gold, gold)
	}
	if _, err := e.Do(play.Take{Item: e.goods()[0]}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(play.Leave{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := e.State().(play.Looting); ok {
		t.Fatal("still in the treasure screen after Leave")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if left := char.ReadItems(b[vaultCoins*4:]); len(left) != 1 || left[0] != items[0] {
		t.Errorf("vault.dat holds %d items, want the one not taken (the first, as the second led)", len(left))
	}
	_ = core.NoItem
}
