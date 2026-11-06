package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"gobox/gpu"
)

// castRules is a fight whose turn the game publishes, with an aimed spell.
type castRules struct{ got []game.Decision }

func (r *castRules) State() game.State {
	return game.State{Mode: game.ModeCombat, Combat: &game.Combat{Size: image.Pt(6, 6),
		Units: []game.Unit{{ID: 1, At: image.Pt(1, 1), Active: true}, {ID: 2, At: image.Pt(3, 1), Foe: true}},
		Turn:  &game.Turn{Unit: 1, Options: []game.Option{{Label: "Cast", Key: 'C', Do: game.Cast{}}}},
		Aims:  []game.SpellRef{{ID: 7, Name: "Magic Missile", Range: 6}}}}
}
func (r *castRules) Apply(d game.Decision) game.Result { r.got = append(r.got, d); return game.Result{} }
func (r *castRules) Pending() *game.Ask                { return nil }
func (r *castRules) Step() game.Result                 { return game.Result{} }

// TestCastAimWithTurn: an aimed spell's square is picked on the board even
// though the game publishes a Turn; Enter casts it there.
func TestCastAimWithTurn(t *testing.T) {
	r := &castRules{}
	f := &Front{Game: classic.NewGame(r, game.Data{}), started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.Game.Answer(front.Action{Kind: front.ActCast, N: 7})
	for _, k := range []uint16{0, 0x4D00, 0x4D00, 13, 0} { // the aim's cursor two squares east, Enter
		f.Act()
		f.Draw(dst)
		f.Tick()
		if k != 0 {
			f.Key(k)
		}
	}
	f.Act()
	for _, d := range r.got {
		if c, ok := d.(game.Cast); ok && c.Spell == 7 {
			return
		}
	}
	t.Fatalf("decisions %#v, want a Cast of spell 7", r.got)
}
