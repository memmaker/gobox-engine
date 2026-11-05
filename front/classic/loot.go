package classic

import (
	"gobox/front"
	"gobox/game"
	"strconv"
)

// detectMagic is the spell the treasure screen offers to cast on the treasure.
const detectMagic = "Detect Magic" // ponytail: by its Gold Box name; a game-given spell id if a game names it otherwise

// Loot is the treasure screen's options from s (nil: not up): Take while
// there are items, Detect while the selected member has Detect Magic
// memorized for them.
func Loot(s game.State) *front.Loot {
	l := s.Loot
	if s.Mode != game.ModeLoot || l == nil {
		return nil
	}
	v := &front.Loot{Take: len(l.Items) > 0}
	for _, n := range l.Money { // Share: only classic.Game has LootShare (front.Game's other makers lack it)
		v.Share = v.Share || n > 0
	}
	if on, set := s.Allowed["LootShare"]; set { // a game that has no Share says so (Silver Blades)
		v.Share = v.Share && on
	}
	v.Pool = s.Allowed["PoolCoins"]
	v.Status = l.Status
	for _, k := range []string{game.Platinum, game.Gold, game.Electrum, game.Silver, game.Copper, game.Credits, game.Gems, game.Jewelry} {
		if n := l.Money[k]; n > 0 {
			v.Money = append(v.Money, strconv.Itoa(n)+" "+k)
		}
	}
	if v.Take && s.Selected >= 0 && s.Selected < len(s.Party) {
		for _, sp := range s.Party[s.Selected].Spells {
			v.Detect = v.Detect || sp.Memorized && sp.Name == detectMagic
		}
	}
	if on, set := s.Allowed["LootDetect"]; set { // the game's own say wins (Curse's three detects, Pools' none)
		v.Detect = v.Take && on
	}
	return v
}
