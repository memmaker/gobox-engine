package pool

import (
	"rules1e"
	"slices"
	"testing"
)

// Camp shows the camp picture (PIC 0x1D) in the view window and its fire
// burns: frames go by on the ticks while camp is up, and stop once it is gone.
func TestCampfireAnimates(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	if !e.animating() || len(e.pic.frames) < 2 {
		t.Fatalf("camp: animating %v, %d frames", e.animating(), len(e.pic.frames))
	}
	first := e.pic.cur
	for i := 0; i < 200 && e.pic.cur == first; i++ {
		e.pic.last -= 1000 // the frame's delay gone by
		e.Step()
		e.Advance()
	}
	if e.pic.cur == first {
		t.Fatal("the campfire does not change frames")
	}
	if e.snap.win != WinPicture || e.pace() != PaceAnim {
		t.Fatalf("camp view: window %v, pace %v", e.snap.win, e.pace())
	}
	e.LeaveCamp()
	if e.animInput {
		t.Fatal("the picture still animates after camp")
	}
}

// In camp every member's context menu carries their camp actions (Icon,
// Drop, Memorize for a caster); outside camp it does not.
func TestCampMemberPowers(t *testing.T) {
	e := quickStarted(t)
	has := func(c *rules1e.Character, k PowerKind) bool {
		return slices.ContainsFunc(e.memberPowers(e.charID(c)), func(p Power) bool { return p.Kind == k })
	}
	if has(e.state().Party[0], PowerDrop) {
		t.Fatal("Drop outside camp")
	}
	e.CampMenu()
	for _, c := range e.state().Party {
		if !has(c, PowerIcon) || !has(c, PowerDrop) {
			t.Fatalf("%s: no camp actions", c.Name())
		}
		if book, _ := e.state().SpellLists(c); len(book) > 0 != has(c, PowerPlan) {
			t.Fatalf("%s: Memorize %v with %d spells", c.Name(), has(c, PowerPlan), len(book))
		}
	}
}

// Icon in a member's camp menu opens the icon editor as a page of its
// own, not under the 3D view.
func TestCampIconEditor(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	e.EditIcon(0)
	if e.iconEd == nil || e.adventure() {
		t.Fatalf("editor %v, adventure %v", e.iconEd != nil, e.adventure())
	}
}
