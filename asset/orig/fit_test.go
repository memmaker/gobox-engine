package orig

import "testing"

// TestFit: a twin of another size is scaled into the frame, aspect kept,
// centred, the rest Clear.
func TestFit(t *testing.T) {
	im := NewImage(1, 1, 4) // 8x4, colour 3
	for i := range im.Data {
		im.Data[i] = 3
	}
	f := im.Fit(1, 8) // 8x8: 8x4 centred, rows 2..5
	for y := range 8 {
		want := byte(Clear)
		if y >= 2 && y < 6 {
			want = 3
		}
		if got := f.Data[y*8]; got != want {
			t.Fatalf("row %d: %d, want %d", y, got, want)
		}
	}
}
