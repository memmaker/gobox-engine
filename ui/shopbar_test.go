package ui

import (
	"gobox/rules"
	"testing"

	"gobox/game"
	"gobox/gpu"
)

type shopPlay struct {
	st   game.State
	sent []game.Action
}

func (s *shopPlay) Version() uint64 { return fresh() }

func (s *shopPlay) State() game.State { return s.st }
func (s *shopPlay) Do(a game.Action) (game.Outcome, error) {
	s.sent = append(s.sent, a)
	return game.Outcome{}, nil
}
func (s *shopPlay) Save() ([]byte, error) { return nil, nil }
func (s *shopPlay) Load([]byte) error     { return nil }

// Every shop bar letter, in the shop menu and in Browse, reaches its Play.Do.
func TestShopBarReachesGame(t *testing.T) {
	for _, c := range []struct {
		keys string
		want game.Action
	}{{"P", game.PoolCoins{Member: 1}}, {"T", game.Take{Item: rules.NoItem}}, {"S", game.LootShare{}}, {"E", game.Leave{}},
		{"BP", game.PoolCoins{Member: 1}}, {"BEE", game.Leave{}}} {
		sp := &shopPlay{st: game.Trading{Common: game.Common{Party: []game.Member{{Name: "A"}, {Name: "B"}}, Selected: 1,
			Allowed: game.Allowed{Plain: []game.Action{game.PoolCoins{Member: 1}, game.Take{}, game.LootShare{}, game.Leave{}}}}}}
		f := &Front{Game: &classicGame{Game: NewGame(sp, game.Data{})}, started: true}
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
