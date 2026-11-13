package audio

import (
	"encoding/binary"
	"sync"

	"gobox/asset/orig"
)

// The sound cards of START.EXE's driver (por's DS:5372).
const (
	Tandy = iota
	Speaker
	None
)

// Tracks plays a soundtrack's recordings (front/present.Soundtrack; pool and the presenter hold it as Tracks).
type Tracks interface {
	Len() int
	Has(n int) bool
	Song() int // the song started last
	Reset()
	Start(n int)
	Playing() bool
	Mix(m []byte, loop bool)
}

// Amiga is an Amiga player the Driver plays instead of the synth:
// *amiga.Player (PoR's Wally Beben player) or *amiga.MX (SotSB's Music-X).
type Amiga interface {
	Song(n int)
	Effect(n int)
	Render(out []byte, rate int)
	Playing() bool
	Sounding() bool
}

// Pack is a sound pack's clips playing, mixed in on top (por's dasp); it
// keeps its own lock, taken inside the driver's.
type Pack interface{ Mix(p []byte) }

// soundtrackLevel brings the recordings (-20 LUFS) down to the voice-over's
// loudness (-29.7 LUFS).
const soundtrackLevel = 0.33

// Driver is START.EXE's sound driver (por's segment 08AE) as a game calls
// it: the Synth, or the Amiga players, the soundtrack and a sound pack, mixed
// into one Read. It implements front.Sound. Nothing sounds and nothing is
// cut until Attach, so headless tools stay silent.
type Driver struct {
	under    bool   // the synth plays under the Amiga players: a sound with no Amiga effect (Curse's title tune)
	synBuf   []byte // its output, mixed in
	mu       sync.Mutex
	on       bool // attached
	typ      int  // Tandy, Speaker or None
	wait     int  // the driver's DS:5374
	cut      int  // output frames until the 0.1 s cut (Stop+Unhook); 0: none
	syn      *Synth
	am       Amiga // nil: the DOS driver
	fx       []int // the Amiga's effect for Effect n, at n-2
	stBuf    []byte
	stSong   int  // the soundtrack song playing
	stNext   int  // the song that follows when a one-shot ends
	stOn     bool // the soundtrack option
	stPlay   bool // the option is on and there are recordings
	t        Tracks
	pack     Pack
	speaking func() bool // a voice-over speaks: the music is silent
	own      func() bool // the player's own music plays (SetOwnMusic): the game's is silent
	vol      func(kind string) float64
}

// NewDriver makes the driver on card ('T' Tandy, 'P' speaker, else None)
// from the game's tables (sounddata.json).
func NewDriver(data []byte, card byte) (*Driver, error) {
	syn, err := NewSynth(data, card)
	if err != nil {
		return nil, err
	}
	d := &Driver{syn: syn, typ: None, speaking: func() bool { return false }, vol: func(string) float64 { return 1 }}
	switch card { // por main 0x4574
	case 'T':
		d.typ = Tandy
	case 'P':
		d.typ = Speaker
	}
	return d, nil
}

// Attach turns the driver on with the Amiga players (nil: none) and their
// effect table, the soundtrack, the sound pack, the voice-over's speaking and
// the volume sliders ("music", "sfx"); any may be nil.
func (d *Driver) Attach(am Amiga, fx []int, t Tracks, p Pack, speaking func() bool, vol func(kind string) float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if p, ok := am.(*orig.Player); ok && p == nil {
		am = nil // a nil *amiga.Player (por's SoundParts without SOUND.BIN)
	}
	d.am, d.fx, d.t, d.pack = am, fx, t, p
	if speaking != nil {
		d.speaking = speaking
	}
	if vol != nil {
		d.vol = vol
	}
	d.stPlay = d.stOn && d.tracks() > 0
	d.on = true
}

// SetAmiga swaps in the Amiga player am and its effect table fx, keeping
// what else Attach gave (a game's Music-X set, game.AmigaMusic).
func (d *Driver) SetAmiga(am Amiga, fx []int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.am, d.fx = am, fx
}

// Read renders the speaker, the Tandy or the Amiga, then mixes in the
// soundtrack and the sound pack's clips: s16le stereo at 44100 Hz. A cut
// due inside p falls on its frame.
func (d *Driver) Read(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(p) / 4 * 4
	if d.cut > 0 && d.cut*4 < n {
		k := d.cut * 4
		d.read(p[:k])
		d.syn.Stop()
		d.syn.Unhook()
		d.cut = 0
		d.read(p[k:n])
		return n, nil
	}
	d.read(p[:n])
	if d.cut > 0 {
		if d.cut -= n / 4; d.cut == 0 {
			d.syn.Stop()
			d.syn.Unhook()
		}
	}
	return n, nil
}

func (d *Driver) read(p []byte) {
	n := d.render(p)
	if d.stPlay && d.typ != None {
		d.mixSoundtrack(p[:n])
	}
	if d.pack != nil && d.typ != None {
		d.pack.Mix(p[:n])
	}
}

// render renders the Amiga players, or the synth at the effects volume.
func (d *Driver) render(p []byte) int {
	n := len(p) / 4 * 4
	if d.am != nil {
		if d.typ == None {
			clear(p[:n])
		} else {
			d.am.Render(p[:n], sndRate)
			switch {
			case !d.am.Playing():
				scale(p[:n], d.vol("sfx"))
			case d.quiet(): // extra: the song plays on unheard
				clear(p[:n])
			default:
				scale(p[:n], d.vol("music"))
			}
			if d.under {
				d.mixSynth(p[:n])
			}
		}
		return n
	}
	n = d.syn.Render(p[:n])
	scale(p[:n], d.vol("sfx")) // the DOS game has no music
	return n
}

// mixSynth adds the synth's output (at the effects volume) to the Amiga
// players' in p.
func (d *Driver) mixSynth(p []byte) {
	if cap(d.synBuf) < len(p) {
		d.synBuf = make([]byte, len(p))
	}
	m := d.synBuf[:len(p)]
	clear(m)
	d.syn.Render(m)
	scale(m, d.vol("sfx"))
	for i := 0; i+1 < len(p); i += 2 {
		v := int(int16(binary.LittleEndian.Uint16(p[i:]))) + int(int16(binary.LittleEndian.Uint16(m[i:])))
		binary.LittleEndian.PutUint16(p[i:], uint16(int16(max(-32768, min(32767, v)))))
	}
}

// mixSoundtrack adds the soundtrack to p at the music volume; the song
// plays on unheard while the voice-over speaks.
func (d *Driver) mixSoundtrack(p []byte) {
	if cap(d.stBuf) < len(p) {
		d.stBuf = make([]byte, len(p))
	}
	m := d.stBuf[:len(p)]
	if d.stNext != 0 && !d.playing() {
		d.playSong(d.stNext, false)
		d.stNext = 0
	}
	if d.track(d.stSong) {
		d.t.Mix(m, d.stNext == 0) // a one-shot followed by stNext does not loop
	} else {
		clear(m)
	}
	if d.quiet() {
		return
	}
	g := d.vol("music") * soundtrackLevel
	for i := 0; i+1 < len(p); i += 2 {
		v := float64(int16(binary.LittleEndian.Uint16(p[i:]))) + g*float64(int16(binary.LittleEndian.Uint16(m[i:])))
		binary.LittleEndian.PutUint16(p[i:], uint16(int16(max(-32768, min(32767, v)))))
	}
}

// Effect = PlaySound 08AE:00DC (n): 0 stops everything and mutes, 1
// unmutes, 255 stops, anything else starts song n-1 (on the Amiga its
// effect). quiet starts nothing (a sound pack plays it, or a step under a
// song) but hooks as the sound would. While Wait is non-zero the caller
// cuts it with CutAfter.
func (d *Driver) Effect(n int, quiet bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.on || d.typ == None {
		return
	}
	if d.wait > 0 {
		d.syn.Hook()
	}
	switch n {
	case 0:
		d.syn.Stop()
		d.syn.Enable(false)
	case 1:
		d.syn.Enable(true)
	case 255:
		d.syn.Stop()
	default:
		if quiet {
		} else if d.am != nil && n >= 2 && n-2 < len(d.fx) {
			d.am.Effect(d.fx[n-2])
		} else {
			d.syn.Start(n - 1)
			d.under = d.am != nil
		}
	}
}

// CutAfter stops the synth and unhooks it after frames more output frames
// (the original's 0.1 s, rate/10), exact to the frame.
func (d *Driver) CutAfter(frames int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.on && frames > 0 {
		d.cut = frames
	}
}

// Wait = 08AE:01A5 (n): 0 hooks the timer interrupt so sounds play on; n > 0
// unhooks it (when it was hooked) and sets the cut mode.
func (d *Driver) Wait(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.on || d.typ == None {
		return
	}
	if n == 0 {
		d.syn.Hook()
	} else if d.wait == 0 {
		d.syn.Unhook()
	}
	d.wait = n
}

// Reset = 08AE:017B: silence everything.
func (d *Driver) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.on || d.typ == None {
		return
	}
	d.syn.Stop()
	d.syn.Unhook()
}

// SetCard switches the card (Tandy, Speaker, None mutes every call).
func (d *Driver) SetCard(c int) {
	d.mu.Lock()
	d.typ = c
	d.mu.Unlock()
}

// Song plays soundtrack song n (0 stops), then song next when n is a
// one-shot that ends; again restarts n when it is the song playing. False
// without the soundtrack: the caller falls back to AmigaSong.
// Do runs f under the driver's lock (a soundtrack change the mixer must not see half done).
func (d *Driver) Do(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	f()
}

func (d *Driver) Song(n, next int, again bool) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.stPlay || !d.on {
		return false
	}
	d.stNext = next
	d.playSong(n, again)
	return true
}

func (d *Driver) playSong(n int, again bool) {
	if d.track(n) && (n != d.t.Song() || again) {
		d.t.Start(n)
	}
	d.stSong = n
}

// playing reports whether the soundtrack song is still on (a one-shot ends).
func (d *Driver) playing() bool { return d.track(d.stSong) && d.t.Playing() }

func (d *Driver) tracks() int {
	if d.t == nil {
		return 0
	}
	return d.t.Len()
}

// track reports whether song n plays from a recording.
func (d *Driver) track(n int) bool { return d.stPlay && d.t.Has(n) }

// AmigaSong starts the Amiga's song n (0 stops).
func (d *Driver) AmigaSong(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.am != nil && d.on {
		d.am.Song(n)
	}
}

// SetSoundtrack turns the soundtrack on or off: on starts song n, off gives
// the channels back to the Amiga's song music.
func (d *Driver) SetSoundtrack(on bool, n, music int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stOn = on
	if !d.on || d.stPlay == (on && d.tracks() > 0) {
		return // Attach takes it up
	}
	d.stPlay = !d.stPlay
	d.stSong = 0
	d.t.Reset()
	if d.stPlay && d.am != nil {
		d.am.Song(0) // the Amiga's song gives way
	}
	if d.stPlay {
		d.playSong(n, false)
	} else if d.am != nil {
		d.am.Song(music)
	}
}

// Track reports whether the soundtrack plays, and its song.
func (d *Driver) Track() (on bool, song int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.on && d.stPlay, d.stSong
}

// Sounding reports whether the Amiga players sound (false without them).
func (d *Driver) Sounding() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.am != nil && d.am.Sounding()
}

// SetOwnMusic silences the game's songs while playing reports true (the
// player's own music plays over them; nil: never).
func (d *Driver) SetOwnMusic(playing func() bool) {
	d.mu.Lock()
	d.own = playing
	d.mu.Unlock()
}

// quiet: the game's songs play on unheard (a voice-over, the player's own music).
func (d *Driver) quiet() bool { return d.speaking() || d.own != nil && d.own() }
