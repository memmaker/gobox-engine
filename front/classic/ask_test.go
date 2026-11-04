package classic

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestAsk(t *testing.T) {
	for _, c := range []struct {
		a    game.Ask
		want front.Ask
	}{
		{game.Ask{Kind: game.Continue, Ms: 300}, front.Ask{Kind: front.AskPause, Ms: 300}},
		{game.Ask{Kind: game.Continue, Prompt: "Done."}, front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "Done."}}},
		{game.Ask{Kind: game.Name, Max: 8, Typed: "SWO"}, front.Ask{Kind: front.AskText, Max: 8, Text: "SWO"}},
		{game.Ask{Kind: game.PickItem, Lines: []game.Line{{Text: "H", Header: true, Off: true}, {Text: "LEAVE", Spell: 3}}},
			front.Ask{Kind: front.AskList, List: []front.ListLine{{Text: "H", Header: true, Off: true}, {Text: "LEAVE", Spell: 3}}}},
	} {
		if q, ok := Ask(game.State{Mode: game.ModeAsk, Pending: &c.a}); !ok || !reflect.DeepEqual(*q, c.want) {
			t.Errorf("%+v: %+v %v", c.a, q, ok)
		}
	}
}
