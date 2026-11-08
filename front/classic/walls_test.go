package classic

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestWalls(t *testing.T) {
	v := front.View{}
	if Walls(v, nil, nil).Walls != nil {
		t.Error("walls without State")
	}
	art := map[string]game.Picture{"wall/1": {W: 2, Pixels: []byte{1, 2}, Clear: -1}, "wall/2": {File: "GEO", Block: 3, Frame: 4}}
	s := game.State{Walls: &game.Walls{Key: "k", Side: [][][4]int{{{1, 0, 2, 0}}}, Fronts: map[int]string{1: "wall/1", 2: "wall/2"}, Sun: true}}
	w := Walls(v, s.Walls, art).Walls
	want := map[int]front.Picture{1: {Pix: []byte{1, 2}, PixW: 2}, 2: {File: "GEO", Block: 3, Frame: 4}}
	if w == nil || w.Key != "k" || !w.Sun || !reflect.DeepEqual(w.FrontPix, want) {
		t.Errorf("walls %+v", w)
	}
}
