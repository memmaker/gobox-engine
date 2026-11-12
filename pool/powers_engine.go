package pool

import (
	"fmt"
	"gobox/core"
	"slices"
)

// owing is the first member owing a pick, and the pick (no Options: none).
func (e *Engine) owing() (core.CharID, core.PowerOwed) {
	for _, id := range e.partyIDs() {
		if p := e.fact(id, core.OwedFact{}).(core.PowerOwed); len(p.Options) > 0 {
			return id, p
		}
	}
	return core.NoChar, core.PowerOwed{}
}

// toughen keeps the ruleset's Toughness in id's hit points.
func (e *Engine) toughen(id core.CharID) {
	if _, err := e.ruleset().Resolve(nil, core.Toughen{Who: id}, nil); err != nil {
		panic(fmt.Sprintf("toughen: %v", err))
	}
}

// takePower adds power id to member who's picks.
func (e *Engine) takePower(who core.CharID, id int) {
	if _, err := e.ruleset().Resolve(nil, core.TakePower{Who: who, Power: id}, nil); err != nil {
		panic(fmt.Sprintf("take power: %v", err))
	}
}

// offerPicks: the computer's members owing a pick take the first
// option when the pick could be offered (picksOpen).
func (e *Engine) offerPicks() {
	for _, id := range e.partyIDs() {
		e.toughen(id)
	}
	if !e.picksOpen() {
		return
	}
	for {
		id, p := e.owing()
		if id == core.NoChar || e.fact(id, core.FactControl).(core.Number) <= 0x7F {
			return
		}
		e.takePower(id, int(p.Options[0].ID))
	}
}

// picksOpen: the party is exploring or at the party menu (no training
// hall over it), where a pick owed is offered (View.Powers).
func (e *Engine) picksOpen() bool {
	hall, _ := e.hallOpen()
	return e.exploring() || e.menu.on && !hall
}

// PickPower takes the pick owed now (View.Powers): power id.
func (e *Engine) PickPower(id int) error {
	who, p := e.owing()
	if who == core.NoChar || !e.picksOpen() || !slices.ContainsFunc(p.Options, func(m core.ManeuverCard) bool { return int(m.ID) == id }) {
		return core.ErrNotNow
	}
	e.takePower(who, id)
	e.toughen(who)
	return nil
}

// powersView is the pick screen's.
func (e *Engine) powersView() *PowerPick {
	id, p := e.owing()
	if id == core.NoChar || !e.picksOpen() {
		return nil
	}
	return &PowerPick{Name: e.ruleset().Sheet(id).Name, What: p.What, Options: p.Options}
}
