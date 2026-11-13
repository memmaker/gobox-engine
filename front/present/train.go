package present

import (
	"fmt"
	"strings"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/ui"
)

// Training (play.Trading): the hall's offer, the question, the new spell.
func (p *Presenter) trainItems() (string, []ui.MenuRow) {
	s := p.state()
	t, party, sel := trainingOf(s), partyOf(s), p.selected()
	leave := ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }}
	if sel < 0 || sel >= len(party) {
		return "No one to train", []ui.MenuRow{leave}
	}
	name := party[sel].Name
	if t.Learning {
		var it []ui.MenuRow
		p.camp.ids = p.camp.ids[:0]
		for _, sp := range t.Learn {
			it = append(it, ui.MenuRow{Text: sp.Name, Key: 0, Do: func() { p.call(func(e *classicGame) { e.Learn(sp.ID); p.training = p.hallUp() }) }})
			p.camp.ids = append(p.camp.ids, sp.ID)
		}
		return name + " may add a spell to the book", it
	}
	if why := t.Why[sel]; why != "" {
		return why, []ui.MenuRow{leave}
	}
	have := play.CommonOf(s).Purse[Coin]
	cost := ui.InkGold
	if have < t.Cost {
		cost = ui.InkRed
	}
	g := func(c rune, s string) string { return string(c) + s + string(c) }
	will := name + " will become\n" + g(ui.InkGold, strings.Join(t.Becomes[sel], " and ")) + "."
	return fmt.Sprintf("%s\n\nTraining costs %s.\nThe party has %s.\n\nDo you wish to train?",
			will, g(cost, price(front.Money{Gold: t.Cost})), g(ui.InkGold, price(front.Money{Gold: have}))),
		[]ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { p.call(func(e *classicGame) { e.Train(); p.training = p.hallUp() }) }}, {Text: "No", Key: 'N', Do: leave.Do}}
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
func (p *Presenter) powerItems() (string, []ui.MenuRow) {
	pp := p.powerPick()
	if ask := pp.Name + pp.What; p.camp.level != ask { // the next pick starts on its first line
		p.camp.level, p.camp.sel, p.camp.top = ask, 0, 0
	}
	var it []ui.MenuRow
	for _, m := range pp.Options {
		it = append(it, ui.MenuRow{Text: m.Name, Key: 0, Do: func() { p.call(func(e *classicGame) { e.PickPower(m.ID) }) }})
	}
	return pp.Name + " chooses " + pp.What, it
}
