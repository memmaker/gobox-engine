package audio

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"testing"
)

// por's tables (por rules/sounddata.json).
//
//go:embed testdata/por.json
var porData []byte

func newDriver(t *testing.T, card byte) *Driver {
	d, err := NewDriver(porData, card)
	if err != nil {
		t.Fatal(err)
	}
	d.Attach(nil, nil, nil, nil, nil, nil)
	return d
}

func loud(b []byte) bool {
	for i := 0; i+1 < len(b); i += 4 {
		if b[i] != 0 || b[i+1] != 0 {
			return true
		}
	}
	return false
}

// soundHash is the SHA-256 of one second of por's every sound on the speaker
// and the Tandy: the synth's output, pinned across refactors (was por's).
const soundHash = "8cba73fa61ee71685ec4925b0d66afb22bcb08d2b5cde816ed53eb5fa38ef530"

func TestSoundHash(t *testing.T) {
	h := sha256.New()
	for _, card := range []byte{'P', 'T'} {
		for n := 2; n <= 12; n++ {
			d := newDriver(t, card)
			d.Wait(0)
			d.Effect(n, false)
			buf := make([]byte, 4*sndRate)
			d.Read(buf)
			h.Write(buf)
		}
	}
	if got := fmt.Sprintf("%x", h.Sum(nil)); got != soundHash {
		t.Errorf("synth output hash %s, want %s", got, soundHash)
	}
}

// A song started on the Tandy set reaches the SN76489, None mutes it.
func TestTandySound(t *testing.T) {
	d := newDriver(t, 'T')
	d.Wait(0)
	d.Effect(1, false)
	d.Effect(2, false)
	buf := make([]byte, 4*sndRate/2)
	d.Read(buf)
	if !loud(buf) {
		t.Fatal("Tandy song is silent")
	}
	d.Effect(255, false)
	d.SetCard(None)
	if d.syn.Song() != 0 {
		t.Fatalf("stop: song %d", d.syn.Song())
	}
}

// Every sound por plays is audible on the PC speaker and on the Tandy.
func TestEverySoundAudible(t *testing.T) {
	for _, card := range []byte{'P', 'T'} {
		for n := 2; n <= 12; n++ {
			d := newDriver(t, card)
			d.Wait(0)
			d.Effect(n, false)
			buf := make([]byte, 4*sndRate)
			d.Read(buf)
			if !loud(buf) {
				t.Errorf("sound %d on %c is silent", n, card)
			}
		}
	}
}

// In the cut mode a sound stops on frame rate/10, whatever the buffers.
func TestCutAfter(t *testing.T) {
	for _, chunk := range []int{4410, 1000, 4096, 8820} {
		d := newDriver(t, 'P')
		d.Wait(10)
		d.Effect(2, false) // a spell flies: sounds for longer than 0.1 s
		d.CutAfter(sndRate / 10)
		out := make([]byte, 0, 4*sndRate)
		for len(out) < 4*sndRate/2 {
			b := make([]byte, 4*chunk)
			d.Read(b)
			out = append(out, b...)
		}
		if !loud(out[:4*sndRate/10]) {
			t.Fatalf("chunk %d: silent before the cut", chunk)
		}
		if loud(out[4*sndRate/10:]) {
			t.Errorf("chunk %d: sound after frame %d", chunk, sndRate/10)
		}
	}
}
