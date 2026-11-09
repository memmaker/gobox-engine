package classic

import (
	"gobox/game"
	"testing"
)

func TestPartySelectedFromIndex(t *testing.T) {
	s := shown{Selected: 1, Party: []game.Member{{Name: "A"}, {Name: "B"}}}
	if p := Party(s.Party, nil, s.Selected); p[0].Selected || !p[1].Selected {
		t.Fatalf("selected %v %v, want member 1", p[0].Selected, p[1].Selected)
	}
	if p := Party(s.Party, []memberCard{{Chosen: true}}, s.Selected); !p[0].Selected || p[1].Selected {
		t.Fatalf("candidate not alone")
	}
}
