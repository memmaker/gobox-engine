package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
	"gobox/rules"
)

// poolAimGame is PoR's spell aim: the cast's aim up on its own, Leave beside it.
type poolAimGame struct{ sent []game.Action }

func (r *poolAimGame) Version() uint64 { return fresh() }
func (r *poolAimGame) State() game.State {
	sq := []rules.BoardPos{{X: 3, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 4, Y: 1}}
	al := game.Allowed{Aimed: []game.Aimed{{Action: game.Cast{Member: 0, Spell: 7}, Squares: sq, Targets: 1}}, Plain: []game.Action{game.Leave{}}}
	return game.Fighting{Common: game.Common{Allowed: al, Party: []game.Member{{Spells: []game.SpellRef{{ID: 7, Name: "Stinking Cloud", Range: 6, Area: 2}}}}},
		Field: game.Battlefield{Size: rules.Size{W: 6, H: 6}, Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, Member: -1, At: rules.BoardPos{X: 3, Y: 1}, Foe: true}}}}
}
func (r *poolAimGame) Do(a game.Action) (game.Outcome, error) {
	r.sent = append(r.sent, a)
	return game.Outcome{}, nil
}
func (r *poolAimGame) Save() ([]byte, error) { return nil, nil }
func (r *poolAimGame) Load([]byte) error     { return nil }

// poolAimSend is what the game is sent after a, with the aim up.
func poolAimSend(a Action) []game.Action {
	r := &poolAimGame{}
	f := &Front{Game: &classicGame{Game: NewGame(r, game.Data{})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	for range 2 {
		f.Act()
		f.Draw(dst)
		f.Tick()
	}
	f.P.inv.acts = append(f.P.inv.acts, a)
	for range 2 {
		f.Act()
		f.Draw(dst)
		f.Tick()
	}
	return r.sent
}

// TestPoolAimClick: a square of the aim casts the spell there.
func TestPoolAimClick(t *testing.T) {
	sent := poolAimSend(Action{Kind: ActAim, X: 2, Y: 1})
	for _, a := range sent {
		if c, ok := a.(game.Cast); ok && len(c.Ats) == 1 {
			return
		}
	}
	t.Fatalf("sent %#v, want the cast at 2,1", sent)
}

// TestPoolAimEsc: Esc on the aim sends the Leave beside it.
func TestPoolAimEsc(t *testing.T) {
	sent := poolAimSend(Action{Kind: ActClose})
	for _, a := range sent {
		if a == game.Action(game.Leave{}) {
			return
		}
	}
	t.Fatalf("sent %#v, want Leave", sent)
}
