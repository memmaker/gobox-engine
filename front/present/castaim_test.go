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
type castRules struct {
	got      []game.Decision
	rulesAim bool // the rules run the aim (BR's Target): their Turn.Aim and square question
}

func (r *castRules) State() game.State {
	return game.State{Mode: game.ModeCombat, Combat: &game.Combat{Size: image.Pt(6, 6),
		Units: []game.Unit{{ID: 1, At: image.Pt(1, 1), Active: true}, {ID: 2, At: image.Pt(3, 1), Foe: true}},
		Turn:  &game.Turn{Unit: 1, Aim: r.rulesAim, Options: []game.Option{{Label: "Cast", Key: 'C', Do: game.Cast{}}}},
		Aims:  []game.SpellRef{{ID: 7, Name: "Magic Missile", Range: 6}}}}
}
func (r *castRules) Apply(d game.Decision) game.Result {
	r.got = append(r.got, d)
	return game.Result{}
}
func (r *castRules) Pending() *game.Ask {
	if r.rulesAim {
		return &game.Ask{Kind: game.Square}
	}
	return nil
}
func (r *castRules) Step() game.Result { return game.Result{} }

// TestCastAimWithTurn: an aimed spell's square is picked on the board even
// though the game publishes a Turn; Enter casts it there.
func TestCastAimWithTurn(t *testing.T) {
	r := &castRules{}
	aimKeys(r, true)
	for _, d := range r.got {
		if c, ok := d.(game.Cast); ok && c.Spell == 7 {
			return
		}
	}
	t.Fatalf("decisions %#v, want a Cast of spell 7", r.got)
}

// TestRulesAimWithTurn: an aim the rules run (BR's Target: Turn.Aim and
// their square question) is theirs: Enter sends them the square (game.Aim),
// not the front's own aim's Continue.
func TestRulesAimWithTurn(t *testing.T) {
	r := &castRules{rulesAim: true}
	aimKeys(r, false)
	for _, d := range r.got {
		if _, ok := d.(game.Aim); ok { // the rules' aim gets the square (BR's Target: its aimed Attack)
			return
		}
	}
	t.Fatalf("decisions %#v, want the square answered", r.got)
}

// aimKeys moves the cursor two squares east and presses Enter, after a cast
// of spell 7 when cast.
func aimKeys(r *castRules, cast bool) {
	f := &Front{Game: classic.NewGame(r, game.Data{}), started: true}
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
