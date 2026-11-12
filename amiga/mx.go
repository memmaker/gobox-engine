package amiga

import (
	"encoding/binary"
	"errors"
	"math"
)

// MX plays a Music-X "MXTX" song file (SotSB's SECRET.MX: 6-byte MIDI-like
// events at 192 ticks per quarter, sampled instruments with envelopes, 4
// voices) and GLIB/DIG8 effects (sounds.glb), the way SotSB's player does
// it (0x5a20c..0x5c3a8, 0x571ee in its program), ticked every vertical
// blank. It has the Player's Song/Effect/Render/Playing/Sounding methods.
type MX struct {
	songs [][]byte // events, 6 bytes each
	inst  [64]*mxInst
	fx    [][]byte // DIG8 sample data
	fxPer []int

	// Songs maps the game's song number (Song n) to the file's song (0-based);
	// nil: Song n plays song n-1.
	Songs map[int]int

	ev          []byte  // the song playing (nil: none)
	p           int     // next event
	time, due   float64 // song time and the next event's, ticks
	tempo, tpf  float64 // BPM, ticks per frame
	ramp, rampN float64 // tempo change per frame, frames left
	ch          [16]int // instrument of each MIDI channel
	v           [4]mxVoice
	f           [2]mxVoice // the effect, on a left and a right channel
	frac        float64
	lp          [2]float64
}

type mxInst struct {
	trans           int // 1/24 semitone
	sustain         int
	data            []byte // first chunk: one-shot then loop
	oneshot, repeat int
	att, rel        [][2]int // time ms, level 0..0x8000
}

type mxVoice struct {
	in         *mxInst
	note, left int // left: ticks<<8 until note-off (<0: off)
	seg, st    int // envelope segment; st 0 free, 1 attack, 2 sustain, 3 release
	env, from  float64
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
		m.ch[i] = i
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
	m.tempo = float64(u16())
	if u16()&0x8000 != 0 { // flags: bit 0 filter, bit 1 velocity volume (ponytail: neither used by SECRET.MX's player setup, ignored)
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
			if k == 0 { // ponytail: only the first (highest) octave chunk; SECRET.MX has one each
				in.data = mx[r : r+l]
			}
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
		m.ev, m.p, m.time, m.due, m.rampN = m.songs[s], 0, 0, 0, 0
		if len(m.ev) >= 6 {
			m.due = float64(binary.BigEndian.Uint16(m.ev[2:]))
		}
		m.setTempo(m.tempo)
	}
}

// Effect plays GLIB entry n once on two channels at volume 45 (0x571ee).
func (m *MX) Effect(n int) {
	if n < 0 || n >= len(m.fx) {
		return
	}
	for i := range m.f {
		m.f[i] = mxVoice{st: 2, data: m.fx[n], end: len(m.fx[n]), per: m.fxPer[n], vol: 45}
	}
}

// Playing: a song is loaded (the music volume applies).
func (m *MX) Playing() bool { return m.ev != nil }

// Sounding: the song has not reached its end.
func (m *MX) Sounding() bool { return m.ev != nil && m.p < len(m.ev) }

func (m *MX) setTempo(bpm float64) {
	m.tempo = bpm
	m.tpf = bpm * 192 / (60 * 50) // 0x5b34c: tempo*0xc000/(VBlankHz*60), 8.8
}

// frame is one vertical blank of 0x5b3cc: note-offs, due events, envelopes, period/volume.
func (m *MX) frame() {
	if m.ev == nil {
		return
	}
	for i := range m.v {
		v := &m.v[i]
		if v.st != 0 && v.st < 3 && v.left >= 0 {
			if v.left -= int(m.tpf * 256); v.left <= 0 {
				m.release(v)
			}
		}
	}
	for m.p+6 <= len(m.ev) && m.time >= m.due {
		e := m.ev[m.p:]
		arg := int(binary.BigEndian.Uint16(e[4:]))
		switch b0 := e[0]; {
		case b0 < 0x80:
			m.noteOn(int(b0), int(e[1]), arg)
		case b0 == 0x80: // tempo b1, ramped over arg ticks
			if float64(arg) <= m.tpf {
				m.setTempo(float64(e[1]))
				m.rampN = 0
			} else {
				m.rampN = float64(arg) / m.tpf
				m.ramp = (float64(e[1]) - m.tempo) / m.rampN
			}
		case b0 == 0xc0:
			m.ch[e[1]&15] = int(e[5] & 63)
		case b0 == 0xff:
			m.p = len(m.ev)
			continue
		} // ponytail: pitch bend, controllers and meta events skipped (none in SECRET.MX)
		m.p += 6
		if m.p+6 <= len(m.ev) {
			m.due += float64(binary.BigEndian.Uint16(m.ev[m.p+2:]))
		}
	}
	if m.rampN > 0 {
		m.setTempo(m.tempo + m.ramp)
		m.rampN--
	}
	m.time += m.tpf
	for i := range m.v {
		m.envelope(&m.v[i])
	}
}

func (m *MX) noteOn(note, b1, dur int) {
	vel := (b1 >> 1) & 0x78
	in := m.inst[m.ch[b1&15]]
	if vel == 0 || in == nil { // velocity only gates: flags bit 1 (velocity volume) is clear in SECRET.MX
		return
	}
	// ponytail: lowest free voice, else one releasing, else voice 0; 0x5a252's pan preference and steal order not traced
	k := -1
	for _, want := range []int{0, 3} {
		for i := range m.v {
			if k < 0 && m.v[i].st == want {
				k = i
			}
		}
	}
	k = max(k, 0)
	v := mxVoice{in: in, note: note, left: dur << 8, st: 1, data: in.data}
	// one-shot then loop (CMD_WRITE oneshot, cycles 1; CMD_WRITE repeat, cycles 0)
	v.end, v.loop, v.end2 = in.oneshot, in.oneshot, in.oneshot+in.repeat
	if in.oneshot == 0 {
		v.end = v.end2
	}
	if len(in.att) == 0 {
		v.env, v.st = float64(in.sustain), 2
	}
	m.v[k] = v
	m.pitch(&m.v[k])
}

func (m *MX) release(v *mxVoice) {
	if len(v.in.rel) == 0 {
		*v = mxVoice{}
		return
	}
	v.st, v.seg, v.from = 3, 0, v.env
}

// envelope steps the attack/release segments (linear ramps over time_ms) and sets the volume.
func (m *MX) envelope(v *mxVoice) {
	if v.st == 0 {
		return
	}
	tab := v.in.att
	if v.st == 3 {
		tab = v.in.rel
	}
	if v.st != 2 && v.seg < len(tab) {
		t, lv := tab[v.seg][0], float64(tab[v.seg][1])
		step := math.Abs(lv-v.from) * 1000 / float64(max(t, 1)*50)
		if t == 0 || math.Abs(lv-v.env) <= step {
			v.env = lv
			v.seg++
			v.from = v.env
		} else if lv > v.env {
			v.env += step
		} else {
			v.env -= step
		}
	}
	if v.seg >= len(tab) {
		if v.st == 1 {
			v.st = 2
		} else if v.st == 3 {
			*v = mxVoice{}
			return
		}
	}
	v.env = min(max(v.env, 0), 0x8000)
	vol := 64
	if e := int(v.env) >> 8; e < 0x80 {
		vol = vol * e >> 7
	}
	v.vol = min(vol, 64)
}

// pitch is 0x5a376: semitones 8.8 to a period via log2.
func (m *MX) pitch(v *mxVoice) {
	p := v.note*256 + v.in.trans*256/24
	l := 0x9fd77 - ((p-0x2d00)*4/3)*16
	if l < 0x6f73d {
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
				v.st = 0
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
