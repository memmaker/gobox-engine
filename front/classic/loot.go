package classic

import (
	"gobox/front"
	"gobox/game"
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
	v.Pool = s.Allowed["PoolCoins"]
	if v.Take && s.Selected >= 0 && s.Selected < len(s.Party) {
		for _, sp := range s.Party[s.Selected].Spells {
			v.Detect = v.Detect || sp.Memorized && sp.Name == detectMagic
		}
	}
	return v
}
