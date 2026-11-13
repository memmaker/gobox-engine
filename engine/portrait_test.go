package engine

import (
	"bytes"
	"gobox/rules1e"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// A dropped portrait is saved as NAME.png beside the record and comes back
// with the character; the paper-doll head stays as it was.
func TestPortraitRoundTrip(t *testing.T) {
	e := &Engine{}
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	c := &rules1e.Character{}
	c.SetName("Yuna")
	c.PortHead, c.Portrait = 3, pngOf(t, image.NewRGBA(image.Rect(0, 0, 4, 6)))
	e.SaveCharacter(e.charID(c), "CHRDATA1")
	id, ok := e.LoadCharacter("CHRDATA1", true)
	if !ok {
		t.Fatal("not loaded")
	}
	d := e.member(id)
	if d.Portrait == nil || e.pngImage(d.Portrait).Bounds() != image.Rect(0, 0, 4, 6) || d.PortHead != 3 {
		t.Errorf("portrait %v head %d after the round trip", d.Portrait, d.PortHead)
	}
	c.Portrait = nil
	e.SaveCharacter(e.charID(c), "CHRDATA1")
	id, _ = e.LoadCharacter("CHRDATA1", true)
	d = e.member(id)
	if d.Portrait != nil {
		t.Error("the png stayed after the portrait was dropped")
	}
}

// A named save carries the party's portraits: CHRDATA<n>.png in the slot.
func TestPortraitSavedWithSlot(t *testing.T) {
	e := &Engine{set: 1, data: "../data/por/sources", adv: "../data/por/sources/assets"} // the save needs the script layout (data/assets/scripts)
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	c := &rules1e.Character{}
	c.SetName("Yuna")
	c.Portrait = pngOf(t, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	e.state().Party = append(e.state().Party, c)
	e.saveTo("X")
	if _, err := os.Stat(filepath.Join(e.slotRoot(), "X", "CHRDATA1.png")); err != nil {
		t.Error("no portrait in the slot:", err)
	}
}

// pngOf is im encoded as a PNG.
func pngOf(t *testing.T, im image.Image) []byte {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
