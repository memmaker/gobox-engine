package audio

import (
	"bufio"
	"encoding/binary"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// The sound pack (extra): WAV files per event in sound.cfg, mixed over the Driver.

// Play starts one of event's files and reports whether the pack had one:
// false when it has nothing for event.
func (d *SoundPack) Play(event string) bool {
	d.mu.Lock()
	var name string
	var w wavClip
	cached := false
	if len(d.events[event]) > 0 {
		files := d.events[event]
		name = files[rand.IntN(len(files))] // not the game's dice: captures with -seed stay the same
		w, cached = d.wavs[name]
	}
	d.mu.Unlock()
	if name == "" {
		return false
	}
	if !cached {
		w = readWav(filepath.Join(d.dir, name)) // outside the lock: the audio thread keeps going
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.wavs[name] = w
	if len(w.pcm) == 0 {
		return false
	}
	// the same file started within 0.1 s (Bless on the whole party) plays once
	if !slices.ContainsFunc(d.playing, func(c playingClip) bool { return &c.pcm[0] == &w.pcm[0] && c.pos < sndRate/10 }) {
		d.playing = append(d.playing, playingClip{wavClip: w})
	}
	return true
}

// wavClip is a decoded WAV: s16 samples, ch (1 or 2) per frame; no pcm when
// the file is missing or not 16-bit 44100 Hz.
type wavClip struct {
	pcm []int16
	ch  int
}

type playingClip struct {
	wavClip
	pos int // frames played
}

// SoundPack is the pack: events from sound.cfg, the WAVs decoded so far and the
// clips playing. Guarded by mu, which the driver's Read takes inside its own.
type SoundPack struct {
	mu      sync.Mutex
	dir     string
	events  map[string][]string
	wavs    map[string]wavClip
	playing []playingClip
	gain    func() float64
}

// LoadSoundPack reads dir/sound.cfg; nil without it. gain is the effects' volume, asked at each Mix.
func LoadSoundPack(dir string, gain func() float64) *SoundPack {
	f, err := os.Open(filepath.Join(dir, "sound.cfg"))
	if err != nil {
		return nil
	}
	defer f.Close()
	d := &SoundPack{dir: dir, gain: gain, events: map[string][]string{}, wavs: map[string]wavClip{}}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if k = strings.TrimSpace(k); ok && !strings.HasPrefix(k, "#") {
			d.events[k] = strings.Fields(v)
		}
	}
	return d
}

// readWav decodes a 16-bit 44100 Hz mono or stereo WAV.
func readWav(path string) wavClip {
	b, err := os.ReadFile(path)
	if err != nil || len(b) < 12 || string(b[:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return wavClip{}
	}
	le16 := binary.LittleEndian.Uint16
	ch := 0
	for i := 12; i+8 <= len(b); {
		n := int(binary.LittleEndian.Uint32(b[i+4:]))
		body := b[i+8 : min(i+8+n, len(b))]
		switch string(b[i : i+4]) {
		case "fmt ":
			if len(body) >= 16 && le16(body) == 1 && binary.LittleEndian.Uint32(body[4:]) == sndRate && le16(body[14:]) == 16 {
				ch = int(le16(body[2:]))
			}
		case "data":
			if ch != 1 && ch != 2 {
				return wavClip{}
			}
			pcm := make([]int16, len(body)/2)
			for j := range pcm {
				pcm[j] = int16(le16(body[2*j:]))
			}
			return wavClip{pcm, ch}
		}
		i += 8 + n + n&1
	}
	return wavClip{}
}

// Mix (Pack) adds the clips playing to p (s16 stereo) on the effects' volume.
func (d *SoundPack) Mix(p []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.playing) == 0 {
		return
	}
	g := d.gain()
	for i := 0; i+3 < len(p); i += 4 {
		l := int(int16(binary.LittleEndian.Uint16(p[i:])))
		r := int(int16(binary.LittleEndian.Uint16(p[i+2:])))
		for k := range d.playing {
			c := &d.playing[k]
			if j := c.pos * c.ch; j+c.ch <= len(c.pcm) {
				l += int(float64(c.pcm[j]) * g)
				r += int(float64(c.pcm[j+c.ch-1]) * g)
				c.pos++
			}
		}
		binary.LittleEndian.PutUint16(p[i:], uint16(int16(max(-32768, min(32767, l)))))
		binary.LittleEndian.PutUint16(p[i+2:], uint16(int16(max(-32768, min(32767, r)))))
	}
	d.playing = slices.DeleteFunc(d.playing, func(c playingClip) bool { return (c.pos+1)*c.ch > len(c.pcm) })
}
