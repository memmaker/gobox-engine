package importer_test

import (
	"encoding/json"
	"gobox/adventure"
	"gobox/importer"
	"gobox/pool"
	"os"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
)

// TestStartIsPORTitle: the engine shows start() as POR's title (ovr02:0626):
// steps 1 and 2 (TITLE 1 and 2), the credits page after them, the
// lightning in EGA, TITLE 1 alone without the Full intro.
func TestStartIsPORTitle(t *testing.T) {
	e, err := pool.New("../data/por", "../data/por/sources")
	if err != nil {
		t.Skip("no game files:", err)
	}
	e.Cfg.Graphics = 'E'
	seq := *importer.PorStart()
	if err := e.SetStart(&seq); err != nil {
		t.Fatal(err)
	}
	e.FullIntro.Store(true)
	e.Start(e.Run)
	sc, ok := e.State().(game.Scene)
	if !ok || sc.Kind != game.SceneTitle || !sc.Full {
		t.Fatalf("not the title: %+v", e.State())
	}
	for n := 1; n <= 2; n++ {
		if l := seq.Step(n).Layers; len(l) != 1 || l[0].File != "TITLE" || l[0].Block != n {
			t.Errorf("step %d: %+v", n, l)
		}
	}
	if !sc.Intro.Lightning || !reflect.DeepEqual(sc.Intro.Pictures, []int{1, 2, -1}) || seq.Step(1).Song != "1" {
		t.Errorf("intro %+v, song %q", *sc.Intro, seq.Step(1).Song)
	}
	e.FullIntro.Store(false)
	e.Start(e.Run) // State is built per change: the setting takes from the next one
	if short := e.State().(game.Scene); !reflect.DeepEqual(short.Intro.Pictures, []int{1}) {
		t.Errorf("short intro %+v", short.Intro)
	}
	if len(seq.Boxes) != 2 {
		t.Errorf("credits boxes %v", seq.Boxes)
	}
}

// TestEndIsPORFinale: the engine takes end() (the manifest's End) as the finale's pictures,
// three steps with their pixels loaded and the animations' frames.
func TestEndIsPORFinale(t *testing.T) {
	e, err := pool.New("../data/por", "../data/por/sources")
	if err != nil {
		t.Skip("no game files:", err)
	}
	e.Cfg.Graphics = 'E'
	if err := e.SetEnd(importer.PorEnd()); err != nil {
		t.Fatal(err)
	}
	var ends int
	for _, st := range importer.PorEnd().Steps {
		for _, l := range st.Layers {
			if len(l.Frames) > 0 {
				ends++
			}
		}
	}
	if ends != 3 || len(importer.PorEnd().Steps) != 3 {
		t.Errorf("%d animations in %d steps", ends, len(importer.PorEnd().Steps))
	}
}

// TestImportedStartCarriesCredits: the imported manifest's Start (data/por,
// written by `gobox import`) carries start()'s credits page: its step, lines,
// boxes.
func TestImportedStartCarriesCredits(t *testing.T) {
	b, err := os.ReadFile("../data/por/manifest.json")
	if err != nil {
		t.Skip("no imported POR:", err)
	}
	var m importer.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	want := importer.PorStart()
	if m.Start == nil || !slices.ContainsFunc(m.Start.Steps, func(s adventure.SequenceStep) bool { return s.Kind == adventure.StepCredits }) {
		t.Fatalf("imported Start has no credits step: %+v", m.Start)
	}
	if !reflect.DeepEqual(m.Start.Credits, want.Credits) || !reflect.DeepEqual(m.Start.Boxes, want.Boxes) {
		t.Errorf("imported credits %+v %+v\nwant %+v %+v (reimport)", m.Start.Credits, m.Start.Boxes, want.Credits, want.Boxes)
	}
}
