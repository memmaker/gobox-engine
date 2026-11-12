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
	if st.Purse.Gold != gold {
		t.Errorf("the vault's platinum was banked on opening: gold %d, was %d", st.Purse.Gold, gold)
	}
	if _, err := e.Do(play.Take{Item: e.goods()[0]}); err != nil {
		t.Fatal(err)
	}
	mine := core.NoItem
	for _, h := range e.listedItems(e.selected()) {
		if !e.itemCard(e.selected(), h).Readied {
			mine = h
			break
		}
	}
	if mine == core.NoItem {
		t.Fatal("the member has no unreadied item to deposit")
	}
	if _, err := e.Do(play.Drop{Member: 0, Item: mine}); err != nil { // Deposit
		t.Fatal(err)
	}
	if len(e.goods()) != 2 {
		t.Fatalf("%d items in the vault after the deposit, want the one left and the deposit", len(e.goods()))
	}
	if _, err := e.Do(play.Leave{}); err != nil { // the money is left unbanked
		t.Fatal(err)
	}
	if _, ok := e.State().(play.Looting); ok {
		t.Fatal("still in the treasure screen after Leave")
	}
	if st.Purse.Gold != gold {
		t.Errorf("leaving banked the vault's platinum: gold %d, was %d", st.Purse.Gold, gold)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := binary.LittleEndian.Uint32(b); got != 3 {
		t.Errorf("vault.dat holds %d platinum, want the 3 left", got)
	}
	// the taken item was the member's first, so it is the one deposited, last in the file
	left := char.ReadItems(b[vaultCoins*4:]) // the shown name is not the record's: compare types
	if len(left) != 2 || left[0].Type != items[1].Type || left[1].Type != items[0].Type {
		t.Errorf("vault.dat holds %v, want the arrow not taken, then the javelin deposited", left)
	}
	e.push(cont{kind: kNop})
	e.Start(e.Vault) // reopened: the money is still there, and Take banks it
	if _, err := e.Do(play.Take{Item: core.NoItem}); err != nil {
		t.Fatal(err)
	}
	if st.Purse.Gold <= gold {
		t.Errorf("Take on the money line did not bank the vault's platinum: gold %d, was %d", st.Purse.Gold, gold)
	}
}
