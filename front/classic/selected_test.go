package classic

import (
	"gobox/game"
	"testing"
)

func TestPartySelectedFromIndex(t *testing.T) {
	s := shown{Selected: 1, Party: []game.Member{{Name: "A"}, {Name: "B"}}}
	if p := Party(s.Party, s.Selected, nil); p[0].Selected || !p[1].Selected {
		t.Fatalf("selected %v %v, want member 1", p[0].Selected, p[1].Selected)
	}
	s.Party[0].Chosen = true // a candidate marked: it alone
	if p := Party(s.Party, s.Selected, nil); !p[0].Selected || p[1].Selected {
		t.Fatalf("candidate not alone")
	}
}
