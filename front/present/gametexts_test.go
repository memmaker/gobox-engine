package present

import (
	"testing"

	"gobox/game"
)

func TestGameDataTexts(t *testing.T) {
	old := texts
	defer func() { texts = old }()
	texts = nil
	gameData(game.Data{Texts: map[string]string{"Trade to whom?": "Trade with Whom?"}})
	if got := Text("Trade to whom?"); got != "Trade with Whom?" {
		t.Fatalf("Text = %q", got)
	}
}
