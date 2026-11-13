package engine

import (
	"gobox/importer/pictures"
	"gobox/rules1e"
	"testing"

	"gobox/rules"
)

// TestForeignPicRoundTrip: a picture with no POR number saves as the pool
// pick alone (its record byte 0) and loads back as the same PicID.
func TestForeignPicRoundTrip(t *testing.T) {
	e := testEngine(t)
	c := e.state().Fresh(rules1e.Kind{Race: rules1e.RaceHuman, Class: rules1e.ClsFighter}, nil)
	c.SetName("Gamma")
	c.PortHead = 5
	const foreign rules.PicID = "pool:foreign-head"
	pictures.Set(foreign, &c.PortHead, 0, &c.PoolIDs.PortraitHeadID, func(b byte) rules.PicID { return e.portPic("HEAD", b) }, func(id rules.PicID) (byte, bool) { return e.portByte("HEAD", id) })
	if c.PortHead != 0 {
		t.Fatalf("stale head byte %d", c.PortHead)
	}
	e.SaveCharacter(e.charID(c), "")
	gotID, ok := e.LoadCharacter("Gamma", false)
	if !ok {
		t.Fatal("load")
	}
	got := e.member(gotID)
	if id := pictures.Pick(got.PoolIDs.PortraitHeadID, e.portPic("HEAD", got.PortHead)); id != foreign || got.PortHead != 0 {
		t.Fatalf("loaded %q byte %d", id, got.PortHead)
	}
}
