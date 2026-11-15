package engine

import (
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Voice-over (extra): story texts, journal texts and first visits are read
// aloud from data/assets/speech/<key>.ogg (Ogg Vorbis, rendered by voice/render.py,
// 44100 Hz). On unless switched off in the remastered view's Esc options; each text speaks once per save
// (VOICED.GOB beside SEEN.GOB); Enter or Esc silences it. The Amiga music is
// muted while it speaks.

// voiceOffFile in the save folder marks the voice as switched off.
const voiceOffFile = "VOICE.OFF"

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
	if e.vox == nil {
		return // headless: nothing plays
	}
	traceSound("voice %q", key)
	if err := e.vox.Say(b); err != nil {
		log.Printf("voice %s: %v", key, err)
	}
}

func (e *Engine) stopVoice() {
	if e.vox != nil {
		e.vox.Stop()
	}
}
