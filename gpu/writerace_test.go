package gpu

import "testing"

// TestWritePixelsAfterSubmit: pixels written after a draw was submitted never
// reach that draw (WebGPU's queue order), however busy the GPU still is.
func TestWritePixelsAfterSubmit(t *testing.T) {
	needDevice()
	src := NewImage(1, 1)
	src.WritePixels([]byte{255, 0, 0, 255})
	busy := NewImage(4096, 4096)
	dst := NewImage(1, 1)
	for range 20 {
		for range 50 { // keep the GPU busy so the write below lands while the draw is queued
			op := &DrawImageOptions{}
			op.GeoM.Scale(4096, 4096)
			busy.DrawImage(src, op)
		}
		dst.DrawImage(src, nil)
		submit()
		src.WritePixels([]byte{0, 255, 0, 255})
		if r, g, _, _ := dst.At(0, 0).RGBA(); r == 0 || g != 0 {
			t.Fatalf("the draw submitted before the write shows the written pixels (r %d g %d)", r>>8, g>>8)
		}
		src.WritePixels([]byte{255, 0, 0, 255})
	}
}
