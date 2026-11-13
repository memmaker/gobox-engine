package pool

import (
	"bytes"
	"gobox/char"
	grules "gobox/rules"
	"os"
	"rules1e"
	"testing"
)

func TestPoolIDs(t *testing.T) {
	e := testEngine(t)
	e.Cfg.SaveDir = t.TempDir()
	load := func(name string) (*rules1e.Character, []byte) {
		id, ok := e.LoadCharacter(name, true)
		if !ok {
			t.Fatal("load", name)
		}
		c := e.member(id)
		b, _ := os.ReadFile(e.savePath(grules.MakeFileName(name) + ".sav"))
		return c, b
	}

	old := &rules1e.Character{}
	old.SetName("Olda")
	e.SaveCharacter(e.charID(old), "olda")
	c, b := load("olda")
	if (c.PoolIDs != rules1e.PoolIDs{}) || !bytes.Equal(b, char.SheetBytes(old.Sheet)) {
		t.Fatalf("no IDs: got %v, %d bytes (record %d)", c.PoolIDs, len(b), charRecSize)
	}

	nw := &rules1e.Character{}
	nw.SetName("Newa")
	nw.SetPoolID("portraithead", "pod/head3")
	nw.SetPoolID("portraitbody", "pod/body3")
	nw.SetPoolID("iconhead", "pod/chead2")
	nw.SetPoolID("iconbody", "pod/cbody4")
	e.SaveCharacter(e.charID(nw), "newa")
	c, b = load("newa")
	if c.PortraitHeadID != "pod/head3" || c.PortraitBodyID != "pod/body3" || c.IconHeadID != "pod/chead2" || c.IconBodyID != "pod/cbody4" {
		t.Fatalf("round trip: got %v", c.PoolIDs)
	}
	if !bytes.Equal(b[:charRecSize], char.SheetBytes(nw.Sheet)) || !bytes.HasPrefix(b[charRecSize:], []byte(rules1e.PoolMagic)) {
		t.Fatal("the original record bytes changed or the trailer is missing")
	}
}

// The pool picks a member stores come back in State.
func TestPoolIDsState(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.Advance()
	e.state().Party[0].PoolIDs = rules1e.PoolIDs{PortraitHeadID: "head_001", PortraitBodyID: "body_002", IconHeadID: "combat_head_003", IconBodyID: "combat_body_004"}
	m := commonOf(e.build()).Party[0]
	if m.PortraitHeadID != "head_001" || m.PortraitBodyID != "body_002" || m.IconHeadID != "combat_head_003" || m.IconBodyID != "combat_body_004" {
		t.Errorf("play member: %q %q %q %q", m.PortraitHeadID, m.PortraitBodyID, m.IconHeadID, m.IconBodyID)
	}
}
