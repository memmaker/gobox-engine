package classic

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestResults(t *testing.T) {
	if Results(front.View{}, nil).Results != nil {
		t.Error("results without State")
	}
	s := game.State{Results: &game.Results{Shares: []string{"a"}, Outcome: "won", Exp: "100", Taken: "gold", Harmed: []game.Harm{{Name: "X", What: "dead"}}, End: "end", EndAt: [4]int{1, 2, 3, 4}}}
	want := &front.Results{Shares: []string{"a"}, Outcome: "won", Exp: "100", Taken: "gold", Harmed: []front.Harm{{Name: "X", What: "dead"}}, End: "end", EndAt: front.CellRect{X1: 1, Y1: 2, X2: 3, Y2: 4}}
	if got := Results(front.View{}, s.Results).Results; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v", got)
	}
}
