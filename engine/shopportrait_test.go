package engine

import (
	"gobox/game/play"
	"strings"
	"testing"
)

// TestShopPortrait: the dev shop shows its shopkeeper as a shop's script
// does (ShowPortrait, ovr08:0511): the head at (3,3) and the body five
// rows below it, both on the page.
func TestShopPortrait(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("shop")
	seed := uint32(1)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	if !playOn(e, 200, func(s play.State) bool { _, ok := s.(play.Trading); return ok }) {
		t.Fatal("no shop")
	}
	head, body := false, false
	for _, p := range e.pagePictures() {
		head = head || strings.HasPrefix(p.File, "HEAD") && p.At.Min.Y == 24
		body = body || strings.HasPrefix(p.File, "BODY") && p.At.Min.Y == 64
	}
	if !head || !body {
		t.Errorf("shopkeeper: head %v, body %v", head, body)
		for _, p := range e.pagePictures() {
			t.Logf("%s %d at %v", p.File, p.Block, p.At)
		}
	}
}
