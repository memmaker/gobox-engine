package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/fs"
	"os"
	"slices"
	"testing"

	"github.com/memmaker/opl2"

	"gobox/audio/glbr"
)

// render is sequencer s's register writes, tick by tick to its end, and the
// first sec seconds of their PCM (opl2 at adlibRate, as AdLib.Mix).
func render(mk func(w func(int, byte)) *glbr.Sequencer, sec int) (regs []int, pcm []byte) {
	chip := opl2.New()
	rs := opl2.NewResampler(chip, adlibRate)
	tick := 0
	s := mk(func(r int, v byte) { chip.Write(r, v); regs = append(regs, tick, r, int(v)) })
	acc, alive := 0, true
	for i := 0; i < adlibRate*sec; i++ {
		for acc += 1193182; alive && acc >= adlibRate*16384; acc -= adlibRate * 16384 {
			tick++
			if alive = s.Tick(); !alive {
				s.Silence()
			}
		}
		pcm = binary.LittleEndian.AppendUint16(pcm, uint16(int16(rs.Next())))
	}
	for ; alive; tick++ {
		if alive = s.Tick(); !alive {
			s.Silence()
		}
	}
	return regs, pcm
}

// TestMIDIEqualsAdLib: every song of PoD's MUSIC.DAX (and BR1's and BR2's
// when present) as MIDI and OPLVoices plays write for write, and so sample
// for sample, as the GLBR song.
func TestMIDIEqualsAdLib(t *testing.T) {
	for _, f := range []string{"testdata/MUSIC.DAX", "../../buckrogers-go/data/dos/MUSIC.DAX", "../../buckrogers2-go/data/dos/MUSIC.DAX"} {
		dax, err := os.ReadFile(f)
		if errors.Is(err, fs.ErrNotExist) && f != "testdata/MUSIC.DAX" {
			t.Log(f, "missing")
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		music, err := glbr.Blocks(dax)
		if err != nil {
			t.Fatal(f, err)
		}
		n := 0
		for id, b := range music {
			if glbr.New(b, 0, func(int, byte) {}, false) == nil {
				continue
			}
			mid, voices, err := glbr.SongMIDI(b, 0)
			if err != nil {
				t.Fatal(f, id, err)
			}
			wantR, wantP := render(func(w func(int, byte)) *glbr.Sequencer { return glbr.New(b, 0, w, false) }, 5)
			gotR, gotP := render(func(w func(int, byte)) *glbr.Sequencer {
				s, err := glbr.NewMIDI(mid, voices, w)
				if err != nil {
					t.Fatal(f, id, err)
				}
				return s
			}, 5)
			if !slices.Equal(gotR, wantR) || !bytes.Equal(gotP, wantP) {
				t.Errorf("%s song %d: %d register writes, want %d; pcm equal %v", f, id, len(gotR)/3, len(wantR)/3, bytes.Equal(gotP, wantP))
			}
			n++
		}
		if n < 2 {
			t.Fatalf("%s: %d songs", f, n)
		}
	}
}
