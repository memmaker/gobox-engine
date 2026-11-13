package amigaimport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/dax"
	"gobox/ega"
)

const amigaDir = "/Users/felix/Desktop/Games/Goldbox/Amiga"

// The archive conversions the later games share; Base/Frames say how the
// unindexed libraries map to DOS block ids (checked against the DOS data).
var later = map[string]Gfx{
	"8x8d": {}, "bigpic": {}, "body": {}, "head": {}, "cbody": {}, "chead": {},
	"cpic": {}, "comspr": {}, "tiles": {}, "back": {}, "skygrnd": {}, "gen": {},
	"dungcom": {Frames: true, Base: 1}, "wildcom": {Frames: true, Base: 1}, "randcom": {Frames: true, Base: 1},
	"sky":   {Base: 0xfa},
	"title": {}, "intro": {},
	"pic": {Sprite: true, XOR: true}, "final": {Sprite: true, XOR: true}, "frame": {Sprite: true},
	"sprit": {Sprite: true},
}

var games = map[string]Spec{
	"cotab": {Gfx: later, Disks: []string{"CurseA", "CurseB"}},
	"sotsb": {Gfx: later, Disks: []string{"Secret 1", "Secret 2"}},
	"pod":   {Gfx: later, Disks: []string{"POD_Disk1_v.1.00", "POD_Disk2_v.1.00", "POD_Disk3_v.1.00"}},
}

func importGame(t *testing.T, s Spec) string {
	if _, err := os.Stat(amigaDir); err != nil {
		t.Skip("no Amiga disks")
	}
	out := t.TempDir()
	if err := Import(amigaDir, out, s); err != nil {
		t.Fatal(err)
	}
	return out
}

// frames decodes every block of out/NAME.DAX and returns the frame count.
func frames(t *testing.T, out, name string, sprite bool) (blocks, n int) {
	f, err := dax.Open(filepath.Join(out, name+".DAX"))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	for _, e := range f.Entries {
		b := f.BlockShared(e.ID)
		if sprite {
			fs := dax.ParseSprite(b, true, 0, true)
			if len(fs) == 0 || fs[0].Image.Height == 0 {
				t.Errorf("%s %d: no frames", name, e.ID)
			}
			n += len(fs)
		} else {
			im, err := dax.ParseImage(b, true, 0)
			if err != nil || im.Count == 0 || len(im.Data) != im.Count*im.BPP {
				t.Errorf("%s %d: bad picture", name, e.ID)
				continue
			}
			n += im.Count
		}
		blocks++
	}
	return blocks, n
}

func TestGames(t *testing.T) {
	want := map[string]map[string][2]int{ // archive: least blocks, least frames
		"cotab": {"CHEAD": {56, 56}, "CBODY": {128, 128}, "CPIC": {130, 130}, "HEAD": {33, 33}, "BODY": {24, 24},
			"PIC": {40, 100}, "SPRIT": {40, 120}, "DUNGCOM": {1, 25}, "WILDCOM": {1, 34}, "8X8D": {18, 1000}, "TITLE": {3, 3}, "BIGPIC": {4, 4}},
		"sotsb": {"CHEAD": {56, 56}, "CBODY": {128, 128}, "CPIC": {100, 100}, "PIC": {38, 90}, "SPRIT": {34, 100},
			"DUNGCOM": {1, 30}, "8X8D": {19, 1000}, "TITLE": {5, 5}, "BIGPIC": {8, 8}},
		"pod": {"CHEAD": {28, 28}, "CBODY": {128, 128}, "CPIC": {240, 240}, "PIC": {49, 72}, "SPRIT": {29, 87},
			"DUNGCOM": {6, 150}, "8X8D": {6, 1000}, "TITLE": {5, 5}, "BIGPIC": {7, 7}, "FINAL": {11, 11}, "FRAME": {4, 92}},
	}
	for game, s := range games {
		t.Run(game, func(t *testing.T) {
			out := importGame(t, s)
			ents, _ := os.ReadDir(out)
			var names []string
			for _, e := range ents {
				names = append(names, e.Name())
			}
			t.Log(names)
			for name, w := range want[game] {
				g := s.Gfx[strings.ToLower(name)]
				b, n := frames(t, out, name, g.Sprite)
				if b < w[0] || n < w[1] {
					t.Errorf("%s: %d blocks %d frames, want at least %v", name, b, n, w)
				}
			}
		})
	}
}

// Pools of Darkness pictures have 32 colours: written one byte a pixel,
// they come back past the first palette slot.
func TestWide(t *testing.T) {
	out := importGame(t, games["pod"])
	f, err := dax.Open(filepath.Join(out, "PIC.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	fs := dax.ParseSprite(f.BlockShared(f.Entries[0].ID), false, 0, true)
	for _, c := range fs[0].Image.Data {
		if c >= ega.AmigaBase+16 {
			return
		}
	}
	t.Error("no colour past 15")
}

// POR keeps its single-disk behaviour: with Disks set to its data disk,
// the same archives come out, and Disks keeps other games' disks out.
func TestPORDisks(t *testing.T) {
	gfx := map[string]Gfx{"pic": {Sprite: true, XOR: true}, "chead": {}, "cpic": {}, "8x8d": {Split: true}}
	old := importGame(t, Spec{Gfx: gfx, MinGfx: 3})
	nu := importGame(t, Spec{Gfx: gfx, Disks: []string{"POOLDATA"}})
	for _, n := range []string{"PIC.DAX", "CHEAD.DAX", "CPIC.DAX", "8X8D.DAX"} {
		a, err1 := os.ReadFile(filepath.Join(old, n))
		b, err2 := os.ReadFile(filepath.Join(nu, n))
		if err1 != nil || err2 != nil || string(a) != string(b) {
			t.Errorf("%s differs (%v %v)", n, err1, err2)
		}
	}
}

// Curse and Silver Blades colour pictures without a palette of their own
// from the fixed one in the program; it lands near the DOS EGA colours.
func TestFixedPalette(t *testing.T) {
	for _, c := range []struct {
		game, exe, dos, name string
		pal                  int
	}{
		{"cotab", "Curse", "/Users/felix/Projects/gobox-engine/data/cotab/sources/dos", "CHEAD", 0x47410},
		{"sotsb", "Secret", "/Users/felix/Projects/gobox-engine/data/sotsb/sources/dos", "CBODY", 0x4fdea},
	} {
		t.Run(c.game, func(t *testing.T) {
			s := games[c.game]
			s.Exe, s.Gfx = c.exe, map[string]Gfx{strings.ToLower(c.name): {Pal: c.pal}}
			out := importGame(t, s)
			am, err := dax.Open(filepath.Join(out, c.name+".DAX"))
			if err != nil {
				t.Fatal(err)
			}
			pal, err := dax.Open(filepath.Join(out, c.name+"PAL.DAX"))
			if err != nil {
				t.Fatal(err)
			}
			dos, err := dax.Open(filepath.Join(c.dos, c.name+".DAX"))
			if err != nil {
				t.Skip("no DOS data")
			}
			n := 0
			for _, e := range dos.Entries {
				a, aerr := dax.ParseImage(am.BlockShared(e.ID), false, 0)
				d, derr := dax.ParseImage(dos.BlockShared(e.ID), false, 0)
				p := pal.BlockShared(e.ID)
				if aerr != nil || derr != nil || len(a.Data) != len(d.Data) || len(p) != 48 {
					t.Fatalf("block %d: no picture or palette", e.ID)
				}
				for i := 0; i < len(a.Data); i += 97 {
					c := int(a.Data[i]) - ega.AmigaBase
					got, want := p[c*3:c*3+3], ega.Palette[d.Data[i]&15]
					for k := range 3 {
						if diff := int(got[k]) - int(want[k]); diff < -0x40 || diff > 0x40 {
							t.Fatalf("block %d pixel %d: %x, DOS %x", e.ID, i, got, want)
						}
					}
					n++
				}
			}
			if n < 100 {
				t.Errorf("only %d pixels checked", n)
			}
		})
	}
}
