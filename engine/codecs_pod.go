package engine

import (
	"fmt"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
)

// podCodecs are pod's encodings of canonical fields (its layout's enc=).
var podCodecs = map[ecl.Encoding]codec{
	"PodSelControl": {ignoreWrites: true}, // SpecialWriteZ1 has no case for it
	// TCharRecord.status, 0..9 (OKAY 0); the write is adventure.podStatusWrite (SpecialWriteZ1 $1ED).
	"PodSelStatus": {
		toGame: func(_ *Engine, c int64) int64 {
			st := adventure.StatusKinds[c]
			n, ok := adventure.PodStatus[st]
			if !ok {
				panic(fmt.Sprintf("pod: SelStatus: no C_Status for %s", st))
			}
			return n
		},
	},
	// TCharRecord.coins[0..2]: platinum, gems, jewelry, words at bytes 0, 2, 4. The import rewrote the
	// platinum word to gold (ecl.GoldCoins): the canonical elements 6 (gold), 10 and 12 (selCoin),
	// the high bytes the odd elements after them.
	"PodSelCoins": {element: func(i int) int {
		if i < 2 {
			return 6 + i
		}
		return 8 + i
	}},
	// TCharRecord.spells[0..140]: slot 0 first, where the canonical element i is slot i+1.
	"PodSelMemorized": {element: func(i int) int { return i - 1 }},
	// ReadGbvm $10: facing div 2 in a dungeon, the raw 0..7 on the area map; the write (0..3, less 4 till it
	// fits) is the canonical one.
	"PodFacing": {toGame: func(e *Engine, c int64) int64 {
		if e.world.In3D == 0 {
			return int64(e.facing)
		}
		return c
	}},
	"PodSymbolSet": {}, // decoded only (adventure.Canonical)
	// TCharRecord.side: 0 the party's, 1 the other; the canonical 0 party, 0x80 auto (party), 0x81 monsters.
	"PodSelSide": {
		toGame: func(_ *Engine, c int64) int64 {
			if c == int64(adventure.ScriptSide[rules.SideMonsters]) {
				return adventure.PodSideOther
			}
			return adventure.PodSideParty
		},
	},
}
