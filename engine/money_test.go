package engine

import (
	"gobox/char"
	"gobox/rules"
	"gobox/rules1e"
	"testing"
)

func TestPurse(t *testing.T) {
	if p := rules1e.CoinsToPurse([7]uint16{100, 10, 1, 3, 2, 4, 5}); p != (rules1e.Purse{Gold: 14, Copper: 100, Gems: 4, Jewelry: 5}) {
		t.Fatalf("coinsToPurse = %+v", p)
	}
	want := rules1e.Purse{Gold: 1234, Copper: 7, Gems: 2, Jewelry: 3}
	if got := char.ReadPurse(char.AppendPurse(nil, want)); got != want {
		t.Fatalf("save round trip = %+v", got)
	}
	// 150 copper twice is a gold piece and 100 copper, not rounded per pile
	var p rules1e.Purse
	p.Add(rules1e.CoinsToPurse([7]uint16{150}), rules.GoldBoxRates)
	p.Add(rules1e.CoinsToPurse([7]uint16{150}), rules.GoldBoxRates)
	if p != (rules1e.Purse{Gold: 1, Copper: 100}) {
		t.Fatalf("two piles = %+v", p)
	}
	if got := char.ReadPurse(char.AppendPurse(nil, want)[:12]); got != (rules1e.Purse{Gold: 1234, Gems: 2, Jewelry: 3}) {
		t.Fatalf("old save = %+v", got)
	}
}
