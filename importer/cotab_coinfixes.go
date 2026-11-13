package importer

import (
	"slices"

	"gobox/ecl"
)

// scriptFixes are all of COTAB's import special cases: the journal's and the coins'.
var cotabScriptFixes = slices.Concat(JournalFixes, coinFixes)

// coinFixes: the sites where the import's gold (ecl.GoldCoins) cannot
// keep the original's byte arithmetic on the platinum word.
var coinFixes = []ecl.Fix{
	// area 35: the original added the winnings one platinum at a time to the word's low byte, carrying at
	// 255 into its high byte (Word7CC4); in gold the carry would reset the purse at 1275 gold.
	// CHANGES: COTAB script, area 35 winnings in gold.
	{Area: 35, Old: `    if SelCoins_6 == 1275 { goto L_89DF }
    SelCoins_6 = SelCoins_6 + 5
    goto L_89EE
`, New: `    &SelCoins_6 += 5 # gold and copper only (Felix 2026-11-12): a platinum won is 5 gold, no byte carry
    goto L_89EE
`},
}
