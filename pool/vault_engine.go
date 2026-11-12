package pool

import (
	"encoding/binary"
	"fmt"
	"gobox/char"
	"gobox/core"
	"math"
	"os"
	"slices"
)

// The vault (pod O046/O06E:30B1): what the party leaves behind stays in vault.dat
// in the save directory: the pool's platinum, gems and jewelry as three
// longints, then 63-byte item records. The vault screen is the treasure screen
// (PendLoot) over those items; the money is banked on opening and the items
// not taken are written back on leaving.
// ponytail: no Deposit (the items menu's Drop) and no money left in the vault;
// both need rules (a treasure the party can add to and read back as records).

const (
	vaultFile  = "vault.dat"
	vaultBlock = -1 // treasure's block number for the vault's items (AddTreasureBlock)
	// the three longints of the file, and their place among core.SetLoot's seven coins
	vaultCoins                          = 3
	coinPlatinum, coinGems, coinJewelry = 4, 5, 6
)

// vault is the open vault: its item records, by the handle the treasure gave each.
type vault struct {
	items []char.Item
	byID  map[core.ItemHandle]char.Item
}

// Vault = pod's vault screen: the treasure screen over vault.dat until the party leaves.
func (e *Engine) Vault() {
	coins, items, err := readVault(e.savePath(vaultFile))
	if err != nil {
		panic(fmt.Sprintf("vault: %v", err))
	}
	e.vault = &vault{items: items, byID: map[core.ItemHandle]char.Item{}}
	e.do(core.ClearLoot{})
	for coins != [vaultCoins]uint64{} { // SetLoot holds 16 bits a kind: bank it in turns
		var c [7]uint16
		for i, k := range []int{coinPlatinum, coinGems, coinJewelry} {
			n := uint16(min(coins[i], math.MaxUint16))
			c[k], coins[i] = n, coins[i]-uint64(n)
		}
		e.do(core.SetLoot{Coins: c})
		e.do(core.BankLoot{})
	}
	if len(items) > 0 {
		e.do(core.ClearTreasure{})
		e.do(core.AddTreasureBlock{Block: vaultBlock})
		goods := e.goods()
		if len(goods) != len(items) {
			panic(fmt.Sprintf("vault: %d items listed of %d", len(goods), len(items)))
		}
		for i, h := range goods { // the block's last item leads (AddTreasureBlock)
			e.vault.byID[h] = items[len(items)-1-i]
		}
	}
	e.push(cont{kind: kVaultDone})
	e.lootMenu()
}

// vaultDone: the vault is left; the items not taken are written back.
func (e *Engine) vaultDone() {
	var left []char.Item
	for _, h := range e.goods() {
		left = append(left, e.vault.byID[h])
	}
	e.vault = nil
	if err := writeVault(e.savePath(vaultFile), left); err != nil {
		e.say("Can't save the vault: " + err.Error())
	}
	e.do(core.ClearTreasure{})
	e.afterEnd()
}

// readVault is vault.dat's pool and items; a missing file is an empty vault.
func readVault(path string) ([vaultCoins]uint64, []char.Item, error) {
	var coins [vaultCoins]uint64
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return coins, nil, nil
	}
	if err != nil {
		return coins, nil, err
	}
	head := vaultCoins * 4
	if len(b) < head {
		return coins, nil, fmt.Errorf("%s: %d bytes, the pool needs %d", path, len(b), head)
	}
	for i := range coins {
		coins[i] = uint64(max(int32(binary.LittleEndian.Uint32(b[i*4:])), 0))
	}
	rest := b[head:]
	return coins, char.ReadItems(rest[:len(rest)/char.ItemSize*char.ItemSize]), nil
}

// writeVault writes vault.dat: an empty pool and items.
func writeVault(path string, items []char.Item) error {
	return os.WriteFile(path, slices.Concat(make([]byte, vaultCoins*4), char.WriteItems(items)), 0o644)
}
