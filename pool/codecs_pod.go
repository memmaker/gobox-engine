package pool

import (
	"fmt"

	"gobox/core"
	"gobox/ecl/script"
)

// pod's values (pod-re C_Status, TCharRecord.side) and POR's canonical ones.
const (
	podStatusOkay      = 0 // C_Status.OKAY
	podStatusDead      = 6 // C_Status.DEAD_STATUS
	podStatusPetrified = 7 // C_Status.PETRIFIED
	podStatusGone      = 8 // C_Status.GONE
	podSideParty       = 0 // TCharRecord.side: the party's
	podSideOther       = 1 // the other side
	podSymbolSetZero   = 4 // TileSet 4 loads symbol set 0 (WriteGbvm $39)
	podSymbolSets      = 5 // TileSet writes from here on are ignored

	canonStatusOkay   = 1    // SelStatus: active
	canonStatusOut    = 0x80 // inactive (any write >= 0x80 but 0x87)
	canonStatusStoned = 0x87
	canonStatusDied   = 0 // a write: the member died
)

// podCodecs are pod's encodings of canonical fields (its layout's enc=).
var podCodecs = map[script.Encoding]codec{
	// TCharRecord.status, 0..8 (OKAY 0); SpecialWriteZ1 $1ED: inPlay = status = OKAY.
	"PodSelStatus": {
		toGame: func(_ *Engine, c int64) int64 {
			if c == canonStatusOkay {
				return podStatusOkay
			}
			return podStatusGone // ponytail: the canonical read says only "not active"; gone stands for every other status
		},
		toCanon: func(_ *Engine, g int64) int64 {
			switch g {
			case podStatusOkay:
				return canonStatusOkay // POR has no revive write: 1 only stores
			case podStatusDead:
				return canonStatusDied
			case podStatusPetrified:
				return canonStatusStoned
			}
			return canonStatusOut
		},
	},
	// TCharRecord.coins[0..2]: platinum, gems, jewelry, words at bytes 0, 2, 4. Platinum is the canonical
	// coin pair 8/9; gems and jewelry have no canonical coin, so they go to the odd (no field) elements and
	// stay script cells only.
	"PodSelCoins": {element: func(i int) int { return [6]int{8, 9, 1, 3, 5, 7}[i] }},
	// TCharRecord.spells[0..140]: slot 0 first, where the canonical element i is slot i+1.
	"PodSelMemorized": {element: func(i int) int { return i - 1 }},
	// ReadGbvm $10: facing div 2 in a dungeon, the raw 0..7 on the area map; the write (0..3, less 4 till it
	// fits) is the canonical one.
	"PodFacing": {toGame: func(e *Engine, c int64) int64 {
		if e.worldWord(core.WordIn3D) == 0 {
			return int64(e.facing)
		}
		return c
	}},
	// TileSet (WriteGbvm $39): 0..3 the set, 4 set 0, 5 and up no change.
	"PodSymbolSet": {toCanon: func(e *Engine, g int64) int64 {
		switch {
		case g == podSymbolSetZero:
			return 0
		case g >= podSymbolSets:
			return int64(e.set)
		}
		return g
	}},
	// TCharRecord.side: 0 the party's, 1 the other; the canonical 0 party, 0x80 auto (party), 0x81 monsters.
	"PodSelSide": {
		toGame: func(_ *Engine, c int64) int64 {
			if c == int64(scriptSide[core.SideMonsters]) {
				return podSideOther
			}
			return podSideParty
		},
		toCanon: func(_ *Engine, g int64) int64 {
			switch g {
			case podSideParty:
				return int64(scriptSide[core.SideParty])
			case podSideOther:
				return int64(scriptSide[core.SideMonsters])
			}
			panic(fmt.Sprintf("pod: SelSide %d", g))
		},
	},
}
