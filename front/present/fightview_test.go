package present

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// viewGame is a fight in which the series' View key opens the active
// member's sheet, front state.
type viewGame struct{}

func (viewGame) Version() uint64 { return fresh() }

func (viewGame) State() play.State {
	return play.Fighting{Common: play.Common{Party: []game.Member{{Name: "MAREN"}}, Allowed: play.Allowed{Plain: []play.Action{play.EndTurn{}}}}, Field: play.Battlefield{Size: image.Pt(6, 6), Turn: 1,
		Units: []game.Unit{{ID: 1, Member: 0, At: image.Pt(1, 1)}, {ID: 2, Member: -1, At: image.Pt(3, 1), Foe: true}}}}
}
func (viewGame) Do(play.Action) (play.Outcome, error) { return play.Outcome{}, nil }
func (viewGame) Save() ([]byte, error)                { return nil, nil }
func (viewGame) Load([]byte) error                    { return nil }

// TestFightViewDrawn: V mid-fight puts the member's sheet on screen, and
// its keys are the sheet's; Esc goes back to the turn's board.
func TestFightViewDrawn(t *testing.T) {
	f := &Front{Game: &classicGame{Game: NewGame(viewGame{}, game.Data{VerbKeys: map[string]rune{"View": 'V'}})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.P.bars = gpu.NewImageFromImage(barArt()) // the cards' bars, as the cards load them
	frame := func() { f.Act(); f.Draw(dst); f.Tick() }
	frame()
	f.Key('v')
	frame()
	frame()
	if f.P.fight || !f.P.sheetUp() {
		t.Fatalf("sheet %v, fight drawn %v", f.P.sheetUp(), f.P.fight)
	}
	f.Key(27)
	frame()
	frame()
	if f.P.sheetUp() || !f.P.fight {
		t.Fatalf("after Esc: sheet %v, fight drawn %v", f.P.sheetUp(), f.P.fight)
	}
}
