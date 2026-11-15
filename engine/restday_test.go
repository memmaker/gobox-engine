package engine

import (
	"testing"

	"gobox/game"
)

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

// TestCampSaysWhatRestDoes: the camping state says whether a rest would
// heal someone (Heals) and how long the planned spells take (Study), which
// is when gobox offers Fix; none with the party whole and nothing planned.
func TestCampSaysWhatRestDoes(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	c, ok := e.build().(game.Camping)
	if !ok || c.Heals {
		t.Fatalf("a whole party in camp: %T heals %v", e.build(), c.Heals)
	}
	e.state().Party[0].HP--
	if c := e.build().(game.Camping); !c.Heals {
		t.Fatal("a wounded member: Heals is not set, Fix is not offered")
	}
}
