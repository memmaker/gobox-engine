package ui

// The points screen (scrAllot): the points a new character has to spend (game.Allotment: BR's career and general
// skill points), one list of rows with their value, bonus and total, what a row still wants beside it. Enter, Right
// and + put a point on the row, Subtract and Left take one back, Next shows the next allotment. The game applies its
// rules (game.Spend) and refuses what they refuse.

import (
	"fmt"
	"strconv"
	"strings"

	"gobox/game"
	"gobox/ui/input"
)

// openAllot starts the points screen on the first allotment.
func (p *Presenter) openAllot() {
	s := &p.scr
	s.allot, s.allotRow = 0, 0
	s.sub, s.sel, s.top = scrAllot, 0, 0
}

// spendBy puts one point (by 1) on row of the allotment shown, or takes one back (-1).
func (p *Presenter) spendBy(row, by int) {
	p.sendA(Action{Kind: ActSpend, N: p.scr.allot, X: row, Y: by})
}

// wantText is what a row still wants, "needs 30 Treat Light Wounds, 25 ...".
func wantText(r game.AllotRow) string {
	var w []string
	for _, x := range r.Wants {
		w = append(w, strconv.Itoa(x.Level)+" "+x.Name)
	}
	return strings.Join(w, ", ")
}

// allotLayout is the points screen's rows by item: the row each item is, -1 for a group heading (a row of a new group is
// preceded by its heading, game.AllotRow.Group).
func allotLayout(a game.Allotment) (items []int) {
	for i, r := range a.Rows {
		if r.Group != "" && (i == 0 || a.Rows[i-1].Group != r.Group) {
			items = append(items, -1)
		}
		items = append(items, i)
	}
	return items
}

// leaveAllot leaves the points screen: from allotment from on, each one with points left asks its question (skills.pas
// EditSkills: "some career points unspent. Exit?", the general one), No stays on that allotment.
func (p *Presenter) leaveAllot(cr *FrontCreation, from int) {
	s := &p.scr
	for i := from; i < len(cr.Allotments); i++ {
		if a := cr.Allotments[i]; a.Left > 0 && a.Unspent != "" {
			s.sub, s.allot, s.sel, s.top = scrAllot, i, 0, 0
			p.askThen(a.Unspent, func() { p.leaveAllot(cr, i+1) })
			return
		}
	}
	s.open(scrMain)
}

// allotItems are the points screen's title and rows.
func (p *Presenter) allotItems(cr *FrontCreation) (string, []MenuRow) {
	s := &p.scr
	if s.allot >= len(cr.Allotments) {
		s.allot = 0
	}
	if len(cr.Allotments) == 0 {
		return "", []MenuRow{{Text: "Exit", Key: 'E', Do: func() { s.open(scrMain) }}}
	}
	a := cr.Allotments[s.allot]
	layout := allotLayout(a)
	if s.sel < len(layout) && layout[s.sel] >= 0 {
		s.allotRow = layout[s.sel]
	}
	w := 0
	for _, r := range a.Rows {
		w = max(w, len(r.Name))
	}
	var it []MenuRow
	for item, i := range layout {
		if i < 0 { // a group's heading: nothing to pick
			it = append(it, MenuRow{Text: a.Rows[layout[item+1]].Group})
			continue
		}
		r := a.Rows[i]
		text := fmt.Sprintf("%-*s %3d %3d %3d", w, r.Name, r.Value, r.Bonus, r.Total)
		if !r.Open {
			if need := wantText(r); need != "" {
				text += "  needs " + need
			}
		}
		it = append(it, MenuRow{Text: text, Key: 0, Do: func() { p.spendBy(i, 1) }})
	}
	it = append(it, MenuRow{Text: "Subtract", Key: 'S', Do: func() { p.spendBy(s.allotRow, -1) }})
	if len(cr.Allotments) > 1 {
		it = append(it, MenuRow{Text: "Next", Key: 'N', Do: func() { s.allot, s.sel, s.top = (s.allot+1)%len(cr.Allotments), 0, 0 }})
	}
	it = append(it, MenuRow{Text: "Exit", Key: 'E', Do: func() { p.leaveAllot(cr, 0) }})
	return fmt.Sprintf("%s: %d  at most %d a skill", a.Name, a.Left, a.Max), it
}

// allotKey takes Left and Right (the pad's too) on a row: a point back or onto it.
func (p *Presenter) allotKey(k uint16) bool {
	s := &p.scr
	cr := p.creation()
	if cr == nil || s.allot >= len(cr.Allotments) {
		return false
	}
	layout := allotLayout(cr.Allotments[s.allot])
	if s.sel >= len(layout) || layout[s.sel] < 0 {
		return false
	}
	switch actOf(k) {
	case input.InputLeft:
		p.spendBy(layout[s.sel], -1)
	case input.InputRight:
		p.spendBy(layout[s.sel], 1)
	default:
		return false
	}
	return true
}
