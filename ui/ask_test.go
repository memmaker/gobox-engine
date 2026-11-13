package ui

import (
	"reflect"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

func TestAsk(t *testing.T) {
	for _, c := range []struct {
		a    play.Ask
		want FrontAsk
	}{
		{play.Ask{Kind: game.Continue, Ms: 300}, FrontAsk{Kind: AskPause, Ms: 300}},
		{play.Ask{Kind: game.Continue, Prompt: "Done."}, FrontAsk{Kind: AskMenu, Menu: Menu{Prompt: "Done."}}},
		{play.Ask{Kind: game.Name, Max: 8, Typed: "SWO"}, FrontAsk{Kind: AskText, Max: 8, Text: "SWO"}},
		{play.Ask{Kind: game.PickItem, Lines: []game.Line{{Text: "H", Header: true, Off: true}, {Text: "LEAVE", Spell: 3}}},
			FrontAsk{Kind: AskList, Entries: []FrontEntry{{Text: "H", Header: true, Off: true}, {Text: "LEAVE", Spell: 3}}}},
	} {
		if q, ok := Ask(game.ModeAsk, &c.a); !ok || !reflect.DeepEqual(*q, c.want) {
			t.Errorf("%+v: %+v %v", c.a, q, ok)
		}
	}
}

// A menu the game opens over the 3D view, keeping ModeExplore, is the
// question up and ends the exploring (Esc closes it, not the options).
func TestAskOverExplore(t *testing.T) {
	s := shown{Mode: game.ModeExplore, Pending: &play.Ask{Kind: game.Choice, Options: []play.Option{{Label: "Exit", Key: 'E'}}}}
	if q, ok := Ask(s.Mode, s.Pending); !ok || q == nil || len(q.Commands) != 1 {
		t.Fatalf("no question over the 3D view: %+v", q)
	}
	over := play.Exploring{Common: play.Common{Ask: &play.Question{Replies: []string{"Exit"}}}}
	if (&frontState{}).flagsOf(over, nil).Exploring {
		t.Fatal("still exploring under the menu")
	}
}
