package importpor

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"reflect"
	"slices"
	"testing"

	"gobox/game/play"
	"gobox/games/por/porgame"
	"gobox/importer"
)

// TestStartIsPORTitle: the engine shows start() as POR's title (ovr02:0626):
// TITLE 1 and 2 named with their pixels, the credits page after them, the
// lightning in EGA, TITLE 1 alone without the Full intro; Data's credits are
// start()'s.
func TestStartIsPORTitle(t *testing.T) {
	t.Chdir("..")
	e, err := porgame.New("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip("no game files:", err)
	}
	e.Cfg.Graphics = 'E'
	seq := *start()
	if err := e.SetStart(&seq); !errors.Is(err, fs.ErrNotExist) { // the test folder has no art/TITLE.PNG (the import copies it)
		t.Fatalf("SetStart without TITLE.PNG: %v", err)
	}
	seq.Backdrop = nil
	if err := e.SetStart(&seq); err != nil {
		t.Fatal(err)
	}
	e.FullIntro.Store(true)
	e.Start(e.Run)
	sc, ok := e.State().(play.Scene)
	if !ok || sc.Kind != play.SceneTitle || !sc.Full {
		t.Fatalf("not the title: %+v", e.State())
	}
	for n := 1; n <= 2; n++ {
		if l := sc.Steps[n].Layers; len(l) != 1 || l[0].File != "TITLE" || l[0].Block != n || l[0].Pixels == nil {
			t.Errorf("step %d: %+v", n, l)
		}
	}
	if !sc.Intro.Lightning || !reflect.DeepEqual(sc.Intro.Pictures, []int{1, 2, -1}) || sc.Steps[1].Song != "1" {
		t.Errorf("intro %+v, song %q", *sc.Intro, sc.Steps[1].Song)
	}
	e.FullIntro.Store(false)
	e.Start(e.Run) // State is built per change: the setting takes from the next one
	if short := e.State().(play.Scene); !reflect.DeepEqual(short.Intro.Pictures, []int{1}) {
		t.Errorf("short intro %+v", short.Intro)
	}
	d := e.Data()
	if len(d.CreditLines) != len(seq.Credits) || len(d.CreditBoxes) != 2 {
		t.Errorf("credits %d lines, boxes %v", len(d.CreditLines), d.CreditBoxes)
	}
}

// TestEndIsPORFinale: the engine takes end() (the manifest's End) as the finale's pictures,
// three steps with their pixels loaded and the animations' frames.
func TestEndIsPORFinale(t *testing.T) {
	t.Chdir("..")
	e, err := porgame.New("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip("no game files:", err)
	}
	e.Cfg.Graphics = 'E'
	if err := e.SetEnd(end()); err != nil {
		t.Fatal(err)
	}
	var ends int
	for _, st := range end().Steps {
		for _, l := range st.Layers {
			if len(l.Frames) > 0 {
				ends++
			}
		}
	}
	if ends != 3 || len(end().Steps) != 3 {
		t.Errorf("%d animations in %d steps", ends, len(end().Steps))
	}
}

// TestImportedStartCarriesCredits: the imported manifest's Start (data/por,
// written by `gobox import`) carries start()'s credits page: its step, lines,
// boxes.
func TestImportedStartCarriesCredits(t *testing.T) {
	b, err := os.ReadFile("../../../data/por/manifest.json")
	if err != nil {
		t.Skip("no imported POR:", err)
	}
	var m importer.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	want := start()
	if m.Start == nil || !slices.ContainsFunc(m.Start.Steps, func(s importer.SequenceStep) bool { return s.Kind == importer.StepCredits }) {
		t.Fatalf("imported Start has no credits step: %+v", m.Start)
	}
	if !reflect.DeepEqual(m.Start.Credits, want.Credits) || !reflect.DeepEqual(m.Start.Boxes, want.Boxes) {
		t.Errorf("imported credits %+v %+v\nwant %+v %+v (reimport)", m.Start.Credits, m.Start.Boxes, want.Credits, want.Boxes)
	}
}
