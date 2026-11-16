package ui

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/rules"
)

type openMap struct{}

func (openMap) Name() string                    { return "" }
func (openMap) Size() image.Point               { return image.Pt(8, 8) }
func (openMap) Edge(image.Point, int) game.Edge { return game.Edge{} }
func (openMap) Departure(image.Point, int) bool { return false }
func (openMap) Event(image.Point) bool          { return false }

type walkerGame struct{ sent []game.Action }

func (w *walkerGame) Version() uint64 { return fresh() }

func (w *walkerGame) State() game.State {
	return game.Exploring{Map: game.ExploreMap{Walk: openMap{}, Automap: allKnown(8, 8)}}
}

// allKnown is a w x h automap the party has seen all of, with no walls.
func allKnown(w, h int) game.Automap {
	am := game.Automap{Size: rules.Size{W: w, H: h}}
	for y := range h {
		for x := range w {
			am.Known = append(am.Known, rules.MapPos{X: x, Y: y})
		}
	}
	return am
}
func (w *walkerGame) Do(a game.Action) (game.Outcome, error) {
	w.sent = append(w.sent, a)
	return game.Outcome{}, nil
}
func (w *walkerGame) Save() ([]byte, error) { return nil, nil }
func (w *walkerGame) Load([]byte) error     { return nil }

// TestAnyInputStopsWalk: any key or click stops the auto-walk, and does
// nothing else: no action reaches the game, no menu opens.
func TestAnyInputStopsWalk(t *testing.T) {
	for name, input := range map[string]func(*Front){
		"key":   func(f *Front) { f.Key('x') },
		"enter": func(f *Front) { f.Key('\r') },
		"click": func(f *Front) { f.Click(10, 10) },
		"view":  func(f *Front) { f.ViewClick(10, 10) },
	} {
		w := &walkerGame{}
		f := &Front{Game: &classicGame{Game: NewGame(w, game.Data{})}, started: true}
		f.P.SetEngine(f.Game)
		if !f.Game.WalkTo(3, 0) {
			t.Fatal("no walk")
		}
		if f.P.pacing().Pace != game.PaceWalk {
			t.Fatal("not walking")
		}
		input(f)
		for _, c := range f.P.Calls() {
			c(f.Game)
		}
		if f.P.pacing().Pace == game.PaceWalk {
			t.Errorf("%s did not stop the walk", name)
		}
		if len(w.sent) != 0 {
			t.Errorf("%s sent %v", name, w.sent)
		}
	}
}
