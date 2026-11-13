//go:build devviews

package engine

import "testing"

// TestInsideNotSky (BUG-CLASSES 63): Where.Inside is the map cell's (> 0x7F,
// as ovr27 picks SkyInside), never the sky colour or the hour: a night sky
// (8) or none (0) outdoors leaves it false.
func TestInsideNotSky(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	want := e.MapCell(e.posX, e.posY) > 0x7F
	for _, sky := range []uint16{0, 8, 11} {
		e.state().Area.Sky, e.state().Area.SkyInside = sky, sky
		if got := commonOf(e.build()).Where.Inside; got != want {
			t.Errorf("sky %d: Inside %v, the cell says %v", sky, got, want)
		}
	}
}
