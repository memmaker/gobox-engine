package audio

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"gobox/asset/orig"
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

// fakeTracks: song 27 is a one-shot that ends at once.
type fakeTracks struct{ song, starts int }

func (f *fakeTracks) Len() int             { return 2 }
func (f *fakeTracks) Has(n int) bool       { return true }
func (f *fakeTracks) Song() int            { return f.song }
func (f *fakeTracks) Reset()               { f.song = 0 }
func (f *fakeTracks) Start(n int)          { f.song = n; f.starts++ }
func (f *fakeTracks) Playing() bool        { return f.song != 27 }
func (f *fakeTracks) Mix(m []byte, _ bool) { clear(m) }

// A one-shot gives way to the song that follows it; again restarts it.
func TestSongNext(t *testing.T) {
	d, err := NewDriver(porData, 'T')
	if err != nil {
		t.Fatal(err)
	}
	tr := &fakeTracks{}
	d.SetSoundtrack(true, 0, 0)
	d.Attach(nil, nil, tr, nil, nil, nil)
	if !d.Song(27, 11, true) || !d.Song(27, 11, true) || tr.starts != 2 {
		t.Fatalf("again: %d starts", tr.starts)
	}
	d.Read(make([]byte, 400))
	if on, song := d.Track(); !on || song != 11 || tr.song != 11 {
		t.Fatalf("after the one-shot: on %v song %d", on, song)
	}
}

// A Music-X player (amiga.MX) plays through the Driver's Amiga hooks; a nil
// *amiga.Player (por without SOUND.BIN) leaves the DOS driver.
func TestDriverMX(t *testing.T) {
	d := newDriver(t, 'T')
	d.Attach((*orig.Player)(nil), nil, nil, nil, nil, nil)
	d.Effect(3, false)
	if d.am != nil {
		t.Fatal("nil *amiga.Player kept")
	}
	// one song: note 60 (channel 0, velocity 0x40) for a quarter, end; one looped 32-byte square instrument
	mx := []byte("MXTX\x00\x78\x00\x00\x00\x01\x00\x00\x00\x02" +
		"\x3c\x80\x00\x00\x00\xc0" + "\xff\x00\x00\xc0\x00\x00" +
		"\x00\x01" + "\x00\x00\x00\x00\x80\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x20\x00\x00\x00\x00")
	for i := range 32 {
		mx = append(mx, byte(int8(min(1, max(-1, 16-i))*100)))
	}
	glb := []byte("GLIB\x00\x00\x00\x00\x00\x01\x00\x00DIG8\x00\x00\x00\x18\x00\x00\x00\x20" + "\x1c\xfa\x00\x00\x00\x02\x40\x40")
	m, err := orig.NewMX(mx, glb)
	if err != nil {
		t.Fatal(err)
	}
	d.Attach(m, []int{0}, nil, nil, nil, nil)
	d.AmigaSong(1)
	b := make([]byte, 4*sndRate/2)
	d.Read(b)
	if !loud(b) || !d.Sounding() {
		t.Errorf("song: loud %v, sounding %v", loud(b), d.Sounding())
	}
	d.Read(b) // the song's end (0xff) is a quarter (0.5 s at 120) in
	if d.Sounding() {
		t.Error("song still going after its end")
	}
	d.AmigaSong(0)
	d.Effect(2, false) // effect 2 -> GLIB entry 0
	d.Read(b[:16])
	if !loud(b[:16]) {
		t.Error("effect silent")
	}
}
