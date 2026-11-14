package ui

import (
	"gobox/game"
	"strconv"
)

// detectMagic is the spell the treasure screen offers to cast on the treasure.
const detectMagic = "Detect Magic" // ponytail: by its Gold Box name; a game-given spell id if a game names it otherwise

// Loot is the treasure screen's options (nil: not up): Take while
// there are items, Detect while the selected member has Detect Magic
// memorized for them.
func Loot(mode game.Mode, l *game.Loot, allowed map[string]bool, party []game.Member, sel int) *FrontLoot {
	if mode != game.ModeLoot || l == nil {
		return nil
	}
	v := &FrontLoot{Take: len(l.Items) > 0}
	v.Share = allowed["LootShare"] // the rules' (none under the shared purse)
	v.Pool = allowed["PoolCoins"]
	v.Drop = allowed["Drop"]
	v.Status = l.Status
	for _, k := range []string{game.Platinum, game.Gold, game.Electrum, game.Silver, game.Copper, game.Credits, game.Gems, game.Jewelry} {
		if n := l.Money[k]; n > 0 {
			v.Money = append(v.Money, strconv.Itoa(n)+" "+k)
		}
	}
	if v.Take && sel >= 0 && sel < len(party) {
		for _, sp := range party[sel].Spells {
			v.Detect = v.Detect || sp.Memorized && sp.Name == detectMagic
		}
	}
	if on, set := allowed["LootDetect"]; set { // the game's own say wins (Curse's three detects, Pools' none)
		v.Detect = v.Take && on
	}
	return v
}
