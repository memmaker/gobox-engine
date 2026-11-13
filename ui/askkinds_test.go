package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// Yes/no, square and direction questions are classic's too.
func TestAskKinds(t *testing.T) {
	for k, want := range map[game.AskKind]AskKind{game.YesNo: AskConfirm, game.Square: AskSquare, game.Direction: AskMove} {
		q, ok := Ask(game.ModeAsk, &play.Ask{Kind: k, Prompt: "?"})
		if !ok || q == nil || q.Kind != want {
			t.Errorf("kind %d: %+v %v", k, q, ok)
		}
	}
}
