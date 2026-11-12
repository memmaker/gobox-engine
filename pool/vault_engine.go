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
// The money is the loot's: it stays in the vault until the party takes it
// (Take on the money line), and what is left is written back, as pod and sotsb
// do. Deposit is the items menu's Drop while the vault is open (pod's dropItem).

const (
	vaultFile  = "vault.dat"
	vaultBlock = -1 // treasure's block number for the vault's items (AddTreasureBlock)
	// the three longints of the file, and their place among core.SetLoot's seven coins
	vaultCoins                          = 3
	coinPlatinum, coinGems, coinJewelry = 4, 5, 6
	goldPerPlatinum                     = 5
)

// vault is the open vault: its item records, which the treasure reads (vaultBlock).
type vault struct{ items []char.Item }

// Vault = pod's vault screen: the treasure screen over vault.dat until the party leaves.
func (e *Engine) Vault() {
	coins, items, err := readVault(e.savePath(vaultFile))
	if err != nil {
		panic(fmt.Sprintf("vault: %v", err))
	}
	e.vault = &vault{items: items}
	e.do(core.ClearLoot{})
	for coins != [vaultCoins]uint64{} { // AddLoot holds 16 bits a kind: add it in turns
		var c [7]uint16
		for i, k := range []int{coinPlatinum, coinGems, coinJewelry} {
			n := uint16(min(coins[i], math.MaxUint16))
			c[k], coins[i] = n, coins[i]-uint64(n)
		}
		e.do(core.AddLoot{Coins: c})
	}
	if len(items) > 0 {
		e.do(core.AddTreasureBlock{Block: vaultBlock})
	}
	e.push(cont{kind: kVaultDone})
	e.lootMenu()
}

// vaultDone: the vault is left; the money and items not taken are written back.
func (e *Engine) vaultDone() {
	e.vault = nil
	records := e.fact(core.NoChar, core.TreasureRecordsFact{}).(core.ItemRecords)
	loot := e.purse(true)
	e.do(core.ClearLoot{}) // and the treasure
	if err := writeVault(e.savePath(vaultFile), loot, records); err != nil {
		e.say("Can't save the vault: " + err.Error())
	}
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

// writeVault writes vault.dat: the pool's platinum, gems and jewelry (gold
// under a platinum piece and copper cannot be left, as in pod), then the item records.
func writeVault(path string, loot core.Purse, records core.ItemRecords) error {
	head := make([]byte, 0, vaultCoins*4)
	for _, n := range []uint64{loot.Gold / goldPerPlatinum, loot.Gems, loot.Jewelry} {
		head = binary.LittleEndian.AppendUint32(head, uint32(min(n, math.MaxInt32)))
	}
	return os.WriteFile(path, slices.Concat(head, records), 0o644)
}
