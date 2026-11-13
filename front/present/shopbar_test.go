package present

import (
	"gobox/core"
	"testing"

	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

type shopPlay struct {
	st   play.State
	sent []play.Action
}

func (s *shopPlay) State() play.State { return s.st }
func (s *shopPlay) Do(a play.Action) (play.Outcome, error) {
	s.sent = append(s.sent, a)
	return play.Outcome{}, nil
}
func (s *shopPlay) Save() ([]byte, error) { return nil, nil }
func (s *shopPlay) Load([]byte) error     { return nil }

// Every shop bar letter, in the shop menu and in Browse, reaches its Play.Do.
func TestShopBarReachesGame(t *testing.T) {
	for _, c := range []struct {
		keys string
		want play.Action
	}{{"P", play.PoolCoins{Member: 1}}, {"T", play.Take{Item: core.NoItem}}, {"S", play.LootShare{}}, {"E", play.Leave{}},
		{"BP", play.PoolCoins{Member: 1}}, {"BEE", play.Leave{}}} {
		sp := &shopPlay{st: play.Trading{Common: play.Common{Party: []game.Member{{Name: "A"}, {Name: "B"}}, Selected: 1,
			Allowed: play.Allowed{Plain: []play.Action{play.PoolCoins{Member: 1}, play.Take{}, play.LootShare{}, play.Leave{}}}}}}
		f := &Front{Game: NewGame(sp, game.Data{}), started: true}
		f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
		f.P.SetEngine(f.Game)
		dst := gpu.NewImage(640, 400)
		for _, k := range c.keys {
			f.Act()
			f.Draw(dst)
			f.Key(uint16(k))
		}
		f.Act()
		if len(sp.sent) != 1 || sp.sent[0] != c.want {
			t.Errorf("%s: sent %#v, want %#v", c.keys, sp.sent, c.want)
		}
	}
}
