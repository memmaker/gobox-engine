package vace

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var dosDir = filepath.Join("..", "..", "data", "pod", "sources", "dos")

func load(t *testing.T, base string) *Anim {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dosDir, base+".BMA")); err != nil {
		t.Skip("no Pools of Darkness files")
	}
	a, err := Load(dosDir, base)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func hash(f Frame) string {
	h := sha256.New()
	h.Write(f.Image.Pix)
	for _, c := range f.Image.Palette {
		r, g, b, _ := c.RGBA()
		fmt.Fprint(h, r>>8, g>>8, b>>8)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func TestAnims(t *testing.T) {
	for _, c := range []struct {
		base   string
		frames int
		hashes map[int]string
		song   string
	}{
		{"INTRO", 133, map[int]string{0: "5e20607fa120986a", 10: "4547d2b13f46fad7", 132: "ba24034af085fff7"}, "P1"},
		{"FINAL", 58, map[int]string{0: "af0169acbb7e59e5", 10: "e2156e32f8b5b2e6", 57: "40eef4d3d885a26f"}, "P3"},
	} {
		a := load(t, c.base)
		if len(a.Frames) != c.frames {
			t.Errorf("%s: %d frames, want %d", c.base, len(a.Frames), c.frames)
		}
		if b := a.Frames[0].Image.Bounds(); b.Dx() != 320 || b.Dy() != 200 {
			t.Errorf("%s: size %v", c.base, b)
		}
		for i, want := range c.hashes {
			if got := hash(a.Frames[i]); got != want {
				t.Errorf("%s frame %d: hash %s, want %s", c.base, i, got, want)
			}
		}
		if len(a.Cues) == 0 || a.Cues[0].Song != c.song {
			t.Errorf("%s: cues %v", c.base, a.Cues)
		}
		t.Logf("%s: %d frames, %v, cues %v, hashes 0:%s 10:%s last:%s", c.base, len(a.Frames), a.Len().Round(time.Second), a.Cues, hash(a.Frames[0]), hash(a.Frames[10]), hash(a.Frames[len(a.Frames)-1]))
	}
}

func TestSong(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(dosDir, "P1.SMS"))
	if err != nil {
		t.Skip("no Pools of Darkness files")
	}
	ch, err := Song(b)
	if err != nil || len(ch) != 2 || ch[0].Tag != "GLBR" || ch[1].Tag != "SLBR" {
		t.Fatalf("%v %v", ch, err)
	}
}
