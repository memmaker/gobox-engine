package rules

// Purse is the party's one shared money: gold pieces, copper (under a gold
// piece) and gems and jewelry to appraise. uint64, so nothing wraps at uint32.
type Purse struct{ Gold, Copper, Gems, Jewelry uint64 }

// Rates are a game's coin rates as data. CopperPerGold: copper pieces in a
// gold piece (200 in the Gold Box games). Coin: the worth in copper of the
// leading coin kinds of a record, from copper up (1, 10, 100, 200, 1000).
type Rates struct {
	CopperPerGold uint64
	Coin          []uint64
}

// GoldBoxRates: copper, silver, electrum, gold, platinum (cotab, sotsb, por).
var GoldBoxRates = Rates{CopperPerGold: 200, Coin: []uint64{1, 10, 100, 200, 1000}}

// Add puts q into p; q's copper rolls over into gold.
func (p *Purse) Add(q Purse, r Rates) {
	p.Gold += q.Gold
	p.AddCopper(q.Copper, r)
	p.Gems += q.Gems
	p.Jewelry += q.Jewelry
}

// AddCopper adds n copper, changing every CopperPerGold into a gold piece.
func (p *Purse) AddCopper(n uint64, r Rates) {
	p.Copper += n
	p.Gold += p.Copper / r.CopperPerGold
	p.Copper %= r.CopperPerGold
}

// Pay takes gold from the purse; false (nothing taken) when short, so it never wraps below 0.
func (p *Purse) Pay(gold uint64) bool {
	if gold > p.Gold {
		return false
	}
	p.Gold -= gold
	return true
}

// Empty is true when the purse holds nothing.
func (p Purse) Empty() bool { return p == Purse{} }

// FromCoins counts a record's coins as gold and copper: the first
// len(r.Coin) kinds are money, the next two gems and jewelry (cotab and
// sotsb: 7 uint16 at record offset +251; widen them to uint64 first).
func FromCoins(coins []uint64, r Rates) Purse {
	var t uint64
	for i, v := range r.Coin {
		t += coins[i] * v
	}
	n := len(r.Coin)
	return Purse{Gold: t / r.CopperPerGold, Copper: t % r.CopperPerGold, Gems: coins[n], Jewelry: coins[n+1]}
}
