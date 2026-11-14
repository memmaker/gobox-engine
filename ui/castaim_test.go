package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
	"gobox/rules"
)

// castGame is a fight whose turn offers an aimed spell; lone: the aim is all
// the turn allows (BR's Target), which gobox asks for at once.
type castGame struct {
	sent []game.Action
	lone bool
}

func (r *castGame) Version() uint64 { return fresh() }

func (r *castGame) State() game.State {
	sq := []rules.BoardPos{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1}}
	al := game.Allowed{Aimed: []game.Aimed{{Action: game.Cast{Spell: 7}, Squares: sq, Targets: 1}}, Plain: []game.Action{game.EndTurn{}}}
	if r.lone {
		al = game.Allowed{Aimed: []game.Aimed{{Action: game.Shoot{}, Squares: sq, Targets: 1}}}
	}
	return game.Fighting{Common: game.Common{Allowed: al, Party: []game.Member{{Spells: []game.SpellRef{{ID: 7, Name: "Magic Missile", Range: 6}}}}},
		Field: game.Battlefield{Size: rules.Size{W: 6, H: 6}, Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, Member: -1, At: rules.BoardPos{X: 3, Y: 1}, Foe: true}}}}
}
func (r *castGame) Do(a game.Action) (game.Outcome, error) {
	r.sent = append(r.sent, a)
	return game.Outcome{}, nil
}
func (r *castGame) Save() ([]byte, error) { return nil, nil }
func (r *castGame) Load([]byte) error     { return nil }

// TestCastAimWithTurn: an aimed spell's square is picked on the board even
// though the game publishes a turn; Enter casts it there.
func TestCastAimWithTurn(t *testing.T) {
	r := &castGame{}
	aimKeys(r, true)
	for _, a := range r.sent {
		if c, ok := a.(game.Cast); ok && c.Spell == 7 {
			return
		}
	}
	t.Fatalf("actions %#v, want a Cast of spell 7", r.sent)
}

// TestRulesAimWithTurn: an aim the game leaves as the turn's only action (BR's
// Target) is answered with the square: Enter sends it the shot.
func TestRulesAimWithTurn(t *testing.T) {
	r := &castGame{lone: true}
	aimKeys(r, false)
	for _, a := range r.sent {
		if _, ok := a.(game.Shoot); ok {
			return
		}
	}
	t.Fatalf("actions %#v, want the square answered", r.sent)
}

// aimKeys moves the cursor two squares east and presses Enter, after a cast
// of spell 7 when cast.
func aimKeys(r *castGame, cast bool) {
	f := &Front{Game: &classicGame{Game: NewGame(r, game.Data{})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	if cast {
		f.Game.Answer(Action{Kind: ActCast, N: 7})
	}
	for _, k := range []uint16{0, 0x4D00, 0x4D00, 13, 0} { // the aim's cursor two squares east, Enter
		f.Act()
		f.Draw(dst)
		f.Tick()
		if k != 0 {
			f.Key(k)
		}
	}
	f.Act()
}
