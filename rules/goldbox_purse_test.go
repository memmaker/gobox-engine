package rules

import "testing"

// porPurse is por-go/engine/money.go, copied as the oracle.
type porPurse struct{ gold, copper, gems, jewelry uint32 }

func (p *porPurse) add(q porPurse) {
	p.gold += q.gold
	p.addCopper(q.copper)
	p.gems += q.gems
	p.jewelry += q.jewelry
}

func (p *porPurse) addCopper(n uint32) {
	p.copper += n
	p.gold += p.copper / 200
	p.copper %= 200
}

var porCoinCopper = [5]uint32{1, 10, 100, 200, 1000}

func porCoinsToPurse(c [7]uint16) porPurse {
	var t uint32
	for i, v := range porCoinCopper {
		t += uint32(c[i]) * v
	}
	return porPurse{gold: t / 200, copper: t % 200, gems: uint32(c[5]), jewelry: uint32(c[6])}
}

func TestPurseOracle(t *testing.T) {
	cases := [][7]uint16{{100, 10, 1, 3, 2, 4, 5}, {}, {150}, {199, 19, 1, 0, 0, 1, 1}, {65535, 65535, 65535, 65535, 65535, 9, 9}}
	var want porPurse
	var got Purse
	for _, c := range cases {
		w := porCoinsToPurse(c)
		in := make([]uint64, 7)
		for i, v := range c {
			in[i] = uint64(v)
		}
		g := FromCoins(in, GoldBoxRates)
		if uint64(w.gold) != g.Gold || uint64(w.copper) != g.Copper || uint64(w.gems) != g.Gems || uint64(w.jewelry) != g.Jewelry {
			t.Fatalf("FromCoins(%v) = %+v, por %+v", c, g, w)
		}
		want.add(w)
		got.Add(g, GoldBoxRates)
		want.addCopper(77)
		got.AddCopper(77, GoldBoxRates)
		if uint64(want.gold) != got.Gold || uint64(want.copper) != got.Copper || uint64(want.gems) != got.Gems || uint64(want.jewelry) != got.Jewelry {
			t.Fatalf("after %v: %+v, por %+v", c, got, want)
		}
	}
}

func TestPurseNoWrapAndPay(t *testing.T) {
	p := Purse{Gold: 1 << 32}
	p.AddCopper(200, GoldBoxRates)
	if p.Gold != 1<<32+1 || p.Copper != 0 {
		t.Fatalf("%+v", p)
	}
	if p.Pay(1<<32+2) || !p.Pay(1) || p.Gold != 1<<32 {
		t.Fatalf("pay: %+v", p)
	}
	if !(Purse{}).Empty() || p.Empty() {
		t.Fatal("Empty")
	}
}
