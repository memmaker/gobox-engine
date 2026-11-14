package engine

import (
	"gobox/game"
	"strings"
	"testing"
)

// TestShopPortrait: the dev shop shows its shopkeeper as a shop's script
// does (ShowPortrait, ovr08:0511): the head and the body, both in State.
func TestShopPortrait(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("shop")
	seed := uint32(1)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	if !playOn(e, 200, func(s game.State) bool { _, ok := s.(game.Trading); return ok }) {
		t.Fatal("no shop")
	}
	pt := commonOf(e.build()).Portrait
	if pt == nil || !strings.HasPrefix(pt.Head.File, "HEAD") || !strings.HasPrefix(pt.Body.File, "BODY") || pt.Head.Pixels == nil || pt.Body.Pixels == nil {
		t.Errorf("shopkeeper: portrait %+v", pt)
	}
}
