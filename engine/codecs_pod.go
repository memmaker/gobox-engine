package engine

import (
	"fmt"

	"gobox/ecl"
	"gobox/rules"
)

// pod's values (pod-re C_Status, TCharRecord.side) and POR's canonical ones.
const (
	podSideParty     = 0 // TCharRecord.side: the party's
	podSideOther     = 1 // the other side
	podSymbolSetZero = 4 // TileSet 4 loads symbol set 0 (WriteGbvm $39)
	podSymbolSets    = 5 // TileSet writes from here on are ignored

)

// podStatus is C_Status's number for each status (pod-re includes/c_status.pas; 2 has no name).
var podStatus = map[rules.Status]int64{
	rules.StatusOkay: 0, rules.StatusAnimated: 1, rules.StatusRunning: 3, rules.StatusUnconscious: 4,
	rules.StatusDying: 5, rules.StatusDead: 6, rules.StatusStoned: 7, rules.StatusGone: 8,
}

// podCodecs are pod's encodings of canonical fields (its layout's enc=).
var podCodecs = map[ecl.Encoding]codec{
	"PodSelControl": {ignoreWrites: true}, // SpecialWriteZ1 has no case for it
	// TCharRecord.status, 0..8 (OKAY 0); SpecialWriteZ1 $1ED: inPlay = status = OKAY.
	"PodSelStatus": {
		toGame: func(e *Engine, _ int64) int64 { // the member's own status, not the canonical active bit
			st := e.ruleset().Sheet(e.selected()).Status
			n, ok := podStatus[st]
			if !ok {
				panic(fmt.Sprintf("pod: SelStatus: no C_Status for %s", st))
			}
			return n
		},
		toCanon: func(_ *Engine, g int64) int64 {
			switch g {
			case podStatus[rules.StatusOkay]:
				return canonStatusOkay // the revive
			case podStatus[rules.StatusDead]:
				return canonStatusDied
			case podStatus[rules.StatusStoned]:
				return canonStatusStoned
			}
			return canonStatusOut
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
		if e.worldWord(rules.WordIn3D) == 0 {
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
			if c == int64(scriptSide[rules.SideMonsters]) {
				return podSideOther
			}
			return podSideParty
		},
		toCanon: func(_ *Engine, g int64) int64 {
			switch g {
			case podSideParty:
				return int64(scriptSide[rules.SideParty])
			case podSideOther:
				return int64(scriptSide[rules.SideMonsters])
			}
			panic(fmt.Sprintf("pod: SelSide %d", g))
		},
	},
}
