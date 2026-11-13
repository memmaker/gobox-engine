package audio

import (
	"encoding/binary"
	"sync"
)

// Voice plays the voice-over's recordings one after another (game.Voice);
// it is an Output source.
type Voice struct {
	mu    sync.Mutex
	queue [][]byte // s16 stereo PCM, the first one playing
	gain  func() float64
}

// NewVoice makes the voice; gain is its volume, asked at each Read.
func NewVoice(gain func() float64) *Voice { return &Voice{gain: gain} }

// Say decodes the Ogg Vorbis recording b and queues it.
func (v *Voice) Say(b []byte) error {
	pcm, err := VorbisPCM(b)
	if err != nil {
		return err
	}
	v.mu.Lock()
	v.queue = append(v.queue, pcm)
	v.mu.Unlock()
	return nil
}

// Stop silences the voice.
func (v *Voice) Stop() {
	v.mu.Lock()
	v.queue = nil
	v.mu.Unlock()
}

// Speaking reports whether a recording is queued or playing.
func (v *Voice) Speaking() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.queue) > 0
}

func (v *Voice) Read(p []byte) (int, error) {
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
	scaleS16(p[:n], v.gain())
	return n, nil
}

// scaleS16 turns s16 samples down by g.
func scaleS16(p []byte, g float64) {
	if g >= 1 {
		return
	}
	for i := 0; i+1 < len(p); i += 2 {
		s := float64(int16(binary.LittleEndian.Uint16(p[i:])))
		binary.LittleEndian.PutUint16(p[i:], uint16(int16(s*g)))
	}
}
