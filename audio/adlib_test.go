package audio

import (
	"fmt"
	"os"
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
	a, err := NewAdLib(AdLibData{Music: m, Sounds: z, Effects: fx, Header: true, Skip: []int{4}, Rate: 9000})
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
