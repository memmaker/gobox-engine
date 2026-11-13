package pool

import "testing"

// TestCampDayIsClock: in camp the published Day is the clock's, the one Rest
// Until rests by, even when the sky's colour says day at night.
func TestCampDayIsClock(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = func() {
		e.state().Area.Clock[3], e.state().Area.Clock[2], e.state().Area.Clock[1] = 23, 0, 0
		e.state().Area.Sky, e.state().Area.SkyInside = 9, 9 // a day colour
		e.CampMenu()
	}
	e.Init()
	e.Start(e.Play)
	if !e.campUp() {
		t.Fatal("camp not up")
	}
	if day := commonOf(e.build()).Day; day || e.state().IsDay() {
		t.Errorf("23:00 in camp: Day %v, isDay %v", day, e.state().IsDay())
	}
}
