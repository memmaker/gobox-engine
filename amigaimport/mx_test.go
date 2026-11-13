package amigaimport

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/asset/orig"
)

// SotSB's Amiga music and effects: Spec.Files copies SECRET.MX and
// sounds.glb off disk 1, and amiga.MX renders them.
func TestSotSBMusic(t *testing.T) {
	out := importGame(t, Spec{Disks: []string{"Secret 1", "Secret 2"}, Files: []string{"SECRET.MX", "sounds.glb"}})
	mx, err1 := os.ReadFile(filepath.Join(out, "SECRET.MX"))
	glb, err2 := os.ReadFile(filepath.Join(out, "sounds.glb"))
	if err1 != nil || err2 != nil || len(mx) != 20164 {
		t.Fatalf("files: %v %v, SECRET.MX %d bytes (want 20164)", err1, err2, len(mx))
	}
	const rate = 44100
	song := func(n int) (secs int, pcm []byte) {
		m, err := orig.NewMX(mx, glb)
		if err != nil {
			t.Fatal(err)
		}
		if m.Len() != 7 {
			t.Fatalf("%d songs, want 7", m.Len())
		}
		m.Song(n)
		buf := make([]byte, 4*rate)
		for secs = 0; m.Sounding() && secs < 600; secs++ {
			m.Render(buf, rate)
			pcm = append(pcm, buf...)
		}
		return secs, pcm
	}
	for n := 1; n <= 7; n++ {
		secs, pcm := song(n)
		t.Logf("song %d: %d s", n, secs)
		if secs == 0 || secs >= 600 || !slices.ContainsFunc(pcm, func(b byte) bool { return b != 0 }) {
			t.Errorf("song %d: %d s, silent %v", n, secs, !slices.ContainsFunc(pcm, func(b byte) bool { return b != 0 }))
		}
		if n == 1 {
			if _, again := song(1); !bytes.Equal(pcm, again) {
				t.Error("song 1 not deterministic")
			}
		}
	}
	m, _ := orig.NewMX(mx, glb)
	for e := range 11 {
		m.Effect(e)
		buf := make([]byte, 4*rate)
		m.Render(buf, rate)
		if !slices.ContainsFunc(buf, func(b byte) bool { return b != 0 }) {
			t.Errorf("effect %d silent", e)
		}
	}
}
