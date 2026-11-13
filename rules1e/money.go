package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
)

// The party's money. The original gave every character a purse of seven
// coin kinds plus a treasure pool (ovr22); here the party shares one purse
// and every coin is counted in gold pieces plus the copper left over.

// Purse is the party's money (see grules.Purse): gold pieces, copper (under
// 200, a gold piece), gems and jewelry.
type Purse = goldbox.Purse

func (e *RuleState) ClearLoot() { e.Loot = Purse{} } // DS:67F4

// CoinsToPurse counts coins of the seven kinds (the last two gems and
// jewelry) as gold and copper (coin worth in copper: DS:0D24, 200 to the
// gold piece: DS:0D30).
func CoinsToPurse(c [7]uint16) Purse {
	w := make([]uint64, 7)
	for i, v := range c {
		w[i] = uint64(v)
	}
	return goldbox.FromCoins(w, goldbox.GoldBoxRates)
}

// takeRecordCoins moves the coins in a character record (starting gold,
// money in an old save) into the party's purse.
func (e *RuleState) takeRecordCoins(c *Character) {
	var coins [7]uint16
	for i := range coins {
		coins[i] = c.Coins[i]
		c.Coins[i] = 0
	}
	e.MoneySound(CoinsToPurse(coins))
	e.Purse.Add(CoinsToPurse(coins), goldbox.GoldBoxRates)
}

// Pay takes gold from the purse; false (and a message) when there is not enough.
func (e *RuleState) Pay(gold uint64) bool {
	if !e.Purse.Pay(gold) {
		e.notice("Not enough money.")
		return false
	}
	return true
}

// purseCoin is the purse's word of coin c: its gold or its copper, the only
// money there is (the scripts' other coins became gold at import, script.GoldCoins).
func purseCoin(p *Purse, c core.Coin) (*uint64, bool) {
	switch c {
	case Gold:
		return &p.Gold, true
	case Copper:
		return &p.Copper, true
	}
	return nil, false
}

// pursePieces is p's count of unappraised jewelry (jewel) or gems.
func pursePieces(p *Purse, jewel bool) *uint64 {
	if jewel {
		return &p.Jewelry
	}
	return &p.Gems
}
