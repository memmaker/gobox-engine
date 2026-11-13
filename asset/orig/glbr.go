// glbr.go is SOUND.COM's AdLib song sequencer (GLBR/SLBR song blocks,
// pod-go docs/MUSIC.md) and its SID twin, as register writes, and the songs
// as data (MIDI plus OPLVoice records, midi.go). It needs only the standard
// library, so importers use it; gobox/audio puts the writes on the chips.

package orig

import "encoding/binary"

// The AdLib half of SOUND.COM (see pod-go docs/MUSIC.md), register for register.
var (
	adOpSlot = [...]int{0, 0, 1, 2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 16, 17, 18, 19, 20, 21}
	adMap    = [...]int{1, 2, 3, 7, 8, 9, 13, 17, 15, 18, 14}
	adFnum   = [...]int{343, 365, 385, 408, 433, 459, 486, 515, 546, 579, 614, 650}
)

type adTrack struct {
	next, p, rem, vel, st int
	ev                    []byte
}

type Sequencer struct {
	w     func(reg int, val byte)
	sid   *sidDev // the Innovation SSI-2001's SID in the OPL's place (nil: AdLib)
	inst  [8][]byte
	tr    []adTrack
	div   int
	now   int
	mask  int
	oplch [6]int
	ins   [6]int

	rec   func(scoreEvent) // SongMIDI: each call on a track's voice
	score []scoreEvent     // a song MIDI's calls, played in place of tr (NewMIDI)
	end   int              // ... and the tick it ends on
}

// GlbrNew parses a song block (song 0) and writes the init registers; nil if malformed.
// The SLBR starts where the GLBR header's u32 says (PoD 0x1BC, SotSB 0x1B2). div is the
// tempo divisor (SotSB's driver: 0x76); 0 takes the song header's tempo word / 84 (PoD).
func GlbrNew(b []byte, div int, w func(int, byte), sid bool) *Sequencer {
	if len(b) < 12 || string(b[:4]) != "GLBR" {
		return nil
	}
	so := int(binary.BigEndian.Uint32(b[4:]))
	if so < 12 || so+0x0C > len(b) || string(b[so:so+4]) != "SLBR" {
		return nil
	}
	cnt := int(binary.BigEndian.Uint16(b[10:]))
	if cnt < 8 {
		return nil
	}
	s := &Sequencer{w: w}
	off := 12 + 2*cnt
	for i := 0; i < 8; i++ {
		if off+12 > len(b) {
			return nil
		}
		s.inst[i] = b[off : off+12]
		off += int(binary.BigEndian.Uint16(b[12+2*i:]))
	}
	sl := b[so:]
	if len(sl) < 0x0C {
		return nil
	}
	base := 0x0A + 2*(int(sl[8])+1)
	if base+0x0E > len(sl) {
		return nil
	}
	h := sl[base+int(binary.BigEndian.Uint16(sl[0x0A:])):]
	if len(h) < 0x0E {
		return nil
	}
	if s.div = div; div == 0 {
		s.div = int(binary.BigEndian.Uint16(h)) / 84
	}
	if s.div == 0 {
		return nil
	}
	vel, nt := int(h[2]), int(h[3])
	for t := 1; t < min(nt, 6); t++ {
		o := 0x0E + int(binary.BigEndian.Uint16(h[4+2*t:]))
		if o >= len(h) {
			return nil
		}
		s.tr = append(s.tr, adTrack{next: 1, vel: vel, st: 3, ev: h[o:]})
	}
	if sid {
		if s.sid = newSIDDev(b, cnt); s.sid == nil {
			return nil
		}
		s.sid.init(w)
		return s
	}
	s.initOPL()
	return s
}

// initOPL writes the driver's init registers.
func (s *Sequencer) initOPL() {
	s.w(0xBD, 0x20)
	s.w(0x08, 0x40)
	for k := 1; k < 19; k++ {
		s.w(0xE0+adOpSlot[k], 0)
	}
	s.w(0x01, 0)
}

func (s *Sequencer) alloc(i int) int {
	if f := int(s.inst[i][1]); f != 0 {
		s.mask |= 1 << f
		return f
	}
	n := 0
	for s.mask&(1<<n) != 0 && n <= 6 {
		n++
	}
	s.mask |= 1 << n
	return n
}

func (s *Sequencer) program(ch, i int) {
	r := s.inst[i]
	ops := []int{adOpSlot[adMap[ch]], adOpSlot[adMap[ch]+3]} // ponytail: rhythm voices (ch>6) unused by data, not ported
	for k, o := range ops {
		s.w(0xBD, 0x20)
		s.w(0x08, 0x40)
		s.w(0x40+o, r[4+k])
		if k == 0 {
			s.w(0xC0+ch, r[10])
		}
		s.w(0x60+o, r[6+k])
		s.w(0x80+o, r[8+k])
		s.w(0x20+o, r[2+k])
	}
}

func (s *Sequencer) setInst(t, i int) {
	if s.sid != nil {
		s.sid.setInst(s.w, t, i)
		return
	}
	if s.rec != nil {
		s.rec(scoreEvent{Track: t, Kind: scoreProgram, Vel: i})
	}
	s.mask &^= 1 << s.oplch[t]
	c := s.alloc(i)
	if c > 6 {
		c = 6 // ponytail: fixed rhythm channels unsupported
	}
	s.program(c, i)
	s.oplch[t], s.ins[t] = c, i
}

// note sounds (vel >= 0) or releases (vel < 0) slot t.
func (s *Sequencer) note(t, note, vel int) {
	if s.sid != nil {
		s.sid.note(s.w, t, note, vel)
		return
	}
	if s.rec != nil {
		if vel < 0 {
			s.rec(scoreEvent{Track: t, Kind: scoreOff})
		} else {
			s.rec(scoreEvent{Track: t, Kind: scoreOn, Note: note, Vel: vel})
		}
	}
	if vel >= 0 && t >= 3 {
		vel = vel * 100 >> 7
	}
	c, r := s.oplch[t], s.inst[s.ins[t]]
	s.w(0xA0+c, 0x4A)
	s.w(0xB0+c, 0x0E)
	if vel < 0 {
		return
	}
	tl := int(r[5])
	s.w(0x40+adOpSlot[adMap[c]+3], byte((0x3F-((0x3F-tl&0x3F)*vel>>7))|tl&0xC0))
	oc := min(max(note/12-1, 0), 7)
	f := adFnum[note%12]
	s.w(0xA0+c, byte(f))
	s.w(0xB0+c, byte(f>>8&3|oc<<2|0x20))
}

func dur(code int) int {
	x := 0x6900 >> (code & 7)
	if code&8 != 0 {
		x = (x >> 1) * 3
	}
	if code&0x30 != 0 {
		n := (code & 0x30) >> 3
		x = x / (n + 1) * n
	}
	return x
}

// Silence keys every voice off.
func (s *Sequencer) Silence() {
	if s.sid != nil {
		s.sid.silence(s.w)
		return
	}
	for c := 0; c < 9; c++ {
		s.w(0xB0+c, 0x0E)
	}
	s.w(0xBD, 0x20)
}

// Tick advances one 72.8 Hz PIT tick (the sequencer runs on even ones); false once every
// track has ended (the caller silences).
func (s *Sequencer) Tick() bool {
	s.now++
	if s.score != nil || s.end != 0 {
		return s.replay()
	}
	if s.now&1 != 0 {
		return true
	}
	for i := range s.tr {
		t := &s.tr[i]
		if t.next == 0 || t.next > s.now {
			continue
		}
		if t.st == 0 {
			s.note(i, 0, -1)
			t.st = 3
			t.next++
			continue
		}
		for t.p+1 < len(t.ev) && t.ev[t.p] > 0x80 && t.ev[t.p] < 0xFF {
			switch t.ev[t.p] {
			case 0x81:
				if t.st < 3 {
					s.note(i, 0, -1)
				}
				t.st = 3
				s.setInst(i, i)
			case 0x84:
				t.vel = 0x7F
			}
			t.p += 2
		}
		if t.p+1 >= len(t.ev) || t.ev[t.p] == 0xFF {
			t.next = 0
			if t.st < 3 {
				s.note(i, 0, -1)
			}
			t.st = 3
			continue
		}
		b0, b1 := int(t.ev[t.p]), int(t.ev[t.p+1])
		d := dur(b1)
		t.next += d / s.div
		if t.rem += d % s.div; t.rem >= s.div {
			t.next++
			t.rem -= s.div
		}
		if b0 == 0x80 {
			if t.st < 3 {
				s.note(i, 0, -1)
			}
			t.st = 3
		} else {
			if t.st != 2 {
				if t.st < 3 {
					s.note(i, 0, -1)
				}
				s.note(i, b0, t.vel)
			}
			switch {
			case b1&0x40 != 0:
				t.st = 2
			case t.p+2 < len(t.ev) && int(t.ev[t.p+2]) == b0:
				t.next--
				t.st = 0
			default:
				t.st = 1
			}
		}
		t.p += 2
	}
	for _, t := range s.tr {
		if t.next != 0 {
			return true
		}
	}
	return false
}

// replay makes this tick's calls of a song MIDI's score; false at its end.
func (s *Sequencer) replay() bool {
	for len(s.score) > 0 && s.score[0].Tick <= s.now {
		e := s.score[0]
		s.score = s.score[1:]
		switch e.Kind {
		case scoreOn:
			s.note(e.Track, e.Note, e.Vel)
		case scoreOff:
			s.note(e.Track, 0, -1)
		case scoreProgram:
			s.setInst(e.Track, e.Vel)
		}
	}
	return s.now < s.end
}

// Writer is the register write s plays through.
func (s *Sequencer) Writer() func(int, byte) { return s.w }

// SID: s plays on the SID.
func (s *Sequencer) SID() bool { return s.sid != nil }

// Now is the PIT tick s last played.
func (s *Sequencer) Now() int { return s.now }

// Div is s's tempo divisor.
func (s *Sequencer) Div() int { return s.div }

// TrackEvents is each track's count of events before its end.
func (s *Sequencer) TrackEvents() []int {
	var ev []int
	for _, tr := range s.tr {
		k := 0
		for p := 0; p+1 < len(tr.ev) && tr.ev[p] != 0xFF; p += 2 {
			k++
		}
		ev = append(ev, k)
	}
	return ev
}
