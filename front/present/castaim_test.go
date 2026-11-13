package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// castGame is a fight whose turn offers an aimed spell; lone: the aim is all
// the turn allows (BR's Target), which gobox asks for at once.
type castGame struct {
	sent []play.Action
	lone bool
}

func (r *castGame) State() play.State {
	sq := []image.Point{{1, 1}, {2, 1}, {3, 1}, {4, 1}}
	al := play.Allowed{Aimed: []play.Aimed{{Action: play.Cast{Spell: 7}, Squares: sq, Targets: 1}}, Plain: []play.Action{play.EndTurn{}}}
	if r.lone {
		al = play.Allowed{Aimed: []play.Aimed{{Action: play.Shoot{}, Squares: sq, Targets: 1}}}
	}
	return play.Fighting{Common: play.Common{Allowed: al, Party: []game.Member{{Spells: []game.SpellRef{{ID: 7, Name: "Magic Missile", Range: 6}}}}},
		Field: play.Battlefield{Size: image.Pt(6, 6), Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, At: image.Pt(1, 1)}, {ID: 2, Member: -1, At: image.Pt(3, 1), Foe: true}}}}
}
func (r *castGame) Do(a play.Action) (play.Outcome, error) {
	r.sent = append(r.sent, a)
	return play.Outcome{}, nil
}
func (r *castGame) Save() ([]byte, error) { return nil, nil }
func (r *castGame) Load([]byte) error     { return nil }

// TestCastAimWithTurn: an aimed spell's square is picked on the board even
// though the game publishes a turn; Enter casts it there.
func TestCastAimWithTurn(t *testing.T) {
	r := &castGame{}
	aimKeys(r, true)
	for _, a := range r.sent {
		if c, ok := a.(play.Cast); ok && c.Spell == 7 {
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
		if _, ok := a.(play.Shoot); ok {
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
		f.Game.Answer(front.Action{Kind: front.ActCast, N: 7})
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
