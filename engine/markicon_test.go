package engine

import (
	"path/filepath"
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// IconRemove is offered in the icon editor and puts the ruleset's own icon
// back: no PNG, the default colours.
func TestIconRemoveRestoresVanilla(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	e.Advance()
	e.EditIcon(0)
	if e.iconEd == nil {
		t.Fatal("no icon editor")
	}
	who := e.iconEd.who
	p := e.iconPics(who)
	p.PNG = rules.PNG("custom")
	e.do(rules.SetIconPics{Who: who, Pics: p})
	if _, err := e.doNow(game.IconColour{Part: 0, Shade: 0, Value: 5}); err != nil {
		t.Fatal("IconColour:", err)
	}
	if _, err := e.doNow(game.IconRemove{}); err != nil {
		t.Fatal("IconRemove:", err)
	}
	got := e.iconPics(who)
	if got.PNG != nil || got.Colours[0] != (rules.IconColour{Low: 1, High: 9}) {
		t.Errorf("after IconRemove: PNG %v, colour 0 %+v, want none and {1 9}", got.PNG != nil, got.Colours[0])
	}
}

// Mark and Unmark through Do: offered while exploring, published on the
// automap, kept with the save.
func TestMarkUnmark(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	e.Advance()
	at := rules.MapPos{X: 3, Y: 4}
	if _, err := e.doNow(game.Mark{At: at, Text: "secret door"}); err != nil {
		t.Fatal("Mark:", err)
	}
	marked := func() bool {
		m := game.CommonOf(e.build()).Automap
		i := slices.Index(m.Marked, at)
		return i >= 0 && m.Notes[i] == "secret door"
	}
	if !marked() {
		t.Fatal("mark not published")
	}
	if err := e.saveTo("X"); err != nil {
		t.Fatal(err)
	}
	e.ext.labels = nil
	e.readLabels(filepath.Join(e.slotRoot(), "X", labelsFile))
	if !marked() {
		t.Fatal("mark lost over save and load")
	}
	if _, err := e.doNow(game.Unmark{At: at}); err != nil {
		t.Fatal("Unmark:", err)
	}
	if marked() {
		t.Error("mark still there after Unmark")
	}
	if _, err := e.doNow(game.Unmark{At: at}); err == nil {
		t.Error("Unmark of an empty square not refused")
	}
}
