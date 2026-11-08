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
			front.Ask{Kind: front.AskList, Entries: []front.Entry{{Text: "H", Header: true, Off: true}, {Text: "LEAVE", Spell: 3}}}},
	} {
		if q, ok := Ask(game.ModeAsk, &c.a); !ok || !reflect.DeepEqual(*q, c.want) {
			t.Errorf("%+v: %+v %v", c.a, q, ok)
		}
	}
}

// A menu the game opens over the 3D view, keeping ModeExplore, is the
// question up and ends the exploring (Esc closes it, not the options).
func TestAskOverExplore(t *testing.T) {
	s := game.State{Mode: game.ModeExplore, Pending: &game.Ask{Kind: game.Choice, Options: []game.Option{{Label: "Exit", Key: 'E'}}},
		Flags: &game.Flags{Exploring: true}}
	if q, ok := Ask(s.Mode, s.Pending); !ok || q == nil || len(q.Commands) != 1 {
		t.Fatalf("no question over the 3D view: %+v", q)
	}
	if Flags(front.View{}, s.Flags, OverMenu(s.Mode, s.Pending)).Exploring {
		t.Fatal("still exploring under the menu")
	}
}
