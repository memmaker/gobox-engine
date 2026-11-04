package parity

import (
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestBorderParity: the border pieces classic takes from Data.Art, blank
// view in, are the ones por's art holds (borderSum), over the dev scenes.
func TestBorderParity(t *testing.T) {
	seen := 0
	for _, scene := range []string{"", "shop", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			c := classic.Border(front.View{}, e.Data().Art)
			if got := sum([]any{c.BorderPix, c.BorderCapsPix}); got != wantBorder {
				t.Errorf("%s: border %s, por %s", scene, got, wantBorder)
			}
			if c.BorderPix[0].Pix != nil {
				seen++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	if seen == 0 {
		t.Error("no border seen")
	}
}
