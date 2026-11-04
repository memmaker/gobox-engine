package parity

import (
	"image"
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestBorderParity: the border pieces classic takes from Data.Art equal
// por's, over the dev scenes.
func TestBorderParity(t *testing.T) {
	seen := 0
	for _, scene := range []string{"", "shop", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			v := e.View()
			b := v
			b.Border, b.BorderPix, b.BorderCaps, b.BorderCapsPix = [8]image.Image{}, [8]front.Picture{}, [4]image.Image{}, [4]front.Picture{}
			c := classic.Border(b, e.Data().Art)
			if !reflect.DeepEqual(c.BorderPix, v.BorderPix) || !reflect.DeepEqual(c.BorderCapsPix, v.BorderCapsPix) {
				t.Errorf("%s: border differs", scene)
			}
			if v.BorderPix[0].Pix != nil {
				seen++
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	if seen == 0 {
		t.Error("no border seen")
	}
}
