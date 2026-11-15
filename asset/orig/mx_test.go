package orig

import (
	"encoding/binary"
	"os"
	"slices"
	"testing"
)

// mxTestInst is an instrument of a synthetic MXTX file: no attack, full
// sustain, an optional release.
type mxTestInst struct {
	slot, octaves, oneshot, repeat int
	rel                            [][2]uint16
}

const mxTestBPM = 120 // 7.68 ticks a frame

// mxTestFile is an MXTX file of one song.
func mxTestFile(flags uint16, insts []mxTestInst, events ...[6]byte) []byte {
	be := binary.BigEndian
	b := be.AppendUint16([]byte("MXTX"), mxTestBPM)
	b = be.AppendUint16(b, flags)
	b = be.AppendUint16(b, 1)
	b = be.AppendUint32(b, uint32(len(events)))
	for _, e := range events {
		b = append(b, e[:]...)
	}
	b = be.AppendUint16(b, uint16(len(insts)))
	for _, in := range insts {
		for _, w := range []int{in.slot, 0, 0x8000, in.octaves} {
			b = be.AppendUint16(b, uint16(w))
		}
		b = be.AppendUint32(b, uint32(in.oneshot))
		b = be.AppendUint32(b, uint32(in.repeat))
		b = be.AppendUint16(b, 0)
		b = be.AppendUint16(b, uint16(len(in.rel)))
		for _, s := range in.rel {
			b = be.AppendUint16(be.AppendUint16(b, s[0]), s[1])
		}
		for k := range in.octaves {
			for i := range (in.oneshot + in.repeat) << k {
				b = append(b, byte(int8(100-(i%8)*25)))
			}
		}
	}
	return b
}

func mxTestEvent(b0, b1 byte, delta, arg int) (e [6]byte) {
	e[0], e[1] = b0, b1
	binary.BigEndian.PutUint16(e[2:], uint16(delta))
	binary.BigEndian.PutUint16(e[4:], uint16(arg))
	return e
}

// mxTestNote: vel is a multiple of 8.
func mxTestNote(ch, note, vel, delta, dur int) [6]byte {
	return mxTestEvent(byte(note), byte(vel<<1|ch), delta, dur)
}

// mxTestPlay starts the file's song and runs frames of it.
func mxTestPlay(t *testing.T, file []byte, frames int) *MX {
	t.Helper()
	m, err := NewMX(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.Song(1)
	for range frames {
		m.frame()
	}
	return m
}

func mxTestNotes(m *MX) (n [4]int) {
	for i, v := range m.v {
		if v.st != mxFree {
			n[i] = v.note
		}
	}
	return n
}

var mxTestPlain = []mxTestInst{{slot: 0, octaves: 1, repeat: 16}, {slot: 1, octaves: 1, repeat: 16}}

// 0x5a252: a side's two voices take turns, a full side sends the note to
// the other, and with all four held the side's turn is stolen.
func TestMXVoiceOrder(t *testing.T) {
	const long = 1000
	m := mxTestPlay(t, mxTestFile(0, mxTestPlain,
		mxTestNote(0, 60, 64, 0, long), mxTestNote(0, 61, 64, 0, long), mxTestNote(0, 62, 64, 0, long),
		mxTestNote(0, 63, 64, 0, long), mxTestNote(0, 64, 64, 0, long)), 1)
	if got, want := mxTestNotes(m), [4]int{61, 63, 62, 64}; got != want {
		t.Errorf("five notes at once on a left channel: voices play %v, want %v", got, want)
	}
	m = mxTestPlay(t, mxTestFile(0, mxTestPlain, mxTestNote(1, 60, 64, 0, long)), 1)
	if got, want := mxTestNotes(m), [4]int{0, 0, 60, 0}; got != want {
		t.Errorf("first note of a right channel: voices play %v, want %v", got, want)
	}
}

// 0x5a252: a releasing voice goes before a held one of its side, and before the other side.
func TestMXVoiceReleasing(t *testing.T) {
	rel := []mxTestInst{{slot: 0, octaves: 1, repeat: 16, rel: [][2]uint16{{5000, 0}}}}
	m := mxTestPlay(t, mxTestFile(0, rel,
		mxTestNote(0, 60, 64, 0, 1), mxTestNote(0, 61, 64, 0, 1000),
		mxTestNote(0, 62, 64, 96, 1000), mxTestNote(0, 63, 64, 0, 1000)), 20)
	if got, want := mxTestNotes(m), [4]int{61, 0, 63, 62}; got != want {
		t.Errorf("voices play %v, want %v", got, want)
	}
}

func mxTestRender(m *MX) (left, right bool) {
	out := make([]byte, 4*2000)
	m.Render(out, 44100)
	for i := 0; i < len(out); i += 4 {
		left = left || out[i] != 0 || out[i+1] != 0
		right = right || out[i+2] != 0 || out[i+3] != 0
	}
	return left, right
}

// 0x5a71c, 0x5a7ee: even channels sound left (voices 0, 3), odd ones right
// (1, 2); a pan controller below the middle is right.
func TestMXPan(t *testing.T) {
	for _, c := range []struct {
		name        string
		events      [][6]byte
		left, right bool
	}{
		{"even channel", [][6]byte{mxTestNote(0, 60, 64, 0, 1000)}, true, false},
		{"odd channel", [][6]byte{mxTestNote(1, 60, 64, 0, 1000)}, false, true},
		{"even channel panned 0", [][6]byte{mxTestEvent(0xb0, 0, 0, mxCCPan<<8), mxTestNote(0, 60, 64, 0, 1000)}, false, true},
		{"odd channel panned 127", [][6]byte{mxTestEvent(0xb0, 1, 0, mxCCPan<<8|127), mxTestNote(1, 60, 64, 0, 1000)}, true, false},
	} {
		m, err := NewMX(mxTestFile(0, mxTestPlain, c.events...), nil)
		if err != nil {
			t.Fatal(err)
		}
		m.Song(1)
		if l, r := mxTestRender(m); l != c.left || r != c.right {
			t.Errorf("%s: sounds left %v right %v, want %v %v", c.name, l, r, c.left, c.right)
		}
	}
}

// mxTestVoice is the one voice playing.
func mxTestVoice(t *testing.T, m *MX) *mxVoice {
	t.Helper()
	i := slices.IndexFunc(m.v[:], func(v mxVoice) bool { return v.st != mxFree })
	if i < 0 {
		t.Fatal("no voice plays")
	}
	return &m.v[i]
}

// 0x5b05c, 0x5a20c: velocity scales the volume only with header flag bit 1.
func TestMXVelocityVolume(t *testing.T) {
	for _, c := range []struct {
		flags     uint16
		vel, want int
	}{{0, 16, 64}, {0, 120, 64}, {mxFlagVel, 16, 64 * 17 >> 7}, {mxFlagVel, 120, 64 * 121 >> 7}} {
		m := mxTestPlay(t, mxTestFile(c.flags, mxTestPlain, mxTestNote(0, 60, c.vel, 0, 1000)), 1)
		if got := mxTestVoice(t, m).vol; got != c.want {
			t.Errorf("flags %#x velocity %d: volume %d, want %d", c.flags, c.vel, got, c.want)
		}
	}
}

// 0x5a7d8, 0x5a20c: controller 7 is the channel's volume (pod.mx sets 0x66).
func TestMXChannelVolume(t *testing.T) {
	m := mxTestPlay(t, mxTestFile(0, mxTestPlain,
		mxTestEvent(0xb0, 0, 0, mxCCVolume<<8|0x66), mxTestNote(0, 60, 64, 0, 1000)), 1)
	if got, want := mxTestVoice(t, m).vol, 64*0x67>>7; got != want {
		t.Errorf("volume %d, want %d", got, want)
	}
}

// 0x5a3ec: a low note takes the next octave chunk at half the period.
func TestMXOctaveChunk(t *testing.T) {
	const low, high = 0x2d, 0x2d + 12
	two := []mxTestInst{{slot: 0, octaves: 2, oneshot: 8, repeat: 8}}
	play := func(insts []mxTestInst, note int) *mxVoice {
		return mxTestVoice(t, mxTestPlay(t, mxTestFile(0, insts, mxTestNote(0, note, 64, 0, 1000)), 1))
	}
	hi, lo, one := play(two, high), play(two, low), play(mxTestPlain, low)
	if hi.oct != 0 || len(hi.data) != 16 || hi.per < 500 || hi.per > 510 {
		t.Errorf("high note: chunk %d of %d bytes, period %d", hi.oct, len(hi.data), hi.per)
	}
	if lo.oct != 1 || len(lo.data) != 32 || lo.end != 16 || lo.end2 != 32 || lo.per != hi.per {
		t.Errorf("low note: chunk %d of %d bytes (one-shot %d), period %d, want chunk 1 of 32 (16), period %d", lo.oct, len(lo.data), lo.end, lo.per, hi.per)
	}
	if one.oct != 0 || one.per < 2*hi.per-1 || one.per > 2*hi.per+1 {
		t.Errorf("low note, one chunk: chunk %d, period %d, want 0, %d", one.oct, one.per, 2*hi.per)
	}
}

// 0x5a900, 0x5a376: a pitch bend moves the sounding note (24 semitones at
// full bend) and leaves it on its octave chunk.
func TestMXPitchBend(t *testing.T) {
	two := []mxTestInst{{slot: 0, octaves: 2, oneshot: 8, repeat: 8}}
	file := mxTestFile(0, two, mxTestNote(0, 0x2d+12, 64, 0, 1000), mxTestEvent(0xe0, 0, 48, 0))
	before := mxTestVoice(t, mxTestPlay(t, file, 1)).per
	v := mxTestVoice(t, mxTestPlay(t, file, 10))
	if v.oct != 0 || v.per < 4*before-3 || v.per > 4*before+3 {
		t.Errorf("bent two octaves down: chunk %d, period %d, want 0, %d", v.oct, v.per, 4*before)
	}
}

// 0x5b560: a tempo change over 0x8000 ticks or more is at once (signed
// compare), a shorter one ramps.
func TestMXTempo(t *testing.T) {
	m := mxTestPlay(t, mxTestFile(0, mxTestPlain, mxTestEvent(0x80, 140, 0, 64000)), 1)
	if m.tempo != 140<<4 {
		t.Errorf("tempo over 64000 ticks: %d/16 after a frame, want 140", m.tempo)
	}
	file := mxTestFile(0, mxTestPlain, mxTestEvent(0x80, 140, 0, 96), mxTestEvent(0xff, 0, 1000, 0))
	m = mxTestPlay(t, file, 5)
	if m.tempo <= mxTestBPM<<4 || m.tempo >= 140<<4 {
		t.Errorf("tempo over 96 ticks: %d/16 after 5 frames, want between 120 and 140", m.tempo)
	}
	if m = mxTestPlay(t, file, 20); m.tempo != 140<<4 {
		t.Errorf("tempo over 96 ticks: %d/16 after 20 frames, want 140", m.tempo)
	}
}

// 0x5b6be: meta 2 n .. meta 3 plays the part n more times.
func TestMXLoop(t *testing.T) {
	frames := func(events ...[6]byte) (n int) {
		m, err := NewMX(mxTestFile(0, mxTestPlain, events...), nil)
		if err != nil {
			t.Fatal(err)
		}
		for m.Song(1); m.Sounding() && n < 1000; n++ {
			m.frame()
		}
		return n
	}
	note, end := mxTestNote(0, 60, 64, 96, 10), mxTestEvent(0xff, 0, 0, 0)
	once := frames(note, end)
	thrice := frames(mxTestEvent(0xa0, 0, 0, mxLoopStart<<8|2), note, mxTestEvent(0xa0, 0, 0, mxLoopEnd<<8), end)
	if once != 13 || thrice != 3*once {
		t.Errorf("a 96-tick part takes %d frames, repeated twice %d; want 13, 39", once, thrice)
	}
}

// A second of every song of the real files plays.
func TestMXRealSongs(t *testing.T) {
	for _, path := range []string{"../../data/sotsb/sources/amiga/SECRET.MX", "../../data/pod/sources/amiga/pod.mx"} {
		t.Run(path, func(t *testing.T) {
			mx, err := os.ReadFile(path)
			if err != nil {
				t.Skip(err)
			}
			m, err := NewMX(mx, nil)
			if err != nil {
				t.Fatal(err)
			}
			if m.Len() == 0 {
				t.Fatal("no songs")
			}
			const rate = 44100
			for n := 1; n <= m.Len(); n++ {
				m.Song(n)
				out := make([]byte, 4*rate)
				m.Render(out, rate)
				if !slices.ContainsFunc(out, func(b byte) bool { return b != 0 }) {
					t.Errorf("song %d: silent", n)
				}
			}
		})
	}
}
