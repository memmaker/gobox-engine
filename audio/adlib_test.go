package audio

import (
	"crypto/sha256"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

func adlibData(t *testing.T) *AdLib {
	m, _ := os.ReadFile("testdata/MUSIC.DAX")
	z, _ := os.ReadFile("testdata/ZOUNDS.DAX")
	fx := map[int]int{}
	for n := 2; n <= 12; n++ {
		fx[n] = n - 1
	}
	a, err := NewAdLib(AdLibData{Music: m, Sounds: z, Effects: fx, Rate: 9000})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// The sequencer's writes on MUSIC block 1 equal podmusic.py regs, tick/reg/val, for 300 ticks.
func TestAdLibRegs(t *testing.T) {
	a := adlibData(t)
	var got []string
	now := 0
	s := newSequencer(a.music[1], func(r int, v byte) { got = append(got, fmt.Sprintf("%d %d %d", now, r, v)) })
	for now < 300 {
		now = s.now + 1
		if !s.tick() {
			break
		}
	}
	b, _ := os.ReadFile("testdata/pod_regs1.txt")
	want := strings.Fields(string(b))
	var w []string
	for i := 0; i+2 < len(want); i += 3 {
		w = append(w, strings.Join(want[i:i+3], " "))
	}
	if len(got) != len(w) {
		t.Fatalf("%d writes, want %d", len(got), len(w))
	}
	for i := range w {
		if got[i] != w[i] {
			t.Fatalf("write %d: %s, want %s", i, got[i], w[i])
		}
	}
}

func TestAdLibPCM(t *testing.T) {
	a := adlibData(t)
	if !a.Has(1) || a.Has(0) || a.Has(99) {
		t.Fatal("Has")
	}
	a.Start(1)
	m := make([]byte, 44100*4) // 1 s
	a.Mix(m, false)
	if !a.Playing() || a.Song() != 1 {
		t.Fatal("not playing")
	}
	nz := false
	for _, v := range m {
		nz = nz || v != 0
	}
	if !nz {
		t.Fatal("music silent")
	}
	a.Reset()
	if !a.Effect(2) || a.Effect(99) {
		t.Fatal("Effect")
	}
	clear(m)
	a.Mix(m, false)
	nz = false
	for _, v := range m {
		nz = nz || v != 0
	}
	if !nz {
		t.Fatal("effect silent")
	}
}

// PoD's VACE songs (P1.SMS, P3.SMS) are song blocks: Named finds them, they
// parse to the expected tracks and tempo, and the first second renders the same.
func TestAdLibNamedSMS(t *testing.T) {
	dir := os.Getenv("HOME") + "/Projects/pod-go/data/dos/"
	named := map[string][]byte{}
	for _, n := range []string{"P1", "P3"} {
		b, err := os.ReadFile(dir + n + ".SMS")
		if err != nil {
			t.Skip("no Pools of Darkness files")
		}
		named[n] = b
	}
	music, _ := os.ReadFile("testdata/MUSIC.DAX")
	a, err := NewAdLib(AdLibData{Music: music, Named: named})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name   string
		div    int
		events []int
		hash   string
	}{
		{"P1", 198, []int{878, 330, 337}, "893dda92c2108d46"},
		{"P3", 167, []int{146, 156, 114}, "aad60885ff9b6a0f"},
	} {
		n := a.Named(c.name)
		if n == 0 {
			t.Fatalf("%s: not a song", c.name)
		}
		s := newSequencer(named[c.name], func(int, byte) {})
		var ev []int
		for _, tr := range s.tr {
			k := 0
			for p := 0; p+1 < len(tr.ev) && tr.ev[p] != 0xFF; p += 2 {
				k++
			}
			ev = append(ev, k)
		}
		a.Start(n)
		m := make([]byte, 4*adlibRate)
		a.Mix(m, false)
		h := fmt.Sprintf("%x", sha256.Sum256(m))[:16]
		t.Logf("%s: block %d div %d events %v hash %s", c.name, n, s.div, ev, h)
		if !slices.ContainsFunc(m, func(b byte) bool { return b != 0 }) {
			t.Errorf("%s: silent", c.name)
		}
		if s.div != c.div || fmt.Sprint(ev) != fmt.Sprint(c.events) || h != c.hash {
			t.Errorf("%s: div %d events %v hash %s, want %d %v %s", c.name, s.div, ev, h, c.div, c.events, c.hash)
		}
	}
	if a.Named("P2") != 0 {
		t.Error("P2 named")
	}
}
