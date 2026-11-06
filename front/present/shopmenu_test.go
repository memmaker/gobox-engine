package present

import (
	"slices"
	"testing"

	"gobox/front"
	"gobox/ui"
)

// The shop menu is built from what the game allows: View always, Take,
// Pool and Share as allowed, so each game gets its original's bar.
func TestShopMenuFromAllowed(t *testing.T) {
	for _, c := range []struct {
		shop front.Shop
		want []string
	}{
		{front.Shop{}, []string{"Buy", "View", "Exit"}},
		{front.Shop{Take: true, Pool: true, Share: true}, []string{"Buy", "View", "Take", "Pool", "Share", "Exit"}},
	} {
		var p Presenter
		_, rows := p.shopItems(front.View{Shop: &c.shop})
		var got []string
		for _, r := range rows {
			got = append(got, r.Text)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%+v: %v, want %v", c.shop, got, c.want)
		}
	}
}

// TestTempleMenuFromAllowed: the temple offers the money acts its game
// allows (SotSB's Pool), each with its key.
func TestTempleMenuFromAllowed(t *testing.T) {
	var p Presenter
	_, rows := p.templeItems(front.View{Temple: &front.Temple{Money: front.Money3{Pool: true}}})
	i := slices.IndexFunc(rows, func(r ui.MenuRow) bool { return r.Text == "Pool" })
	if i < 0 || rows[i].Key != 'P' {
		t.Fatalf("rows %v", rows)
	}
	rows[i].Do()
	if a := p.Actions(); len(a) != 1 || a[0].Kind != front.ActLootPool {
		t.Errorf("Pool sends %v", a)
	}
}
