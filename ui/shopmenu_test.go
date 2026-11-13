package ui

import (
	"slices"
	"testing"

	"gobox/game/play"
)

// The shop menu is built from what the game allows: View always, Take,
// Pool and Share as allowed, so each game gets its original's bar.
func TestShopMenuFromAllowed(t *testing.T) {
	for _, c := range []struct {
		allowed []play.Action
		want    []string
	}{
		{nil, []string{"Buy", "View", "Exit"}},
		{[]play.Action{play.Take{}, play.PoolCoins{}, play.LootShare{}}, []string{"Buy", "View", "Take", "Pool", "Share", "Exit"}},
	} {
		g, _ := playGame(shopState(c.allowed...))
		p := &Presenter{}
		p.SetEngine(g)
		_, rows := p.shopItems()
		var got []string
		for _, r := range rows {
			got = append(got, r.Text)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%v: %v, want %v", c.allowed, got, c.want)
		}
	}
}

// TestTempleMenuFromAllowed: the temple offers the money acts its game
// allows (SotSB's Pool), each with its key.
func TestTempleMenuFromAllowed(t *testing.T) {
	g, _ := playGame(play.Trading{Kind: play.TradeTemple, Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.PoolCoins{}}}}})
	p := &Presenter{}
	p.SetEngine(g)
	_, rows := p.templeItems()
	i := slices.IndexFunc(rows, func(r MenuRow) bool { return r.Text == "Pool" })
	if i < 0 || rows[i].Key != 'P' {
		t.Fatalf("rows %v", rows)
	}
	rows[i].Do()
	if a := p.Actions(); len(a) != 1 || a[0].Kind != ActLootPool {
		t.Errorf("Pool sends %v", a)
	}
}
