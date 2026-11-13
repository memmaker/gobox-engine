package present

// The point buy (scrBuy): a new character's abilities bought on
// game.PointBuyBudget points, each from game.PointBuyMin, the rows the
// abilities with their score and cost, the title the points left. Right,
// +, Enter and a click raise the ability; Left and - lower it; Raise and
// Lower change the ability last highlighted; Done hands the scores to the
// game (Game.PointBuy), which applies its rules.

import (
	"fmt"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/input"
	"slices"

	"gobox/ui"
)

// openBuy starts the point buy with every ability at the minimum.
func (p *Presenter) openBuy(cr *front.Creation) {
	s := &p.scr
	s.buy, s.buyStat = slices.Repeat([]int{game.PointBuyMin}, len(cr.StatNames)), 0
	s.sub, s.sel, s.top = scrBuy, 0, 0
}

// buyBy raises (+1) or lowers (-1) ability i where the range and the budget allow.
func (p *Presenter) buyBy(i, d int) {
	s := &p.scr
	if i < 0 || i >= len(s.buy) {
		return
	}
	old := s.buy[i]
	s.buy[i] += d
	if _, ok := game.PointBuySpent(s.buy); !ok {
		s.buy[i] = old
	}
}

// buyItems are the point buy's title and rows.
func (p *Presenter) buyItems(cr *front.Creation) (string, []ui.MenuRow) {
	s := &p.scr
	if s.sel < len(s.buy) {
		s.buyStat = s.sel
	}
	spent, _ := game.PointBuySpent(s.buy)
	w := 0
	for _, n := range cr.StatNames {
		w = max(w, len(n))
	}
	var it []ui.MenuRow
	for i, v := range s.buy {
		name := ""
		if i < len(cr.StatNames) {
			name = cr.StatNames[i]
		}
		it = append(it, ui.MenuRow{Text: fmt.Sprintf("%-*s  - %2d +  cost %2d", w, name, v, game.PointBuyCost(v)), Do: func() { p.buyBy(i, 1) }})
	}
	return fmt.Sprintf("Points left: %d", game.PointBuyBudget-spent), append(it,
		ui.MenuRow{Text: "Raise", Key: '+', Do: func() { p.buyBy(s.buyStat, 1) }},
		ui.MenuRow{Text: "Lower", Key: '-', Do: func() { p.buyBy(s.buyStat, -1) }},
		ui.MenuRow{Text: "Done", Key: 'D', Do: func() {
			scores := slices.Clone(s.buy)
			s.open(scrMain)
			p.Edit(func(e *classicGame) {
				if err := e.do(play.PointBuy{Scores: scores}); err != nil {
					e.Refuse(err.Error())
				}
			})
		}},
		ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { s.open(scrMain) }})
}

// buyKey takes Left and Right (the pad's too) on an ability's row: it
// goes down or up.
func (p *Presenter) buyKey(k uint16) bool {
	s := &p.scr
	if s.sel >= len(s.buy) {
		return false
	}
	switch actOf(k) {
	case input.Left:
		p.buyBy(s.sel, -1)
	case input.Right:
		p.buyBy(s.sel, 1)
	default:
		return false
	}
	return true
}
