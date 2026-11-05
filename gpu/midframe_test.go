package gpu

import (
	"image"
	"testing"
)

// TestEnsureKeepsOpenPass: uploading the pixels a texture kept from before the
// device (mid-frame, as in the window) neither submits nor ends the open pass.
func TestEnsureKeepsOpenPass(t *testing.T) {
	needDevice()
	tex := &texture{texGPU: &texGPU{}, w: 1, h: 1, px: []byte{255, 255, 255, 255}}
	dst := NewImage(8, 8)
	dst.DrawImage(white(), nil)
	flushBatch() // a pass on dst is open
	pass := dev.pass
	tex.ensure()
	if pass == nil || dev.pass != pass || dev.on != dst.t {
		t.Fatal("the upload ended the open pass")
	}
	if r, _, _, _ := (&Image{tex, image.Rect(0, 0, 1, 1)}).At(0, 0).RGBA(); r == 0 {
		t.Fatal("pixels not uploaded")
	}
}
