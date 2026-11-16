package combat

import (
	"math"
	"testing"
)

// TestFrame: a focus in the frame box leaves the camera still; one outside it
// glides the camera, eased, the shortest way that brings it onto the box's
// edge, never centring it.
func TestFrame(t *testing.T) {
	const aspect = 16.0 / 10
	p := &Presenter{}
	p.initCamera()
	p.at = vec{10, 0, 10}
	p.frame(1.0/60, vec{10.5, 0, 10}, aspect)
	if p.at != (vec{10, 0, 10}) || p.glide.dur != 0 {
		t.Fatalf("in view, moved to %v (glide %+v)", p.at, p.glide)
	}
	focus := vec{40, 0, 10}
	var steps []float64
	for i := 0; i < 120; i++ {
		before := p.at
		p.frame(1.0/60, focus, aspect)
		steps = append(steps, math.Hypot(p.at[0]-before[0], p.at[2]-before[2]))
	}
	if p.glide.dur != 0 {
		t.Fatalf("glide still running after 2 s: %+v", p.glide)
	}
	if steps[0] >= steps[len(steps)/8] { // eases in
		t.Errorf("no ease-in: first step %v, later %v", steps[0], steps[len(steps)/8])
	}
	nx, ny, _ := p.camAt(p.at, aspect).ndc(focus)
	if math.Abs(math.Abs(nx)-frameSide) > 1e-6 && math.Abs(ny-frameTop) > 1e-6 && math.Abs(ny-frameBottom) > 1e-6 {
		t.Errorf("focus at %v,%v: not on the frame's edge", nx, ny)
	}
	if math.Abs(nx) > frameSide+1e-6 || ny > frameTop+1e-6 || ny < frameBottom-1e-6 {
		t.Errorf("focus at %v,%v: out of frame", nx, ny)
	}
}
