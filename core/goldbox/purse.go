package goldbox

import "gobox/core"

// Purse and Rates are core's (one purse shape for every rule set).
type (
	Purse = core.Purse
	Rates = core.Rates
)

// GoldBoxRates: copper, silver, electrum, gold, platinum (cotab, sotsb, por).
var GoldBoxRates = Rates{CopperPerGold: 200, Coin: []uint64{1, 10, 100, 200, 1000}}

// FromCoins is core's.
var FromCoins = core.FromCoins
