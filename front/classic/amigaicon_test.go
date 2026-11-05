package classic

import (
	"strings"
	"testing"

	"gobox/game"
)

// Under F12's Amiga art a member's CHEAD/CBODY numbers resolve to the Amiga set only, else the DOS one.
func TestMemberIconAmiga(t *testing.T) {
	id, am := GameID, Amiga
	defer func() { GameID, Amiga = id, am }()
	GameID = "sotsb"
	m := game.Member{Icon: map[string]int{"head": 0, "body": 0}}
	for _, a := range []bool{false, true} {
		Amiga = func() bool { return a }
		h, b := memberIcon(m)
		if h == "" || b == "" || strings.HasPrefix(h, "amiga_") != a || strings.HasPrefix(b, "amiga_") != a {
			t.Errorf("amiga %v: %q %q", a, h, b)
		}
	}
}
