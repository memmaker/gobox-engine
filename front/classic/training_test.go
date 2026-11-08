package classic

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestTraining(t *testing.T) {
	s := game.State{Party: []game.Member{{Name: "Ann"}}, Purse: map[string]int{game.Gold: 900},
		Training: &game.Training{Cost: 1000, Learning: true, Learn: []game.SpellRef{{ID: 9, Name: "Sleep", Level: 1}, {ID: 30, Name: "Fireball", Level: 3}}}}
	want := &front.Training{Name: "Ann", Price: 1000, Gold: 900, Learn: []front.BookSpell{{ID: 9, Name: "Sleep", Level: "1st Level"}, {ID: 30, Name: "Fireball", Level: "3rd Level"}}}
	if got := Training(s.Training, s.Party, s.Selected, s.Purse); !reflect.DeepEqual(got, want) {
		t.Errorf("learning: %+v", got)
	}
	s.Training = &game.Training{Cost: 1000, Why: map[int]string{0: "Not enough experience"}}
	if got := Training(s.Training, s.Party, s.Selected, s.Purse); got.Why != "Not enough experience" || got.Becomes != nil {
		t.Errorf("refused: %+v", got)
	}
}
