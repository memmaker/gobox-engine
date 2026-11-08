package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"gobox/game/play"
)

type openMap struct{}

func (openMap) Name() string                    { return "" }
func (openMap) Size() image.Point               { return image.Pt(8, 8) }
func (openMap) Edge(image.Point, int) game.Edge { return game.Edge{} }
func (openMap) Departure(image.Point, int) bool { return false }
func (openMap) Event(image.Point) bool          { return false }

type walkerGame struct{ sent []play.Action }

func (w *walkerGame) State() play.State {
	return play.Exploring{Map: play.ExploreMap{Walk: openMap{}}}
}
func (w *walkerGame) Do(a play.Action) (play.Outcome, error) {
	w.sent = append(w.sent, a)
	return play.Outcome{}, nil
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
		f := &Front{Game: classic.NewGame(w, game.Data{}), started: true}
		f.P.SetEngine(f.Game)
		if !f.Game.WalkTo(3, 0) {
			t.Fatal("no walk")
		}
		f.stale()
		if f.view().Pace != front.PaceWalk {
			t.Fatal("not walking")
		}
		input(f)
		for _, c := range f.P.Calls() {
			c(f.Game)
		}
		f.stale()
		if f.view().Pace == front.PaceWalk {
			t.Errorf("%s did not stop the walk", name)
		}
		if len(w.sent) != 0 {
			t.Errorf("%s sent %v", name, w.sent)
		}
	}
}
