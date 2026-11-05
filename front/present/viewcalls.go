// viewcalls.go: the sheet and the inventory are the presenter's: it opens
// them (a card's View and Items, the party menu's and camp's View, the
// turn's Items), and its decisions on them are rule calls (ShowMember,
// Ready, UseItem, Drop, Sell, Buy, TakeAll, CloseItems, CloseSheet).

package present

import (
	"errors"
	"gobox/front"
	"slices"
)

// viewPasses are the decisions that go on to the engine over the sheet
// and the inventory (every screen's).
var viewPasses = []front.ActKind{front.ActHead, front.ActPortrait, front.ActFont, front.ActSound, front.ActResize,
	front.ActMainMenu, front.ActNote, front.ActNoteRemove, front.ActIcon, front.ActMinimap,
	front.ActTrade, front.ActHalve, front.ActJoin}

// viewAct makes a decision on (or opening) the sheet or the inventory a
// rule call; false leaves it to the others.
func (p *Presenter) viewAct(a front.Action) bool {
	v := p.shownView
	if v.Ask != nil || v.Powers != nil {
		return false
	}
	turn := v.Battle != nil && v.Battle.Turn != nil
	n := a.N
	do := func(f func(e front.Game) error) { p.call(func(e front.Game) { f(e) }) }
	switch k := a.Kind; {
	case v.Inventory != nil:
		switch k {
		case front.ActClose:
			do(func(e front.Game) error { return e.CloseItems(true) })
		case front.ActSheet:
			do(func(e front.Game) error { return e.CloseItems(false) })
		case front.ActMember:
			do(func(e front.Game) error { return e.ShowMember(n) })
		case front.ActReady:
			do(func(e front.Game) error { return e.Ready(n, a.X) })
		case front.ActUse:
			p.useCall(n, 0, -1, false)
		case front.ActDrop:
			do(func(e front.Game) error { return e.Drop(n) })
		case front.ActSell:
			do(func(e front.Game) error { return e.Sell(n) })
		case front.ActBuy:
			do(func(e front.Game) error { return e.Buy(n) })
		case front.ActTakeAll:
			do(func(e front.Game) error { return e.TakeAll() })
		case front.ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := slices.IndexFunc(v.Party, func(m front.Member) bool { return m.Selected })
			front.AskForMember(p, "Trade to whom?", sel, func(t int) {
				p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, N: n, X: t, Y: 1})
			})
		case front.ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case v.Sheet != nil:
		switch k {
		case front.ActClose:
			do(func(e front.Game) error { return e.CloseSheet() })
		case front.ActMember:
			do(func(e front.Game) error { return e.ShowMember(n) })
		case front.ActInventory, front.ActOverview:
			if k == front.ActOverview {
				n = -1
			}
			do(func(e front.Game) error { return e.OpenItems(n, false) })
		case front.ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := slices.IndexFunc(v.Party, func(m front.Member) bool { return m.Selected })
			front.AskForMember(p, "Trade to whom?", sel, func(t int) {
				p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, N: n, X: t, Y: 1})
			})
		case front.ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case k == front.ActMember && (v.Exploring || v.Camp != nil || v.MainMenu != nil || turn):
		if turn {
			n = -1
		}
		do(func(e front.Game) error { return e.OpenSheet(n) })
	case k == front.ActInventory && (v.Exploring || v.Camp != nil):
		do(func(e front.Game) error { return e.OpenItems(n, true) })
	case k == front.ActItems && turn:
		do(func(e front.Game) error { return e.OpenItems(-1, false) })
	default:
		return false
	}
	return true
}

// combatCast is the sheet's and the inventory's Cast in a fight: they
// close, and the turn's Cast takes the spell picked.
func (p *Presenter) combatCast() {
	b := p.shownView.Battle
	if b == nil || b.Turn == nil || len(b.Turn.Spells) == 0 {
		return
	}
	spells := b.Turn.Spells
	p.call(func(e front.Game) {
		e.CloseSheet()
		e.CloseItems(true)
		front.AskForSpell(p, spells, func(id int) { p.turnCall(front.Action{Kind: front.ActCast, N: id}, false) })
	})
}

// useCall calls UseItem on item i; a pick or question it is refused with
// is asked, and the call made again with the answer.
func (p *Presenter) useCall(i, spell, target int, sure bool) {
	p.call(func(e front.Game) {
		var (
			who  front.PickMember
			pick front.PickSpell
		)
		switch err := e.UseItem(i, spell, target, sure); {
		case errors.As(err, &who):
			sel := slices.IndexFunc(e.View().Party, func(m front.Member) bool { return m.Selected })
			front.AskForMember(p, string(who), sel, func(t int) { p.useCall(i, spell, t, sure) })
		case errors.As(err, &pick):
			front.AskForSpell(p, pick, func(id int) { p.useCall(i, id, target, sure) })
		default:
			front.AskSure(p, err, func() { p.useCall(i, spell, target, true) })
		}
	})
}
