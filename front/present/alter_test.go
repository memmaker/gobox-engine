package present

import (
	"gobox/front"
	"testing"
)

// A game's Camp.Settings are camp rows that step the value on through ActAlter.
func TestCampSettings(t *testing.T) {
	p := &Presenter{}
	v := front.View{Camp: &front.Camp{Settings: []front.Setting{{Name: "Speed", Value: 9, Max: 9}, {Name: "Level", Max: 4, Names: []string{"Novice"}}}}}
	_, rows := p.campItems(v)
	if rows[0].Text != "Speed: 9" || rows[1].Text != "Level: Novice" {
		t.Fatal(rows[0].Text, rows[1].Text)
	}
	rows[0].Do()
	if a := p.inv.acts[0]; a.Kind != front.ActAlter || a.N != 0 || a.X != 0 {
		t.Fatal(a)
	}
}
