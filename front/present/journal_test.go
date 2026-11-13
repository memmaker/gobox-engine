package present

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestJournalAndPowers(t *testing.T) {
	s := shown{Quests: []game.Quest{{ID: "Rats", Giver: "Bram", Place: "Inn", Brief: []string{"a", "b"}, Done: true}},
		PowerPick: &game.PowerPick{Name: "Ann", What: "Surge", Options: []game.PowerOption{{ID: 3, Name: "Smite", Aim: 2, Range: 4}}},
		Holding:   game.HoldDoor, Door: &game.Door{Locked: true, Pick: true}}
	if got, want := Quests(s.Quests), []front.Quest{{Name: "Rats", Giver: "Bram", Place: "Inn", Text: []string{"a", "b"}, Done: true}}; !reflect.DeepEqual(got, want) {
		t.Errorf("quests %+v", got)
	}
	v := Waits(front.View{}, s.Holding, s.Door, s.PowerPick)
	if v.Pending != front.PendDoor || !v.Door.Pick || v.Door.Bash || v.Powers.Options[0].Aim != front.AimFoe || v.Powers.Options[0].Range != 4 {
		t.Errorf("waits %v %+v %+v", v.Pending, v.Door, v.Powers)
	}
}
