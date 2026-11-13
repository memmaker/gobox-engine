package audio

import (
	"bytes"
	"encoding/binary"
	"slices"

	"github.com/memmaker/opl2"

	"gobox/audio/glbr"
	"gobox/audio/sid"
)

// AdLib plays the GLBR/SLBR song blocks of SOUND.COM's AdLib driver (pod-go docs/MUSIC.md)
// through opl2, and 8SVX/PCM effects. It is a Tracks: song n is block n of the music DAX.
// The game supplies the bytes and the effect -> block mapping; nothing here knows the game.
type AdLib struct {
	music, sounds [][]byte // DAX blocks, by block id
	effects       map[int][]byte
	rate          int // effect sample rate (Hz)

	chip  *opl2.Chip
	sidOn bool      // the songs on the Innovation SSI-2001's SID (SetSID), not the OPL2
	sc    *sid.Chip // ... that chip
	rs    *opl2.Resampler
	song  int            // block playing, 0 none
	named map[string]int // AdLibData.Named: name -> block
	tempo int            // AdLibData.Tempo
	seq   *glbr.Sequencer

	acc int // PIT ticks accumulated, in 1193182/(outRate*16384) units
	fx  []byte
	fxP int // fixed point 16.16 position in fx
}

// AdLibData is what the game supplies.
type AdLibData struct {
	Music   []byte            // MUSIC.DAX: song n is block n
	Sounds  []byte            // ZOUNDS.DAX; may be nil
	Effects map[int]int       // effect number -> Sounds block id
	Rate    int               // effect playback rate, Hz (9000 for PoD)
	Named   map[string][]byte // more song blocks by name (a VACE cue's .SMS file); Named finds them
	Tempo   int               // the sequencer's tempo divisor (SotSB's SOUND.DAX driver: 0x76); 0: the song header's tempo word / 84 (PoD)
}

const adlibRate = 44100

// The SSI-2001's SID clock (14.31818 MHz / 14) and its level next to the OPL2's.
const (
	sidClock = 14318180.0 / 14
	sidGain  = 6700 // the OPL2's loudness (RMS) on PoD's title song
)

// SetSID plays the songs on the Innovation SSI-2001's SID (on) or the AdLib's OPL2; a
// song playing restarts on the other chip.
func (a *AdLib) SetSID(on bool) {
	if a.sidOn == on {
		return
	}
	a.sidOn = on
	if n := a.song; a.seq != nil {
		a.Start(n)
	}
}

// SID: the songs play on the SID.
func (a *AdLib) SID() bool { return a.sidOn }

// NewAdLib parses the DAX containers.
func NewAdLib(d AdLibData) (*AdLib, error) {
	a := &AdLib{effects: map[int][]byte{}, rate: d.Rate, tempo: d.Tempo}
	var err error
	if a.music, err = glbr.GlbrBlocks(d.Music); err != nil {
		return nil, err
	}
	if d.Sounds != nil {
		if a.sounds, err = glbr.GlbrBlocks(d.Sounds); err != nil {
			return nil, err
		}
	}
	for n, id := range d.Effects {
		if id >= len(a.sounds) || a.sounds[id] == nil {
			continue
		}
		a.effects[n] = a.sounds[id] // raw u8 PCM, as the original hands it to DigPak
	}
	a.named = map[string]int{}
	for name, b := range d.Named { // an .SMS file is a song block: reuse an equal one
		n := slices.IndexFunc(a.music, func(m []byte) bool { return bytes.Equal(m, b) })
		if n < 1 {
			n = len(a.music)
			a.music = append(a.music, b)
		}
		a.named[name] = n
	}
	a.chip = opl2.New()
	a.rs = opl2.NewResampler(a.chip, adlibRate)
	return a, nil
}

// Named is the block of the song AdLibData.Named called name (0: none).
func (a *AdLib) Named(name string) int {
	if n := a.named[name]; a.Has(n) {
		return n
	}
	return 0
}

func (a *AdLib) Len() int { return len(a.music) }

// Has: block n holds a song.
func (a *AdLib) Has(n int) bool {
	return n > 0 && n < len(a.music) && glbr.GlbrNew(a.music[n], a.tempo, func(int, byte) {}, false) != nil
}
func (a *AdLib) Song() int     { return a.song }
func (a *AdLib) Playing() bool { return a.seq != nil || a.fx != nil }

// Reset stops the song and the effect.
func (a *AdLib) Reset() {
	if a.seq != nil {
		a.seq.Silence()
	}
	a.seq, a.fx, a.song = nil, nil, 0
}

// Start (re)starts song n from the beginning; 0 or an unknown n stops.
func (a *AdLib) Start(n int) {
	a.Reset()
	if !a.Has(n) {
		return
	}
	a.chip = opl2.New()
	a.rs = opl2.NewResampler(a.chip, adlibRate)
	w := func(r int, v byte) { a.chip.Write(r, v) }
	if a.sidOn {
		a.sc = sid.New(sid.MOS6581, sidClock, adlibRate)
		w = func(r int, v byte) { a.sc.Write(byte(r), v) }
	}
	if a.seq = glbr.GlbrNew(a.music[n], a.tempo, w, a.sidOn); a.seq == nil { // a song without SID instruments: the OPL2
		a.seq = glbr.GlbrNew(a.music[n], a.tempo, func(r int, v byte) { a.chip.Write(r, v) }, false)
	}
	a.song, a.acc = n, 0
}

// Effect plays effect n (PCM) over whatever plays; false if unknown. The caller decides
// whether it also stops the music (Reset).
func (a *AdLib) Effect(n int) bool {
	b, ok := a.effects[n]
	if ok {
		a.fx, a.fxP = b, 0
	}
	return ok
}

// Mix fills m (s16le stereo, 44100 Hz) with the song (restarted at its end when loop) and
// the effect.
func (a *AdLib) Mix(m []byte, loop bool) {
	for i := 0; i+3 < len(m); i += 4 {
		var v int
		if a.seq != nil {
			// 72.8227 Hz PIT tick = 1193182/16384
			for a.acc += 1193182; a.acc >= adlibRate*16384; a.acc -= adlibRate * 16384 {
				if !a.seq.Tick() {
					a.seq.Silence()
					if loop {
						a.seq = glbr.GlbrNew(a.music[a.song], a.tempo, a.seq.Writer(), a.seq.SID())
					} else {
						a.seq = nil
					}
					break
				}
			}
			switch {
			case a.seq != nil && a.seq.SID():
				v = int(a.sc.ClockSample() * sidGain)
			case a.seq != nil:
				v = a.rs.Next()
			}
		}
		if a.fx != nil {
			if p := a.fxP >> 16; p < len(a.fx) {
				v += (int(a.fx[p]) - 128) * 128
				a.fxP += a.rate << 16 / adlibRate
			} else {
				a.fx = nil
			}
		}
		v = max(-32768, min(32767, v))
		binary.LittleEndian.PutUint16(m[i:], uint16(int16(v)))
		binary.LittleEndian.PutUint16(m[i+2:], uint16(int16(v)))
	}
}
