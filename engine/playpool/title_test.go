package playpool

import (
	"image"
	"image/color"
	"os"
	"os/exec"
	"testing"

	"gobox/asset"
	"gobox/gpu"
	"gobox/importer"
	"gobox/ui"
)

// titleFront is game id's Front as Play starts it, headless, on the imported
// data/<id> (skipped when not imported), the full title on.
func titleFront(t *testing.T, id string) *ui.Front {
	t.Helper()
	adv := "../../data/" + id
	f, err := importer.Read(adv)
	if err != nil {
		t.Skipf("%s: not imported: %v", id, err)
	}
	eng, err := New(adv, adv+"/sources")
	if err != nil {
		t.Fatal(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	if err := eng.SetStart(f.Manifest.Start); err != nil {
		t.Fatal(err)
	}
	eng.FullIntro.Store(true)
	data := eng.Data()
	if data.Adventure != adv {
		t.Errorf("%s: Data.Adventure %q, want %q (the presenter's art and Theme.Border)", id, data.Adventure, adv)
	}
	fr := ui.NewFront(ui.NewGame(eng, data), data)
	fr.Start("")
	return fr
}

// lit is the box of dst's non-black pixels.
func lit(dst *gpu.Image) image.Rectangle {
	b := dst.Bounds()
	px := make([]byte, 4*b.Dx()*b.Dy())
	dst.ReadPixels(px)
	var r image.Rectangle
	for i := 0; i < len(px); i += 4 {
		if px[i]|px[i+1]|px[i+2] != 0 {
			x, y := i/4%b.Dx(), i/4/b.Dx()
			r = r.Union(image.Rect(x, y, x+1, y+1))
		}
	}
	return r
}

// TestTitleScreens plays each imported game's full title headless, a key a
// step, to the main menu (cotab's crashed there on an empty logo cut): an
// Amiga title picture whose block has its own palette carries it (cotab's
// were drawn in POR's asset.Amiga slots), and the credits page's frame sits
// on the window's edges at K (cotab's and sotsb's had none: Theme.Border
// did not decode without Data.Adventure).
func TestTitleScreens(t *testing.T) {
	const w, h = 1280, 800
	if id := os.Getenv("TITLE_GAME"); id == "" { // a process per game: the presenter's data (SetData) is the process's
		for _, id := range []string{"por", "cotab", "sotsb", "pod"} {
			cmd := exec.Command(os.Args[0], "-test.run=^TestTitleScreens$", "-test.v")
			cmd.Env = append(os.Environ(), "TITLE_GAME="+id)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s: %v\n%s", id, err, out)
			}
		}
		return
	}
	for _, id := range []string{os.Getenv("TITLE_GAME")} {
		t.Run(id, func(t *testing.T) {
			fr := titleFront(t, id)
			art := asset.OpenArt("../../data/" + id + "/art")
			dst := gpu.NewImage(w, h)
			defer dst.Deallocate()
			k := ui.K(dst.Bounds())
			sawCredits := false
			for step := 0; step < 40 && !fr.TitleDrawn(); step++ {
				fr.Act()
				pg, credits := fr.Game.Page(), fr.Game.Flags().Credits
				fr.Draw(dst)
				fr.Draw(dst) // the step's first frame can still be the last one's (the presenter's page caches a frame behind)
				for _, l := range pg.Layers {
					pc := l.Picture
					b := art.Block(asset.Amiga, pc.File, pc.Block)
					if pc.Pixels == nil || b == nil || len(b.Palette) == 0 || b.ColourBase < asset.AmigaBase {
						continue
					}
					for i := 0; i < len(b.Palette)/3; i++ {
						want := color.RGBA{b.Palette[3*i], b.Palette[3*i+1], b.Palette[3*i+2], 255}
						if len(pc.Palette) <= b.ColourBase+i || pc.Palette[b.ColourBase+i] != want {
							t.Fatalf("step %d: %s %d without its block's palette (colour %d: want %v)", step, pc.File, pc.Block, i, want)
						}
					}
				}
				if credits {
					sawCredits = true
					if r := lit(dst); r.Min.X > 16*k || r.Min.Y > 16*k || r.Max.X < w-16*k || r.Max.Y < h-16*k {
						t.Errorf("step %d: credits lit %v: no frame on the window's edges (K %d)", step, r, k)
					}
				}
				fr.Key(13)
				fr.Tick()
			}
			if !fr.TitleDrawn() {
				t.Errorf("never reached the main menu")
			}
			if m, _ := importer.Read("../../data/" + id); m.Manifest.Start != nil && len(m.Manifest.Start.Credits) > 0 && !sawCredits {
				t.Errorf("credits page never shown")
			}
		})
	}
}
