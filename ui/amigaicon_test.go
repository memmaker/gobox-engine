package ui

import (
	"strings"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// Under F12's Amiga art a member's CHEAD/CBODY numbers resolve to the Amiga set only, else the DOS one.
func TestMemberIconAmiga(t *testing.T) {
	id, art, am := GameID, daxArt.art, daxArt.amiga
	defer func() { GameID, daxArt.art = id, art; switchArt(am) }()
	GameID, daxArt.art = "sotsb", asset.OpenArt("../data/sotsb/art")
	if !hasAmiga() {
		t.Skip("no sotsb Amiga art")
	}
	m := game.Member{Icon: map[string]int{"head": 0, "body": 0}}
	for _, a := range []bool{false, true} {
		switchArt(a)
		h, b := memberIcon(m)
		if h == "" || b == "" || strings.HasPrefix(h, "amiga_") != a || strings.HasPrefix(b, "amiga_") != a {
			t.Errorf("amiga %v: %q %q", a, h, b)
		}
	}
}
