package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// Yes/no, square and direction questions are classic's too.
func TestAskKinds(t *testing.T) {
	for k, want := range map[game.AskKind]front.AskKind{game.YesNo: front.AskConfirm, game.Square: front.AskSquare, game.Direction: front.AskMove} {
		q, ok := Ask(game.ModeAsk, &play.Ask{Kind: k, Prompt: "?"})
		if !ok || q == nil || q.Kind != want {
			t.Errorf("kind %d: %+v %v", k, q, ok)
		}
	}
}
