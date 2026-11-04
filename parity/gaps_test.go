package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestEffectsClues: por's EffectsList and ClueRead and the adapter's
// (served from State.Effects and Data.Clues) agree.
func TestEffectsClues(t *testing.T) {
	seen := false
	for _, scene := range []string{"", "effects", "journal", "talljournal", "parlay"} {
		t.Run(scene, func(t *testing.T) {
			a := engine(t, true, scene)
			g := classic.NewGame(a.Rules(), a.Data())
			for n := range 6 {
				want, got := a.EffectsList(n), g.EffectsList(n)
				seen = seen || len(want) > 0
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("EffectsList(%d): adapter %+v, por %+v", n, got, want)
				}
			}
			if len(g.ClueChapters()) == 0 {
				t.Fatal("no clue chapters")
			}
			for i := range g.ClueChapters() {
				var want, got *front.Read = a.ClueRead(i), g.ClueRead(i)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("ClueRead(%d): adapter %+v, por %+v", i, got, want)
				}
			}
		})
	}
	if !seen {
		t.Fatal("no scene had effects")
	}
}
