package present

import "testing"

// A song loops at its end unless loop is off; a one-shot then ends.
func TestSoundtrackMix(t *testing.T) {
	s := &Soundtrack{song: 1, pcm: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	s.pos = len(s.pcm) - 4
	m := make([]byte, 8)
	s.Mix(m, true)
	if s.pos != 4 || m[4] != 1 || !s.Playing() {
		t.Fatalf("no loop: pos %d m %v", s.pos, m)
	}
	s.Mix(m, false)
	if s.pos != 8 || m[4] != 0 || s.Playing() {
		t.Fatalf("one-shot did not end: pos %d m %v", s.pos, m)
	}
}
