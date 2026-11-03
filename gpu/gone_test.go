package gpu

import "testing"

// TestDrawDeallocated: drawing from or into a deallocated image draws nothing
// instead of submitting a released texture to the GPU.
func TestDrawDeallocated(t *testing.T) {
	a, b := NewImage(4, 4), NewImage(4, 4)
	a.Fill(white().At(0, 0))
	b.DrawImage(a, nil)
	submit()
	a.Deallocate()
	b.DrawImage(a, nil)
	a.DrawImage(b, nil)
	a.Clear()
	submit()
}
