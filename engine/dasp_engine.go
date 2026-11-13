package engine

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

// dasp plays one of event's files and reports whether the pack took the
// sound: false when it is off or has nothing for event.
func (e *Engine) dasp(event string) bool {
	s := &e.snd
	d := s.dasp
	if event == "" || !e.Dasp.Load() || s.d == nil || s.typ == sndNone || d == nil {
		return false
	}
	d.mu.Lock()
	var name string
	var w daspWav
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
	if !slices.ContainsFunc(d.playing, func(c daspClip) bool { return &c.pcm[0] == &w.pcm[0] && c.pos < sndRate/10 }) {
		d.playing = append(d.playing, daspClip{daspWav: w})
	}
	return true
}

// daspDir is the pack's folder under the assets.
const daspDir = "dasp"

// SoundDir is the adventure's Amiga sound (SOUND.BIN) and the sound pack (daspDir).
const SoundDir = "sound"

// speechDir is the adventure's voice-over, one .ogg per text key.
const speechDir = "speech"

// daspWav is a decoded WAV: s16 samples, ch (1 or 2) per frame; no pcm when
// the file is missing or not 16-bit 44100 Hz.
type daspWav struct {
	pcm []int16
	ch  int
}

type daspClip struct {
	daspWav
	pos int // frames played
}

// dasp is the pack: events from sound.cfg, the WAVs decoded so far and the
// clips playing. Guarded by mu, which the driver's Read takes inside its own.
type dasp struct {
	mu      sync.Mutex
	dir     string
	events  map[string][]string
	wavs    map[string]daspWav
	playing []daspClip
}

// loadDasp reads dir/sound.cfg; nil without it.
func loadDasp(dir string) *dasp {
	f, err := os.Open(filepath.Join(dir, "sound.cfg"))
	if err != nil {
		return nil
	}
	defer f.Close()
	d := &dasp{dir: dir, events: map[string][]string{}, wavs: map[string]daspWav{}}
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
func readWav(path string) daspWav {
	b, err := os.ReadFile(path)
	if err != nil || len(b) < 12 || string(b[:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return daspWav{}
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
				return daspWav{}
			}
			pcm := make([]int16, len(body)/2)
			for j := range pcm {
				pcm[j] = int16(le16(body[2*j:]))
			}
			return daspWav{pcm, ch}
		}
		i += 8 + n + n&1
	}
	return daspWav{}
}

// Mix adds the clips playing to p (s16 stereo) on the effects' volume.
func (d *dasp) Mix(p []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.playing) == 0 {
		return
	}
	g := volume("sfx")
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
	d.playing = slices.DeleteFunc(d.playing, func(c daspClip) bool { return (c.pos+1)*c.ch > len(c.pcm) })
}
