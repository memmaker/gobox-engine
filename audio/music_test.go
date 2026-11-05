package audio

import "testing"

// 22050 Hz doubles to 44100 Hz, and the halfway frames interpolate.
func TestResample(t *testing.T) {
	out := resample([]float32{0, 0, 0.5, 0.5}, 22050)
	if len(out) != 4*4 {
		t.Fatalf("%d bytes", len(out))
	}
	if v := int16(uint16(out[4]) | uint16(out[5])<<8); v < 8000 || v > 8400 {
		t.Fatalf("halfway %d", v)
	}
}
