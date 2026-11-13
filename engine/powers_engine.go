package engine

import (
	"fmt"
	"gobox/rules"
	"slices"
)

// owing is the first member owing a pick, and the pick (no Options: none).
func (e *Engine) owing() (rules.CharID, rules.PowerOwed) {
	for _, id := range e.partyIDs() {
		if p := e.fact(id, rules.OwedFact{}).(rules.PowerOwed); len(p.Options) > 0 {
			return id, p
		}
	}
	return rules.NoChar, rules.PowerOwed{}
}

// toughen keeps the ruleset's Toughness in id's hit points.
func (e *Engine) toughen(id rules.CharID) {
	if _, err := e.ruleset().Resolve(nil, rules.Toughen{Who: id}, nil); err != nil {
		panic(fmt.Sprintf("toughen: %v", err))
	}
}

// takePower adds power id to member who's picks.
func (e *Engine) takePower(who rules.CharID, id int) {
	if _, err := e.ruleset().Resolve(nil, rules.TakePower{Who: who, Power: id}, nil); err != nil {
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
		if id == rules.NoChar || e.fact(id, rules.FactControl).(rules.Number) <= 0x7F {
			return
		}
		e.takePower(id, int(p.Options[0].ID))
	}
}

// picksOpen: the party is exploring or at the party menu (no training
// hall over it), where a pick owed is offered (Powers).
func (e *Engine) picksOpen() bool {
	hall, _ := e.hallOpen()
	return e.exploring() || e.menu.on && !hall
}

// pickPower takes the pick owed now (Powers): power id.
func (e *Engine) pickPower(id int) error {
	who, p := e.owing()
	if who == rules.NoChar || !e.picksOpen() || !slices.ContainsFunc(p.Options, func(m rules.ManeuverCard) bool { return int(m.ID) == id }) {
		return rules.ErrNotNow
	}
	e.takePower(who, id)
	e.toughen(who)
	return nil
}

// powersView is the pick screen's.
func (e *Engine) powersView() *PowerPick {
	id, p := e.owing()
	if id == rules.NoChar || !e.picksOpen() {
		return nil
	}
	return &PowerPick{Name: e.ruleset().Sheet(id).Name, What: p.What, Options: p.Options}
}
