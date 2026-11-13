package pool

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"gobox/game/play"
)

// TestInnRestCamps: Yes at the Inn opens camp and the script waits for
// it (Program 9); it must not run on into the watch's "on rest".
func TestInnRestCamps(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	seed := uint32(1)
	e.FixedSeed = &seed
	e.DevAction = func() {
		e.host.Set(e.field(tourDone), 1)
		e.posX, e.posY, e.facing = 4, 12, 4
	}
	e.Init()
	e.Start(e.Play)
	mode := fmt.Sprintf("%T", e.build())
	stepBy(e, 0)
	if e.ext.owed != nil {
		e.doNow(play.RunEvent{})
	}
	if commonOf(e.build()).Ask == nil {
		t.Fatal("no Inn question at (4,13)")
	}
	e.doNow(play.Reply{N: 0})
	watch := func() bool {
		return slices.ContainsFunc(commonOf(e.build()).Text, func(l string) bool { return strings.Contains(l, "ROUSTED") })
	}
	if q := commonOf(e.build()).Ask; q != nil || !e.campUp() {
		t.Fatalf("after Yes: asking %v camp %v watch %v", q, e.campUp(), watch())
	}
	if err := e.LeaveCamp(); err != nil {
		t.Fatal(err)
	}
	if watch() {
		t.Fatal("the watch roused after leaving an unbroken Inn rest")
	}
	if q, now := commonOf(e.build()).Ask, fmt.Sprintf("%T", e.build()); q != nil || now != mode { // 23b: the step's own continuation ran on after camp
		t.Fatalf("after camp: asking %v state %v, want %v", q, now, mode)
	}
}

// TestInnRestNo (BUG-CLASSES 28): No at the Inn camps nobody.
func TestInnRestNo(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	seed := uint32(1)
	e.FixedSeed = &seed
	e.DevAction = func() {
		e.host.Set(e.field(tourDone), 1)
		e.posX, e.posY, e.facing = 4, 12, 4
	}
	e.Init()
	e.Start(e.Play)
	stepBy(e, 0)
	if e.ext.owed != nil {
		e.doNow(play.RunEvent{})
	}
	if q := commonOf(e.build()).Ask; q == nil || len(q.Replies) != 2 || q.Replies[0] != "Yes" {
		t.Fatalf("Inn question: %+v", q)
	}
	for i := 0; i < 5 && commonOf(e.build()).Ask != nil && !e.campUp(); i++ {
		e.doNow(play.Reply{N: 1})
	}
	if e.campUp() {
		t.Fatal("No at the Inn opened camp")
	}
}
