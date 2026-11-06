package present

import (
	"slices"
	"testing"

	"gobox/front"
)

// The shop menu is built from what the game allows: View always, Take,
// Pool and Share as allowed, so each game gets its original's bar.
func TestShopMenuFromAllowed(t *testing.T) {
	for _, c := range []struct {
		shop front.Shop
		want []string
	}{
		{front.Shop{}, []string{"Browse", "View", "Exit"}},
		{front.Shop{Take: true, Pool: true, Share: true}, []string{"Browse", "View", "Take", "Pool", "Share", "Exit"}},
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
