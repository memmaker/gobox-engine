package present

import (
	"image"
	"testing"

	"gobox/front/classic"
	"gobox/game"
	"gobox/gpu"
)

// viewRules is a fight whose View opens the active member's sheet (front
// state, as playrules keeps it) over the turn.
type viewRules struct{ sheet bool }

func (r *viewRules) State() game.State {
	s := game.State{Mode: game.ModeCombat, Party: []game.Member{{Name: "MAREN"}}, Combat: &game.Combat{Size: image.Pt(6, 6),
		Units: []game.Unit{{ID: 1, At: image.Pt(1, 1), Active: true}, {ID: 2, At: image.Pt(3, 1), Foe: true}},
		Turn:  &game.Turn{Unit: 1, Options: []game.Option{{Label: "View", Key: 'V', Do: game.ViewMember{Member: 0}}}}}}
	if r.sheet {
		s.SheetPage, s.Sheet = &game.SheetPage{Name: "MAREN"}, &game.Sheet{}
	}
	return s
}
func (r *viewRules) Apply(d game.Decision) game.Result {
	switch d.(type) {
	case game.ViewMember:
		r.sheet = true
	case game.CloseSheet:
		r.sheet = false
	}
	return game.Result{}
}
func (r *viewRules) Pending() *game.Ask { return nil }
func (r *viewRules) Step() game.Result  { return game.Result{} }

// TestFightViewDrawn: V mid-fight puts the member's sheet on screen, and
// its keys are the sheet's; Esc goes back to the turn's board.
func TestFightViewDrawn(t *testing.T) {
	r := &viewRules{}
	f := &Front{Game: classic.NewGame(r, game.Data{}), started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.P.bars = gpu.NewImageFromImage(barArt()) // the cards' bars, as the cards load them
	frame := func() { f.Act(); f.Draw(dst); f.Tick() }
	frame()
	f.Key('v')
	frame()
	frame()
	if !r.sheet || f.P.fight || !sheetUp(f.P.shownView) {
		t.Fatalf("sheet %v, fight drawn %v", r.sheet, f.P.fight)
	}
	f.Key(27)
	frame()
	frame()
	if r.sheet || !f.P.fight {
		t.Fatalf("after Esc: sheet %v, fight drawn %v", r.sheet, f.P.fight)
	}
}
