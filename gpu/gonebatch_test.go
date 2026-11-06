package gpu

import "testing"

// TestDeallocateDropsItsBatch: draws batched into an image that is
// deallocated before it ever had a texture are dropped, not flushed into a
// pass on the released image at the next submit.
func TestDeallocateDropsItsBatch(t *testing.T) {
	src := NewImage(8, 8)
	defer src.Deallocate()
	dst := NewImage(64, 64)
	dst.DrawImage(src, nil)
	dst.Deallocate()
	other := NewImage(4, 4)
	defer other.Deallocate()
	other.WritePixels(make([]byte, 4*4*4)) // submits: no pass on the released dst
}
