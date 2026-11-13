package engine

import (
	"path/filepath"
)

// The soundtrack (extra, the Esc menu's Audio "Soundtrack"): recordings in
// the adventure's soundtrack folder (the game's AreaMusic.Files) play by
// area and building (areaSong), in place of the Amiga's: Ogg Vorbis at
// 44100 Hz, levelled to -20 LUFS.
// A song without its file is silent. The recordings play in the presenter
// (present.Soundtrack); the rules choose the song.

const soundtrackDir = "soundtrack"

// SoundtrackFiles are the recordings' names (without .ogg) by song (AreaMusic.Files).
func (e *Engine) SoundtrackFiles() map[int]string { return e.music().Files }

// SoundtrackDir is where the recordings are.
func (e *Engine) SoundtrackDir() string { return filepath.Join(e.adv, soundtrackDir) }

// SetSoundtrack turns the soundtrack on or off, restarting the song for
// where the party is in the music now chosen.
func (e *Engine) SetSoundtrack(on bool) {
	s := &e.snd
	s.stOn = on
	if s.d == nil {
		return // SetSound takes it up
	}
	n := 0
	if a := e.music().AmigaSongs; s.music < len(a) {
		n = a[s.music] // the Amiga's song's counterpart; areaMusic follows up
	}
	s.d.SetSoundtrack(on, n, s.music)
}
