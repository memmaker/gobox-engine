package ui

import (
	"fmt"
	"strings"

	"gobox/game"
)

// Training (game.Trading): the hall's offer, the question, the new spell.
func (p *Presenter) trainItems() (string, []MenuRow) {
	s := p.state()
	t, party, sel := trainingOf(s), partyOf(s), p.selected()
	leave := MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(ActClose, 0) }}
	if sel < 0 || sel >= len(party) {
		return "No one to train", []MenuRow{leave}
	}
	name := party[sel].Name
	if t.Learning {
		var it []MenuRow
		p.camp.ids = p.camp.ids[:0]
		for _, sp := range t.Learn {
			it = append(it, MenuRow{Text: sp.Name, Key: 0, Do: func() { p.call(func(e *classicGame) { e.do(game.Learn{Spell: sp.ID}); p.training = p.hallUp() }) }})
			p.camp.ids = append(p.camp.ids, sp.ID)
		}
		return name + " may add a spell to the book", it
	}
	if why := t.Why[sel]; why != "" {
		return why, []MenuRow{leave}
	}
	g := func(c rune, s string) string { return string(c) + s + string(c) }
	will := name + " will become\n" + g(InkGold, strings.Join(t.Becomes[sel], " and ")) + "."
	yes := []MenuRow{{Text: "Yes", Key: 'Y', Do: func() { p.call(func(e *classicGame) { e.do(game.Train{Member: e.sel(-1)}); p.training = p.hallUp() }) }}, {Text: "No", Key: 'N', Do: leave.Do}}
	if t.NoFee {
		return will + "\n\nDo you wish to train?", yes
	}
	have := game.CommonOf(s).Purse[Coin]
	cost := InkGold
	if have < t.Cost {
		cost = InkRed
	}
	return fmt.Sprintf("%s\n\nTraining costs %s.\nThe party has %s.\n\nDo you wish to train?",
			will, g(cost, price(Money{Gold: t.Cost})), g(InkGold, price(Money{Gold: have}))),
		yes
}

// learning is the spells the selected member may learn now (nil: the hall
// offers training).
func (p *Presenter) learning() []game.SpellRef {
	s := p.state()
	if t, n := trainingOf(s), p.selected(); t != nil && t.Learning && n >= 0 && n < len(partyOf(s)) {
		return t.Learn
	}
	return nil
}

// powerItems is the power pick (Powers): one line a power, its rule
// the tip; no hotkeys (the lines are names) and no Exit (the pick is owed).
func (p *Presenter) powerItems() (string, []MenuRow) {
	pp := p.powerPick()
	if ask := pp.Name + pp.What; p.camp.level != ask { // the next pick starts on its first line
		p.camp.level, p.camp.sel, p.camp.top = ask, 0, 0
	}
	var it []MenuRow
	for _, m := range pp.Options {
		it = append(it, MenuRow{Text: m.Name, Key: 0, Do: func() { p.call(func(e *classicGame) { e.do(game.PickPower{ID: m.ID}) }) }})
	}
	return pp.Name + " chooses " + pp.What, it
}
