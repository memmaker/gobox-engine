package classic

import (
	"slices"
	"testing"

	"gobox/game"
)

// Distinct published combat Icons (POD's head/body numbers) give distinct
// figures from the shared pool collection.
func TestIconsDistinct(t *testing.T) {
	old := GameID
	defer func() { GameID = old }()
	GameID = "pod"
	var pics [][]uint8
	for _, ic := range []map[string]int{{"head": 0, "body": 24}, {"head": 9, "body": 23}, {"head": 5, "body": 23}} {
		h, b := memberIcon(game.Member{Icon: ic})
		pic, ok := poolIcon(h, b, nil, 0, false)
		if !ok {
			t.Fatalf("%v: no figure (%q %q)", ic, h, b)
		}
		for _, p := range pics {
			if slices.Equal(p, pic.Pix) {
				t.Fatalf("%v draws as another's figure", ic)
			}
		}
		pics = append(pics, pic.Pix)
	}
}
