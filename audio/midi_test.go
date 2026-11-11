package audio

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"

	"github.com/memmaker/opl2"
)

// render is sequencer s's register writes, tick by tick to its end, and the
// first sec seconds of their PCM (opl2 at adlibRate, as AdLib.Mix).
func render(mk func(w func(int, byte)) *sequencer, sec int) (regs []int, pcm []byte) {
	chip := opl2.New()
	rs := opl2.NewResampler(chip, adlibRate)
	tick := 0
	s := mk(func(r int, v byte) { chip.Write(r, v); regs = append(regs, tick, r, int(v)) })
	acc, alive := 0, true
	for i := 0; i < adlibRate*sec; i++ {
		for acc += 1193182; alive && acc >= adlibRate*16384; acc -= adlibRate * 16384 {
			tick++
			if alive = s.tick(); !alive {
				s.silence()
			}
		}
		pcm = binary.LittleEndian.AppendUint16(pcm, uint16(int16(rs.Next())))
	}
	for ; alive; tick++ {
		if alive = s.tick(); !alive {
			s.silence()
		}
	}
	return regs, pcm
}

// TestMIDIEqualsAdLib: every song of PoD's MUSIC.DAX as MIDI and OPLVoices
// plays write for write, and so sample for sample, as the GLBR song.
func TestMIDIEqualsAdLib(t *testing.T) {
	a := adlibData(t)
	n := 0
	for id, b := range a.music {
		if newSequencer(b, 0, func(int, byte) {}, false) == nil {
			continue
		}
		mid, voices, err := SongMIDI(b, 0)
		if err != nil {
			t.Fatal(id, err)
		}
		wantR, wantP := render(func(w func(int, byte)) *sequencer { return newSequencer(b, 0, w, false) }, 5)
		gotR, gotP := render(func(w func(int, byte)) *sequencer {
			s, err := newMIDISequencer(mid, voices, w)
			if err != nil {
				t.Fatal(id, err)
			}
			return s
		}, 5)
		if !slices.Equal(gotR, wantR) || !bytes.Equal(gotP, wantP) {
			t.Errorf("song %d: %d register writes, want %d; pcm equal %v", id, len(gotR)/3, len(wantR)/3, bytes.Equal(gotP, wantP))
		}
		n++
	}
	if n < 2 {
		t.Fatalf("%d songs", n)
	}
}
