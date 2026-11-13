package present

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// Buy on the loot screen (the view's Take) must take the loot, not buy.
func TestBuyOnLootTakes(t *testing.T) {
	r := &fake{st: play.Looting{Common: play.Common{Party: []game.Member{{}}}, Items: []game.Item{{Handle: 5, Name: "Gun"}}}}
	NewGame(r, game.Data{}).Buy(5)
	if len(r.sent) != 1 || r.sent[0] != (play.Take{Item: 5}) {
		t.Fatalf("Buy on loot sent %#v", r.sent)
	}
}
