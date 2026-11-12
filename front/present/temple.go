package present

import (
	"fmt"
	"gobox/front"
	"gobox/ui"
)

// The temple (View.Temple): the presenter asks; Cure comes agreed.
func (p *Presenter) templeItems(v front.View) (string, []ui.MenuRow) {
	c, t := &p.camp, v.Temple
	name := selectedName(v)
	cure := front.Cure{}
	if c.pick >= 0 && c.pick < len(t.Cures) {
		cure = t.Cures[c.pick]
	}
	switch c.screen {
	case templeHeal:
		var it []ui.MenuRow
		for i, cu := range t.Cures {
			cost := "" // a free service (Silver Blades' temple) shows no price, not 0g
			if cu.Price != 0 {
				cost = price(front.Money{Gold: cu.Price})
			}
			row := ui.MenuRow{Text: fmt.Sprintf("%-20s %6s", cu.Name, cost), Key: 0, Do: func() {
				c.pick = i
				if cu.Needed {
					c.to(templePay, 1)()
				} else {
					c.to(templeNot, 1)()
				}
			}}
			if !cu.Needed && cu.Only { // greyed out: no key, click or button picks it
				row.Do = nil
			}
			it = append(it, row)
		}
		return withGold(name+", how can we help you?", t.Gold), append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
	case templeNot:
		if cure.Declines { // the temple only says why
			return name + " " + cure.Not, []ui.MenuRow{{Text: "OK", Key: 'O', Do: c.back(templeHeal)}}
		}
		return name + " " + cure.Not + " Cast cure anyway?", []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: c.to(templePay, 1)}, {Text: "No", Key: 'N', Do: c.back(templeHeal)}}
	case templePay:
		return withGold(fmt.Sprintf("%s will only cost %s. Pay for cure?", cure.Name, gold(cure.Price)), t.Gold), []ui.MenuRow{
			{Text: "Yes", Key: 'Y', Do: func() { pick := c.pick; p.call(func(e front.Game) { e.Cure(-1, pick) }); c.back(templeHeal)() }},
			{Text: "No", Key: 'N', Do: c.back(templeHeal)}}
	case appraise:
		return p.appraiseItems(t.Appraise, c.back(campMain))
	}
	it := []ui.MenuRow{{Text: "Heal", Key: 'H', Do: c.to(templeHeal, 0)}, {Text: "View", Key: 'V', Do: func() { p.send(front.ActMember, -1) }}}
	it = append(it, p.moneyRows(t.Money)...)
	if appraisable(t.Appraise) {
		it = append(it, ui.MenuRow{Text: "Appraise", Key: 'A', Do: c.to(appraise, 0)})
	}
	return "Temple", append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }})
}

// appraisable: the party has gems or jewelry to appraise (or one is being
// valued): else Appraise is not offered.
func appraisable(a front.Appraise) bool { return a.Gems+a.Jewelry > 0 || a.Valued > 0 }
