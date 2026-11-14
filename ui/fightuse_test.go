package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
	"gobox/rules"
)

// useGame is a fight whose turn offers one usable item (a potion).
type useGame struct{ sent []game.Action }

func (r *useGame) Version() uint64 { return fresh() }
func (r *useGame) State() game.State {
	al := game.Allowed{Plain: []game.Action{game.EndTurn{}, game.UseItem{Member: 0, Item: 3}, game.UseItem{Member: 0, Item: rules.NoItem}}}
	return game.Fighting{Common: game.Common{Allowed: al, Party: []game.Member{{Name: "Kade", Items: []game.Item{{Handle: 3, Name: "Potion"}}}}},
		Field: game.Battlefield{Size: rules.Size{W: 6, H: 6}, Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, Member: -1, At: rules.BoardPos{X: 4, Y: 1}, Foe: true}}}}
}
func (r *useGame) Do(a game.Action) (game.Outcome, error) {
	r.sent = append(r.sent, a)
	return game.Outcome{}, nil
}
func (r *useGame) Save() ([]byte, error) { return nil, nil }
func (r *useGame) Load([]byte) error     { return nil }

// TestFightItemsAsks: the turn's Items puts up the usable items, drawn
// over the board (it was up but invisible: the fight drew only menus).
func TestFightItemsAsks(t *testing.T) {
	r := &useGame{}
	f := &Front{Game: &classicGame{Game: NewGame(r, game.Data{})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	step := func() {
		f.Act()
		f.Draw(dst)
		f.Tick()
	}
	step()
	step()
	f.P.inv.acts = append(f.P.inv.acts, Action{Kind: ActItems, N: -1})
	step()
	step()
	if q := f.P.askUp(); q == nil || q.Kind != AskList || len(q.Entries) != 1 {
		t.Fatalf("after Items: ask %+v, sent %#v", q, r.sent)
	}
	if len(f.P.menuRects) == 0 {
		t.Fatal("the items list is up but not drawn")
	}
}
