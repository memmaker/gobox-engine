package present

import (
	"fmt"
	"gobox/front"
	"gobox/ui"
	"strings"
)

// Training (View.Training): the hall's offer, the question, the new spell.
func (p *Presenter) trainItems(v front.View) (string, []ui.MenuRow) {
	t := v.Training
	leave := ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }}
	switch {
	case len(t.Learn) > 0:
		var it []ui.MenuRow
		p.camp.ids = p.camp.ids[:0]
		for _, s := range t.Learn {
			it = append(it, ui.MenuRow{Text: s.Name, Key: 0, Do: func() { p.call(func(e front.Game) { e.Learn(s.ID); p.training = e.TrainingView() != nil }) }})
			p.camp.ids = append(p.camp.ids, s.ID)
		}
		return t.Name + " may add a spell to the book", it
	case t.Why != "":
		return t.Why, []ui.MenuRow{leave}
	}
	cost := ui.InkGold
	if t.Gold < t.Price {
		cost = ui.InkRed
	}
	g := func(c rune, s string) string { return string(c) + s + string(c) }
	will, costs := t.Will, t.Cost
	if will == "" {
		will = t.Name + " will become\n" + g(ui.InkGold, strings.Join(t.Becomes, " and ")) + "."
	}
	if costs == "" {
		costs = price(front.Money{Gold: t.Price})
	}
	return fmt.Sprintf("%s\n\nTraining costs %s.\nThe party has %s.\n\nDo you wish to train?",
			will, g(cost, costs), g(ui.InkGold, price(front.Money{Gold: t.Gold}))),
		[]ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { p.call(func(e front.Game) { e.Train(); p.training = e.TrainingView() != nil }) }}, {Text: "No", Key: 'N', Do: leave.Do}}
}

// powerItems is the power pick (View.Powers): one line a power, its rule
// the tip; no hotkeys (the lines are names) and no Exit (the pick is owed).
func (p *Presenter) powerItems(v front.View) (string, []ui.MenuRow) {
	pp := v.Powers
	if ask := pp.Name + pp.What; p.camp.level != ask { // the next pick starts on its first line
		p.camp.level, p.camp.sel, p.camp.top = ask, 0, 0
	}
	var it []ui.MenuRow
	for _, m := range pp.Options {
		it = append(it, ui.MenuRow{Text: m.Name, Key: 0, Do: func() { p.call(func(e front.Game) { e.PickPower(m.ID) }) }})
	}
	return pp.Name + " chooses " + pp.What, it
}
