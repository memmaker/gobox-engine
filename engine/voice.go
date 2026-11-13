package engine

import (
	"encoding/binary"
	"gobox/audio"

	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Voice-over (extra): story texts, journal texts and first visits are read
// aloud from data/assets/speech/<key>.ogg (Ogg Vorbis, rendered by voice/render.py,
// 44100 Hz). On unless switched off in the remastered view's Esc options; each text speaks once per save
// (VOICED.GOB beside SEEN.GOB); Enter or Esc silences it. The Amiga music is
// muted while it speaks.

// voiceOffFile in the save folder marks the voice as switched off.
const voiceOffFile = "VOICE.OFF"

type voice struct {
	mu    sync.Mutex
	queue [][]byte // s16 stereo PCM, the first one playing
}

// voiceKey names a script text's file: FNV-64a of the trimmed text, in hex
// (voice/render.py computes the same).
func voiceKey(s string) string {
	h := fnv.New64a()
	h.Write([]byte(strings.TrimSpace(s)))
	return strconv.FormatUint(h.Sum64(), 16)
}

func (e *Engine) voiceOn() bool { return !e.marked(voiceOffFile) }

// marked reports whether the marker file name exists in the save folder: a
// switch the player set (VOICE.OFF) kept across runs.
func (e *Engine) marked(name string) bool {
	_, err := os.Stat(filepath.Join(e.Cfg.SaveDir, name))
	return err == nil
}

// mark creates or removes the marker file name in the save folder. The
// switch itself lives in memory; a marker that cannot be kept shows "Cannot
// write" (its callers are key toggles with no error channel).
func (e *Engine) mark(name string, on bool) {
	p := filepath.Join(e.Cfg.SaveDir, name)
	if !on {
		e.removeSaved(p)
	} else {
		e.writeFile(p, nil)
	}
}

// ToggleVoice switches the voice when toggle and reports whether it is on.
func (e *Engine) ToggleVoice(toggle bool) bool {
	if toggle {
		e.setVoiceOn(!e.voiceOn())
	}
	return e.voiceOn()
}

func (e *Engine) setVoiceOn(on bool) {
	e.mark(voiceOffFile, !on)
	if !on {
		e.stopVoice()
	}
}

// speak queues the text's recording, unless the voice is off, there is no
// recording or this save has heard it (again: Rolf retells it). A recording
// is marked heard even with the voice off: it is lore found (lore.go).
func (e *Engine) speak(key string) {
	if e.ext.voiced[key] && !e.ext.again[key] {
		return
	}
	delete(e.ext.again, key)
	b, err := os.ReadFile(filepath.Join(e.adv, speechDir, key+".ogg"))
	if err != nil {
		return
	}
	if e.ext.voiced == nil {
		e.ext.voiced = map[string]bool{}
	}
	e.ext.voiced[key] = true
	if !e.voiceOn() {
		return
	}
	pcm, err := audio.VorbisPCM(b)
	if err != nil {
		return
	}
	e.vox.mu.Lock()
	e.vox.queue = append(e.vox.queue, pcm)
	e.vox.mu.Unlock()
}

func (e *Engine) stopVoice() {
	e.vox.mu.Lock()
	e.vox.queue = nil
	e.vox.mu.Unlock()
}

// speaking reports whether a recording is queued or playing.
func (v *voice) speaking() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.queue) > 0
}

// VoiceReader is the voice's audio source, in SoundReader's format.
func (e *Engine) VoiceReader() io.Reader { return &e.vox }

func (v *voice) Read(p []byte) (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	n := len(p) / 4 * 4
	o := 0
	for o < n && len(v.queue) > 0 {
		c := copy(p[o:n], v.queue[0])
		o += c
		if v.queue[0] = v.queue[0][c:]; len(v.queue[0]) == 0 {
			v.queue = v.queue[1:]
		}
	}
	clear(p[o:n])
	scale(p[:n], volume("voice"))
	return n, nil
}

// scale turns s16 samples down by g.
func scale(p []byte, g float64) {
	if g >= 1 {
		return
	}
	for i := 0; i+1 < len(p); i += 2 {
		v := float64(int16(binary.LittleEndian.Uint16(p[i:])))
		binary.LittleEndian.PutUint16(p[i:], uint16(int16(v*g)))
	}
}
