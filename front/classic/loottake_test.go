package classic

import (
	"testing"

	"gobox/game"
)

type takeRules struct {
	loreRules
	got game.Decision
}

func (r *takeRules) Apply(d game.Decision) game.Result { r.got = d; return game.Result{} }

// Buy on the loot screen (the view's Take) must take the loot, not buy.
func TestBuyOnLootTakes(t *testing.T) {
	r := &takeRules{loreRules: loreRules{s: game.State{Mode: game.ModeLoot, Party: []game.Member{{}}, Loot: &game.Loot{Items: []game.Item{{Name: "Gun"}}}}}}
	(&Game{R: r}).Buy(0)
	if _, ok := r.got.(game.TakeLoot); !ok {
		t.Fatalf("Buy on loot sent %#v", r.got)
	}
}
