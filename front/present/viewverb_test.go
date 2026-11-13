package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// The series' View (VerbKeys) joins a classic turn's bar at its VerbOrder
// place, as the bar's ActView, when the game offers none.
func TestViewVerb(t *testing.T) {
	d := game.Data{VerbKeys: map[string]rune{"View": 'V'}, VerbOrder: []string{"Move", "View", "Aim", "Done"}}
	c := &play.Combat{Turn: &play.Turn{Options: []play.Option{{Label: "Move"}, {Label: "Aim"}, {Label: "Done"}}}}
	o := turnOptions(c, d)
	if act, _ := turnAct(o[1].Do); len(o) != 4 || o[1].Label != "View" || act != front.ActView || o[1].Key != 'V' {
		t.Fatalf("%+v", o)
	}
	if len(c.Turn.Options) != 3 {
		t.Fatal("the game's turn changed")
	}
}
