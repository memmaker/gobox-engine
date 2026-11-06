package parity

import (
	"reflect"
	"slices"
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
				got := g.EffectsList(n)
				var want []front.Entry
				for _, l := range a.EffectsList(n) {
					want = append(want, front.Entry{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell})
				}
				seen = seen || len(want) > 0
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("EffectsList(%d): adapter %+v, por %+v", n, got, want)
				}
			}
			if len(g.ClueChapters()) == 0 {
				t.Fatal("no clue chapters")
			}
			for i := range g.ClueChapters() {
				want, got := a.ClueRead(i), g.ClueRead(i)
				if got.Title != want.Title || got.Page != want.Page || !slices.Equal(got.Lines, want.Lines) {
					t.Fatalf("ClueRead(%d): adapter %+v, por %+v", i, got, want)
				}
			}
		})
	}
	if !seen {
		t.Fatal("no scene had effects")
	}
}
