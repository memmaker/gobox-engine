package present

import (
	"errors"

	"gobox/game/play"
)

// A shop's, temple's or camp's menu (no map, no picture of its own) draws
// over the location as last shown, never over black: a presenter of its own
// (back) draws that State, its question and text left out, so the menu's
// state (camp, pace, the page images) stays the menu's.

// still is a play.Game held at one State: the backdrop's, which takes no action.
type still struct {
	st  play.State
	ver uint64
}

var errStill = errors.New("the backdrop takes no action")

func (s still) State() play.State                    { return s.st }
func (s still) Version() uint64                      { return s.ver }
func (s still) Do(play.Action) (play.Outcome, error) { return play.Outcome{}, errStill }
func (s still) Save() ([]byte, error)                { return nil, errStill }
func (s still) Load([]byte) error                    { return errStill }

// keepBackdrop takes the State now up as the backdrop, unless a menu screen,
// the inventory, a shop or a temple is up.
func (p *Presenter) keepBackdrop() {
	if p.eng == nil || p.backdrop || p.screenUp() || p.inventoryUp() || p.shopUp() || p.templeUp() {
		return
	}
	p.backAt = still{p.state(), p.eng.G.Version()}
}

// backP is the presenter drawing the backdrop (nil: none kept), made again
// when the backdrop's State changes.
func (p *Presenter) backP() *Presenter {
	if p.backAt.st == nil {
		return nil
	}
	return p.back.get(p.backAt.ver, func() *Presenter {
		c := *p
		c.pg, c.mapImg, c.mapSrc, c.back, c.backdrop, c.q = page{}, nil, nil, memo[uint64, *Presenter]{}, true, nil
		c.SetEngine(&classicGame{Game: NewGame(p.backAt, p.eng.D), p: &c})
		return &c
	})
}
