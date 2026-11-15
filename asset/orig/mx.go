package orig

import (
	"encoding/binary"
	"errors"
	"math"
)

// MX plays a Music-X "MXTX" song file (SotSB's SECRET.MX, PoD's pod.mx:
// 6-byte MIDI-like events at 192 ticks per quarter, sampled instruments
// with envelopes, 4 voices) and GLIB/DIG8 effects (sounds.glb), the way
// SotSB's player does it (0x5a20c..0x5c3a8, 0x571ee in its program: the
// code hunk loaded at 0x10000; PoD's is the same code 0xf1e higher), ticked
// every vertical blank. It has the Player's Song/Effect/Render/Playing/
// Sounding methods.
type MX struct {
	songs [][]byte // events, 6 bytes each
	inst  [64]*mxInst
	fx    [][]byte // DIG8 sample data
	fxPer []int

	// Songs maps the game's song number (Song n) to the file's song (0-based);
	// nil: Song n plays song n-1.
	Songs map[int]int

	bpm    int  // the file's tempo
	velVol bool // header flag: a note's velocity scales its volume

	ev               []byte // the song playing (nil: none)
	p                int    // next event
	time, at         int    // song time, ticks 24.8; the last event's, ticks
	tempo, tpf       int    // BPM 12.4; ticks per frame 24.8
	rampFrom         int    // tempo ramp: from this tempo
	rampBy           uint16 // by this much
	rampLen, rampPos int    // over this many ticks 24.8 (0: none), so far
	loops            []mxLoop
	ch               [16]mxChan
	v                [4]mxVoice // Paula's channels: 0 and 3 left, 1 and 2 right
	last             [2]int     // each side's last pick: 0 (voice 0, 1) or 2 (voice 3, 2)
	f                [2]mxVoice // the effect, on a left and a right channel
	frac             float64
	lp               [2]float64
}

const (
	mxHz      = 50               // vertical blanks per second (PAL)
	mxFrameMs = 1000 << 8 / mxHz // a frame, ms 24.8 (0x5ac94)
	mxMaster  = 64               // master volume (0x5acae)
	mxFull    = 0x80             // a volume factor that scales nothing
	mxLoopMax = 4                // nested repeats

	mxFlagVel = 1 << 1  // header flags: velocity volume
	mxFlagPad = 1 << 15 // 0x100 bytes follow the header

	mxBendMid   = 0x2000 // pitch bend at rest
	mxBendRange = 24     // semitones at full bend, until RPN 0 sets it
	mxPanMid    = 0x40

	// log2 of a period, 16.16
	mxPerBase  = 0x9fd77 // note 0x2d
	mxPerLong  = 0x8fd77 // a longer one takes the next octave chunk
	mxPerShort = 0x6f73d // a shorter one is silent
	mxOctave   = 1 << 16
)

// Voice states in the player's numbering: 0x5a252 takes the lower.
const (
	mxFree = iota
	mxStop // silent, freed next frame
	mxRelease
	mxReleaseStart // note off this frame
	mxSustain
	mxAttack
	mxStart // note on this frame
)

// Controllers 0x5a762 knows (ponytail: not ported, in no song: 5, 0x25 and
// 0x41 portamento, 0x7e/0x7f mono/poly legato, 0x51 and header flag bit 0 the
// LED filter, 0xbfe001 bit 1, which Render has no model of).
const (
	mxCCData     = 6 // of RPN 0: bend range
	mxCCVolume   = 7
	mxCCPan      = 10
	mxCCPedal    = 0x40
	mxCCRPNLow   = 0x64
	mxCCRPNHigh  = 0x65
	mxCCReset    = 0x79
	mxCCNotesOff = 0x7b
)

// Meta events (0xa0) by e[4]; 0 marks a section and 1 signals the game.
const (
	mxLoopStart = 2 // e[5] repeats of what follows
	mxLoopEnd   = 3
)

type mxLoop struct{ p, n int } // first event, repeats left

// mxChan is a MIDI channel (0x415e, 0x14 bytes each).
type mxChan struct {
	inst       int // instrument slot
	vol        int // mxFull or less
	bent, bend int // pitch bend 14-bit; in semitones 8.8
	bendRange  int // semitones
	rpn        int
	right      bool // its notes prefer voices 1 and 2
	pedal      bool
	dirty      bool // its voices recompute their period this frame
}

type mxInst struct {
	trans           int // 1/24 semitone
	sustain         int
	data            [][]byte // octave chunks, each one-shot then loop, each twice as long as the one before
	oneshot, repeat int      // of the first chunk
	att, rel        [][2]int // time ms, level 0..0x8000
}

type mxVoice struct {
	in         *mxInst
	ch         *mxChan
	note, left int      // left: ticks 24.8 until note-off (<0: none)
	st         int      // mxFree..mxStart
	held       bool     // note off waits for the pedal
	vel        int      // velocity's volume factor
	oct        int      // octave chunk
	octSet     bool     // oct is chosen
	tab        [][2]int // envelope segments left
	env, step  int      // level 0..0x8000, its change per frame
	rem        int      // ms 24.8 left of tab[0]
	data       []byte
	pos        float64 // in data
	end, loop  int     // play data[:end], then loop data[loop:end2]
	end2       int
	per, vol   int
}

// NewMX parses the song file (mx; nil: no songs, effects only, as Curse's
// Amiga release) and the effects library (glb; may be nil).
func NewMX(mx, glb []byte) (*MX, error) {
	m := &MX{}
	for i := range m.ch {
		m.ch[i].inst = i
		m.reset(i)
	}
	if len(mx) == 0 {
		m.effects(glb)
		return m, nil
	}
	r := 0
	bad := errors.New("amiga: bad MXTX file")
	u16 := func() int {
		if r+2 > len(mx) {
			r = len(mx) + 1
			return 0
		}
		r += 2
		return int(binary.BigEndian.Uint16(mx[r-2:]))
	}
	if len(mx) < 10 || string(mx[:4]) != "MXTX" {
		return nil, bad
	}
	r = 4
	m.bpm = u16()
	flags := u16() // 0x5c130
	m.velVol = flags&mxFlagVel != 0
	if flags&mxFlagPad != 0 {
		r += 0x100
	}
	for n := u16(); n > 0; n-- {
		if r+4 > len(mx) {
			return nil, bad
		}
		k := int(binary.BigEndian.Uint32(mx[r:])) * 6
		r += 4
		if r+k > len(mx) {
			return nil, bad
		}
		m.songs = append(m.songs, mx[r:r+k])
		r += k
	}
	for n := u16(); n > 0; n-- {
		if r+20 > len(mx) {
			return nil, bad
		}
		h := mx[r:]
		w := func(o int) int { return int(binary.BigEndian.Uint16(h[o:])) }
		in := &mxInst{trans: int(int16(w(2))), sustain: w(4),
			oneshot: int(binary.BigEndian.Uint32(h[8:])), repeat: int(binary.BigEndian.Uint32(h[12:]))}
		slot, oct, na, nr := w(0)&63, w(6), w(16), w(18)
		r += 20
		for i := 0; i < na+nr; i++ {
			e := [2]int{u16(), u16()}
			if i < na {
				in.att = append(in.att, e)
			} else {
				in.rel = append(in.rel, e)
			}
		}
		for k := 0; k < oct; k++ {
			l := (in.oneshot + in.repeat) << k
			if r+l > len(mx) {
				return nil, bad
			}
			in.data = append(in.data, mx[r:r+l])
			r += l
		}
		m.inst[slot] = in
	}
	if r > len(mx) {
		return nil, bad
	}
	m.effects(glb)
	return m, nil
}

// effects reads the GLIB library's DIG8 samples (glb; nil: none).
func (m *MX) effects(glb []byte) {
	if len(glb) >= 0x10 && string(glb[:4]) == "GLIB" {
		n := int(binary.BigEndian.Uint16(glb[8:]))
		for i := 0; i < n && 0x10+4*i+4 <= len(glb); i++ {
			o := int(binary.BigEndian.Uint32(glb[0x10+4*i:]))
			if o+6 > len(glb) {
				break
			}
			rate, l := int(binary.BigEndian.Uint16(glb[o:])), int(binary.BigEndian.Uint32(glb[o+2:]))
			l = min(l, 0x7ff8, len(glb)-o-6)
			if rate == 0 {
				rate = 7418
			}
			m.fx = append(m.fx, glb[o+6:o+6+l])
			m.fxPer = append(m.fxPer, (3579545+rate/2)/rate)
		}
	}
}

// Len is the number of songs in the file.
func (m *MX) Len() int { return len(m.songs) }

// Effects is the number of effects in the library.
func (m *MX) Effects() int { return len(m.fx) }

// Song starts the game's song n once (Songs; 0 or unknown stops) = Sound(14+n-1).
func (m *MX) Song(n int) {
	s, ok := m.Songs[n]
	if m.Songs == nil {
		s, ok = n-1, n > 0
	}
	for i := range m.v {
		m.v[i] = mxVoice{}
	}
	m.ev = nil
	if ok && s >= 0 && s < len(m.songs) {
		m.ev, m.p, m.time, m.at, m.rampLen, m.loops = m.songs[s], 0, 0, 0, 0, nil
		for i := range m.ch { // 0x5b2ea
			m.reset(i)
		}
		m.setTempo(m.bpm << 4)
	}
}

// reset is 0x5a71c: channel k's controllers at rest; odd channels sound right.
func (m *MX) reset(k int) {
	c := &m.ch[k]
	c.bent, c.bend, c.bendRange, c.vol, c.right, c.dirty = mxBendMid, 0, mxBendRange, mxFull, k&1 != 0, true
}

// Effect plays GLIB entry n once on two channels at volume 45 (0x571ee).
func (m *MX) Effect(n int) {
	if n < 0 || n >= len(m.fx) {
		return
	}
	for i := range m.f {
		m.f[i] = mxVoice{st: mxSustain, data: m.fx[n], end: len(m.fx[n]), per: m.fxPer[n], vol: 45}
	}
}

// Playing: a song is loaded (the music volume applies).
func (m *MX) Playing() bool { return m.ev != nil }

// Sounding: the song has not reached its end.
func (m *MX) Sounding() bool { return m.ev != nil && m.p < len(m.ev) }

// setTempo is 0x5b34c: tempo*0xc000/(VBlankHz*60) of the whole BPM.
func (m *MX) setTempo(t int) {
	m.tempo = t
	m.tpf = t >> 4 * 0xc000 / (mxHz * 60)
}

// frame is one vertical blank of 0x5b3cc: note-offs, due events, tempo ramp, envelopes, period/volume.
func (m *MX) frame() {
	if m.ev == nil {
		return
	}
	m.time += m.tpf
	for i := range m.v {
		v := &m.v[i]
		if v.st == mxFree || v.left < 0 {
			continue
		}
		if m.tpf < v.left {
			v.left -= m.tpf
			continue
		}
		v.left = -1
		if v.st > mxReleaseStart { // 0x5b25e
			m.off(v)
		}
	}
	m.events()
	if m.rampLen != 0 { // 0x5b748; the change is taken unsigned, as there
		if m.rampPos += m.tpf; m.rampPos >= m.rampLen {
			m.setTempo(int(uint16(m.rampFrom) + m.rampBy))
			m.rampLen = 0
		} else {
			q := int32(uint32(m.rampBy)*uint32(m.rampPos)) / int32(m.rampLen)
			m.setTempo(int(uint16(m.rampFrom) + uint16(q)))
		}
	}
	for i := range m.v {
		m.envelope(&m.v[i])
	}
	for i := range m.ch {
		m.ch[i].dirty = false
	}
}

// events is 0x5b4c0: the events due by the song time.
func (m *MX) events() {
	for m.p+6 <= len(m.ev) {
		e := m.ev[m.p:]
		due := m.at + int(binary.BigEndian.Uint16(e[2:]))
		if m.time>>8 < due {
			return
		}
		m.at = due
		arg := binary.BigEndian.Uint16(e[4:])
		k := int(e[1] & 15)
		c := &m.ch[k]
		switch b0 := e[0]; {
		case b0 < 0x80:
			m.noteOn(c, int(b0), int(e[1]>>1&0x78), int(arg))
		case b0 == 0x80: // tempo b1, ramped over arg ticks; the compare is signed 16-bit, so 0x8000 ticks and more are at once
			if to := int(e[1]) << 4; int16(uint16(m.tpf>>8)-arg) >= 0 {
				m.setTempo(to)
				m.rampLen = 0
			} else {
				m.rampFrom, m.rampBy, m.rampLen, m.rampPos = m.tempo, uint16(to-m.tempo), int(arg)<<8, 0
			}
		case b0 == 0xa0: // 0x5b68a
			switch e[4] {
			case mxLoopStart:
				if len(m.loops) < mxLoopMax {
					m.loops = append(m.loops, mxLoop{m.p + 6, int(e[5])})
				}
			case mxLoopEnd:
				if n := len(m.loops); n > 0 {
					if l := &m.loops[n-1]; l.n > 0 {
						l.n--
						m.p, m.time, m.at = l.p, 0, 0
						return
					}
					m.loops = m.loops[:n-1]
				}
			}
		case b0 == 0xb0:
			m.control(k, e[4], e[5])
		case b0 == 0xc0:
			c.inst = int(e[5] & 63)
		case b0 == 0xe0: // 0x5a900: pitch bend, 7 bits each
			c.bent = int(e[4]&0x7f)<<7 + int(e[5]&0x7f)
			c.bend = (c.bent - mxBendMid) * int(int16(c.bendRange<<8)) / mxBendMid
			c.dirty = true
		case b0 == 0xff:
			m.p = len(m.ev)
			return
		}
		m.p += 6
	}
}

// control is 0x5a762: controller cc of channel k.
func (m *MX) control(k int, cc, val byte) {
	c := &m.ch[k]
	switch cc {
	case mxCCData:
		if c.rpn != 0 {
			return
		}
		if c.bendRange = int(val); int8(val) > mxBendRange {
			c.bendRange = mxBendRange
		}
		if c.bent != mxBendMid { // unsigned there: a downward bend comes out upward
			c.bend = int(int16(uint32(uint16(c.bent-mxBendMid)) * uint32(c.bendRange) << 8 / mxBendMid))
			c.dirty = true
		}
	case mxCCVolume: // takes effect when the voice's volume is next set (envelope)
		if c.vol = int(val); val != 0 {
			c.vol = int(val + 1)
		}
	case mxCCPan: // below the middle is right; the middle is the other side than at rest
		c.right = int8(val-mxPanMid) < 0 || val == mxPanMid && k&1 == 0
	case mxCCPedal:
		if c.pedal = val&0x40 != 0; !c.pedal {
			for i := range m.v {
				if v := &m.v[i]; v.ch == c && v.held {
					v.held, v.st = false, mxReleaseStart
				}
			}
		}
	case mxCCRPNLow:
		c.rpn = c.rpn&^0xff | int(val)
	case mxCCRPNHigh:
		c.rpn = c.rpn&0xff | int(val)<<8
	case mxCCReset:
		m.reset(k)
	case mxCCNotesOff: // 0x5b2b8
		for i := range m.v {
			if v := &m.v[i]; v.ch == c {
				m.off(v)
			}
		}
	}
}

// off ends the voice's note, once the pedal is up.
func (m *MX) off(v *mxVoice) {
	if v.ch.pedal {
		v.held = true
	} else {
		v.st = mxReleaseStart
	}
}

// noteOn is 0x5af76. Velocity only gates unless the header says it is volume.
func (m *MX) noteOn(c *mxChan, note, vel, dur int) {
	in := m.inst[c.inst]
	if vel == 0 || in == nil || len(in.data) == 0 {
		return
	}
	v := m.alloc(c.right)
	*v = mxVoice{in: in, ch: c, note: note, left: dur << 8, st: mxStart, vel: mxFull}
	if m.velVol {
		v.vel = vel + 1
	}
	m.pitch(v)
	v.data = in.data[v.oct]
	// one-shot then loop (CMD_WRITE oneshot, cycles 1; CMD_WRITE repeat, cycles 0)
	one, rep := in.oneshot<<v.oct, in.repeat<<v.oct
	v.end, v.loop, v.end2 = one, one, one+rep
	if one == 0 {
		v.end = v.end2
	}
}

// alloc is 0x5a252: the voice a note takes, whatever it plays. The other
// side if both of its own are held and the other has one free or releasing;
// of a side's two the one in the lower state, taking turns when equal.
// (Its priorities and locked voices are for the player's own sample calls,
// which the games do not use: 0x571ee has its own audio.device requests.)
func (m *MX) alloc(right bool) *mxVoice {
	side := 0
	if right {
		side = 1
	}
	low := func(s int) int { return min(m.v[s].st, m.v[s^3].st) }
	if low(side) > mxReleaseStart && low(side^1) <= mxReleaseStart {
		side ^= 1
	}
	k := side // 0x5a33a
	if a, b := m.v[k].st, m.v[k^3].st; a > b || a == b && k&2 == m.last[side] {
		k ^= 3
	}
	m.last[side] = k & 2
	return &m.v[k]
}

// envelope is 0x5a468: steps the attack/release segments (linear ramps over time_ms) and sets the volume.
func (m *MX) envelope(v *mxVoice) {
	switch v.st {
	case mxFree:
		return
	case mxStop: // 0x5a6ea
		*v = mxVoice{}
		return
	case mxSustain:
		if !v.ch.dirty {
			return
		}
	case mxStart:
		if len(v.in.att) == 0 {
			v.st, v.env = mxSustain, v.in.sustain
		} else {
			v.st = mxAttack
			v.segment(v.in.att)
		}
	case mxReleaseStart:
		if len(v.in.rel) == 0 {
			v.st, v.vol = mxStop, 0
			return
		}
		v.st = mxRelease
		v.segment(v.in.rel)
	}
	if v.st != mxSustain { // 0x5a54c
		if mxFrameMs < v.rem {
			v.env = min(max(v.env+v.step, 0), 0x8000)
			v.rem -= mxFrameMs
		} else if v.env = v.tab[0][1]; len(v.tab) > 1 {
			v.segment(v.tab[1:])
		} else if v.st == mxRelease {
			v.st, v.vol = mxStop, 0
			return
		} else {
			v.st = mxSustain
		}
	}
	m.volume(v)
	if v.ch.dirty {
		v.octSet = true
		m.pitch(v)
	}
}

// segment starts the envelope's ramp to tab[0] (0x5a436: its step per frame).
func (v *mxVoice) segment(tab [][2]int) {
	v.tab, v.rem, v.step = tab, tab[0][0]<<8, 0
	if ms := tab[0][0]; ms != 0 {
		v.step = (tab[0][1] - v.env) * 1000 / (ms * mxHz)
	}
}

// volume is 0x5a20c: the master volume scaled by velocity, envelope and channel volume.
func (m *MX) volume(v *mxVoice) {
	vol := mxMaster
	for _, f := range [...]int{v.vel, v.env >> 8, v.ch.vol} {
		if f < mxFull {
			vol = vol * f >> 7
		}
	}
	v.vol = min(vol, mxMaster)
}

// pitch is 0x5a376: semitones 8.8 to a period via log2; a new note takes the
// octave chunk that keeps the period short and stays on it.
func (m *MX) pitch(v *mxVoice) {
	p := v.ch.bend + v.note*256 + v.in.trans*256/24
	l := mxPerBase - ((p-0x2d00)*4/3)*16
	if !v.octSet {
		for v.oct = 0; l > mxPerLong && v.oct+1 < len(v.in.data); v.oct++ {
			l -= mxOctave
		}
	} else {
		l -= v.oct * mxOctave
	}
	if l < mxPerShort {
		v.per = 0
		return
	}
	v.per = int(math.Round(math.Exp2(float64(l) / 65536)))
}

// Render mixes the voices (0 and 3 left, 1 and 2 right) and the effect into s16 stereo
// at rate Hz, ticking at 50 Hz, through the A500's fixed low-pass (as Player.Render).
func (m *MX) Render(out []byte, rate int) {
	k := 1 - math.Exp(-2*math.Pi*4420/float64(rate))
	for i := 0; i+3 < len(out); i += 4 {
		if m.frac -= 50; m.frac < 0 {
			m.frac += float64(rate)
			m.frame()
		}
		var lr [2]float64
		for c := range m.v {
			lr[(c+1)>>1&1] += m.v[c].sample(rate)
		}
		for c := range m.f {
			lr[c] += m.f[c].sample(rate)
		}
		for s := range lr {
			m.lp[s] += k * (lr[s]*2 - m.lp[s])
			binary.LittleEndian.PutUint16(out[i+2*s:], uint16(int16(max(-32768, min(32767, m.lp[s])))))
		}
	}
}

// sample is the voice's held Paula output averaged over one output sample, times its volume.
func (v *mxVoice) sample(rate int) float64 {
	if v.st == 0 || v.per == 0 || v.end == 0 {
		return 0
	}
	step := pal / float64(v.per) / float64(rate)
	sum := 0.0
	for t := step; t > 0; {
		ip := int(v.pos)
		if ip >= v.end {
			if v.end2 <= v.loop || v.end2 > len(v.data) { // no loop part: silent after the one-shot
				v.end = 0
				break
			}
			v.pos -= float64(v.end - v.loop)
			v.end = v.end2
			continue
		}
		left := float64(ip+1) - v.pos
		d := min(left, t)
		sum += float64(int8(v.data[ip])) * d
		v.pos += d
		t -= d
	}
	return sum / step * float64(v.vol)
}
