// soundtrack.go: a game's soundtrack recordings, decoded in the background
// and mixed in by the game's sound driver.

package ui

import (
	"os"
	"path/filepath"
	"sync"

	"gobox/audio"
)

// Soundtrack is a game's recordings by song number (Ogg Vorbis, 44100 Hz).
// One song is kept decoded: Start decodes it in the background (silence
// until it is ready); its methods are safe from the game and the audio thread.
type Soundtrack struct {
	mu   sync.Mutex
	have map[int]string // the recordings found, by song
	song int            // the song pcm is (or is being decoded) for
	pcm  []byte         // s16 stereo, nil while decoding
	pos  int
}

// NewSoundtrack finds files (song -> name without .ogg) in dir; a song
// without its file is left out.
func NewSoundtrack(dir string, files map[int]string) *Soundtrack {
	t := &Soundtrack{have: map[int]string{}}
	for n, name := range files {
		if p := filepath.Join(dir, name+".ogg"); fileExists(p) {
			t.have[n] = p
		}
	}
	return t
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// Len is how many songs have recordings.
func (t *Soundtrack) Len() int { return len(t.have) }

// Has reports whether song n has a recording.
func (t *Soundtrack) Has(n int) bool { return t.have[n] != "" }

// Song is the song last started (0: none).
func (t *Soundtrack) Song() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.song
}

// Reset forgets the song.
func (t *Soundtrack) Reset() {
	t.mu.Lock()
	t.song, t.pcm, t.pos = 0, nil, 0
	t.mu.Unlock()
}

// Start starts song n's recording from its beginning.
func (t *Soundtrack) Start(n int) {
	t.mu.Lock()
	t.song, t.pcm, t.pos = n, nil, 0
	t.mu.Unlock()
	p := t.have[n]
	go func() {
		b, err := os.ReadFile(p)
		var pcm []byte
		if err == nil {
			pcm, _ = audio.VorbisPCM(b)
		}
		if pcm == nil {
			pcm = []byte{} // unreadable: an empty track, which ends at once
		}
		t.mu.Lock()
		if t.song == n && t.pcm == nil {
			t.pcm = pcm
		}
		t.mu.Unlock()
	}()
}

// Playing reports whether the recording is still on: decoding, or short
// of its end.
func (t *Soundtrack) Playing() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pcm == nil || t.pos < len(t.pcm)
}

// Mix renders the recording into m (s16 stereo), looping it when loop.
func (t *Soundtrack) Mix(m []byte, loop bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	clear(m)
	for i := 0; i < len(m) && t.pcm != nil; i++ {
		if t.pos == len(t.pcm) {
			if !loop || len(t.pcm) == 0 {
				return
			}
			t.pos = 0
		}
		m[i] = t.pcm[t.pos]
		t.pos++
	}
}
