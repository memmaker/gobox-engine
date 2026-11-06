package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// The series' View (VerbKeys) joins a classic turn's bar at its VerbOrder
// place, as the bar's ActView, when the game offers none.
func TestViewVerb(t *testing.T) {
	d := game.Data{VerbKeys: map[string]rune{"View": 'V'}, VerbOrder: []string{"Move", "View", "Aim", "Done"}}
	b := &front.Battle{Turn: &front.Turn{Can: []front.ActKind{front.ActWalk, front.ActAttack, front.ActEnd},
		Options: []front.TurnOption{{Label: "Move"}, {Label: "Aim"}, {Label: "Done"}}}}
	o := ViewVerb(b, d).Turn.Options
	if len(o) != 4 || o[1].Label != "View" || o[1].Kind != front.ActView || o[1].Key != 'V' {
		t.Fatalf("%+v", o)
	}
	if len(b.Turn.Options) != 3 {
		t.Fatal("the game's turn changed")
	}
}
