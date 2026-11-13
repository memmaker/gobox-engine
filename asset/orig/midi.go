package orig

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// A GLBR song (glbr.go) as data: its notes as a Standard MIDI File and
// its eight instruments as OPLVoice records (the sidecar <song>.opl.json).
// The MIDI's channel n is the song's track n; a program change p on it
// selects OPLVoice p (the driver's 0x81: track n takes instrument n), never a
// General MIDI program. Times are the driver's 72.8227 Hz PIT ticks.

// OPLVoice is one GLBR instrument: its 12 bytes as the driver programs them
// (Sequencer.program, note).
type OPLVoice struct {
	Lead               uint8    // byte 0, not read by the driver
	FixedChannel       uint8    // the OPL channel it must sound on (0: any free one)
	Characteristic     [2]uint8 // registers 0x20 (modulator, carrier): AM, vibrato, sustain, KSR, multiple
	Level              [2]uint8 // 0x40: key scale level, total level (the carrier's scaled by velocity)
	AttackDecay        [2]uint8 // 0x60
	SustainRelease     [2]uint8 // 0x80
	FeedbackConnection uint8    // 0xC0
	Tail               uint8    // byte 11, not read by the driver
}

func voiceOf(r []byte) OPLVoice {
	return OPLVoice{r[0], r[1], [2]uint8{r[2], r[3]}, [2]uint8{r[4], r[5]}, [2]uint8{r[6], r[7]}, [2]uint8{r[8], r[9]}, r[10], r[11]}
}

func (v OPLVoice) bytes() []byte {
	return []byte{v.Lead, v.FixedChannel, v.Characteristic[0], v.Characteristic[1], v.Level[0], v.Level[1],
		v.AttackDecay[0], v.AttackDecay[1], v.SustainRelease[0], v.SustainRelease[1], v.FeedbackConnection, v.Tail}
}

// scoreKind is what a score event does to its track's voice.
type scoreKind int

const (
	scoreOn      scoreKind = iota // note Note at velocity Vel
	scoreOff                      // release the note sounding
	scoreProgram                  // take instrument Vel (setInst)
)

// scoreEvent is one call the sequencer makes on track Track's voice at PIT tick Tick.
type scoreEvent struct {
	Tick, Track int
	Kind        scoreKind
	Note, Vel   int
}

// midiDivision is the MIDI file's ticks per quarter note; midiTempo makes one
// MIDI tick one PIT tick (16384/1193182 s).
const (
	midiDivision = 100
	midiTempo    = midiDivision * 16384 * 1000000 / 1193182 // µs per quarter
)

// ErrNotSong: the bytes are no GLBR song the OPL sequencer plays.
var ErrNotSong = errors.New("not a GLBR song")

// SongMIDI is GLBR song b (tempo divisor div, as AdLibData.Tempo) as a MIDI
// file and its voices.
func SongMIDI(b []byte, div int) ([]byte, []OPLVoice, error) {
	var ev []scoreEvent
	s := NewSequencer(b, div, func(int, byte) {}, false)
	if s == nil {
		return nil, nil, ErrNotSong
	}
	s.rec = func(e scoreEvent) { e.Tick = s.now; ev = append(ev, e) }
	for s.Tick() {
	}
	voices := make([]OPLVoice, len(s.inst))
	for i, r := range s.inst {
		voices[i] = voiceOf(r)
	}
	return writeMIDI(ev, s.now), voices, nil
}

// writeMIDI is events ev as a format-0 MIDI file ending at tick end.
func writeMIDI(ev []scoreEvent, end int) []byte {
	var tr bytes.Buffer
	vlq := func(n int) {
		var b []byte
		b = append(b, byte(n&0x7F))
		for n >>= 7; n > 0; n >>= 7 {
			b = append([]byte{byte(n&0x7F | 0x80)}, b...)
		}
		tr.Write(b)
	}
	vlq(0)
	tempo := uint32(midiTempo)
	tr.Write([]byte{0xFF, 0x51, 3, byte(tempo >> 16), byte(tempo >> 8), byte(tempo)})
	last := map[int]int{} // track -> note sounding
	now := 0
	for _, e := range ev {
		vlq(e.Tick - now)
		now = e.Tick
		ch := byte(e.Track)
		switch e.Kind {
		case scoreOn:
			tr.Write([]byte{0x90 | ch, byte(e.Note), byte(e.Vel)})
			last[e.Track] = e.Note
		case scoreOff:
			tr.Write([]byte{0x80 | ch, byte(last[e.Track]), 0})
		case scoreProgram:
			tr.Write([]byte{0xC0 | ch, byte(e.Vel)})
		}
	}
	vlq(end - now)
	tr.Write([]byte{0xFF, 0x2F, 0})
	var f bytes.Buffer
	f.WriteString("MThd")
	binary.Write(&f, binary.BigEndian, uint32(6))
	binary.Write(&f, binary.BigEndian, [3]uint16{0, 1, midiDivision}) // format 0, one track
	f.WriteString("MTrk")
	binary.Write(&f, binary.BigEndian, uint32(tr.Len()))
	f.Write(tr.Bytes())
	return f.Bytes()
}

// readMIDI is writeMIDI's events and end tick back.
func readMIDI(m []byte) (ev []scoreEvent, end int, err error) {
	if len(m) < 22 || string(m[:4]) != "MThd" || string(m[14:18]) != "MTrk" ||
		binary.BigEndian.Uint16(m[10:]) != 1 || binary.BigEndian.Uint16(m[12:]) != midiDivision {
		return nil, 0, fmt.Errorf("not a song MIDI (one track, division %d)", midiDivision)
	}
	tr := m[22:]
	p, now := 0, 0
	vlq := func() int {
		n := 0
		for p < len(tr) {
			c := tr[p]
			p++
			n = n<<7 | int(c&0x7F)
			if c&0x80 == 0 {
				break
			}
		}
		return n
	}
	for p < len(tr) {
		now += vlq()
		if p >= len(tr) {
			break
		}
		st := tr[p]
		switch {
		case st == 0xFF:
			if p+2 >= len(tr) {
				return nil, 0, errors.New("short meta event")
			}
			kind, n := tr[p+1], int(tr[p+2])
			p += 3 + n
			if kind == 0x2F {
				return ev, now, nil
			}
		case st&0xF0 == 0x90 && p+2 < len(tr):
			ev = append(ev, scoreEvent{Tick: now, Track: int(st & 0xF), Kind: scoreOn, Note: int(tr[p+1]), Vel: int(tr[p+2])})
			p += 3
		case st&0xF0 == 0x80 && p+2 < len(tr):
			ev = append(ev, scoreEvent{Tick: now, Track: int(st & 0xF), Kind: scoreOff})
			p += 3
		case st&0xF0 == 0xC0 && p+1 < len(tr):
			ev = append(ev, scoreEvent{Tick: now, Track: int(st & 0xF), Kind: scoreProgram, Vel: int(tr[p+1])})
			p += 2
		default:
			return nil, 0, fmt.Errorf("event %#x at %d", st, p)
		}
	}
	return nil, 0, errors.New("no end of track")
}

// NewMIDI plays song MIDI m with voices on the OPL2 through w, write
// for write as New plays the GLBR song it came from.
func NewMIDI(m []byte, voices []OPLVoice, w func(int, byte)) (*Sequencer, error) {
	ev, end, err := readMIDI(m)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(ev, func(i, j int) bool { return ev[i].Tick < ev[j].Tick })
	s := &Sequencer{w: w}
	for i, v := range voices {
		if i < len(s.inst) {
			s.inst[i] = v.bytes()
		}
	}
	s.initOPL()
	s.score, s.end = ev, end
	return s, nil
}

// MusicDir is where an adventure folder keeps its songs (WriteSongs).
const MusicDir = "sound/music"

// WriteSongs writes each GLBR song block n of DAX file music (tempo divisor
// tempo, as AdLibData.Tempo) into dir as n.mid and n.opl.json (its
// OPLVoices); it returns the blocks written. Blocks that are no song are left
// out.
func WriteSongs(music []byte, tempo int, dir string) ([]int, error) {
	blocks, err := GlbrBlocks(music)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var ids []int
	for id, b := range blocks {
		mid, voices, err := SongMIDI(b, tempo)
		if errors.Is(err, ErrNotSong) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("song %d: %w", id, err)
		}
		js, err := json.MarshalIndent(voices, "", "\t")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprint(id, ".mid")), mid, 0o644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprint(id, ".opl.json")), js, 0o644); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
