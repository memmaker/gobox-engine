// viewcalls.go: the sheet and the inventory are the presenter's: it opens
// them (a card's View and Items, the party menu's and camp's View, the
// turn's Items), and its decisions on them are rule calls (ShowMember,
// Ready, UseItem, Drop, Sell, Buy, TakeAll, CloseItems, CloseSheet).

package present

import (
	"cmp"
	"errors"
	"gobox/core"
	"gobox/front"
	"slices"
)

// viewPasses are the decisions that go on to the engine over the sheet
// and the inventory (every screen's).
var viewPasses = []front.ActKind{front.ActHead, front.ActPortrait, front.ActFont, front.ActSound, front.ActResize,
	front.ActMainMenu, front.ActNote, front.ActNoteRemove, front.ActIcon,
	front.ActTrade, front.ActHalve, front.ActJoin, front.ActTakeMoney, front.ActLootPool, front.ActLootShare} // the shop's Take, Pool, Share

// viewAct makes a decision on (or opening) the sheet or the inventory a
// rule call; false leaves it to the others.
func (p *Presenter) viewAct(a front.Action) bool {
	v := p.shownView
	if v.Ask != nil || v.Powers != nil {
		return false
	}
	turn := p.board != nil && p.board.Turn != nil
	n := a.N
	do := func(f func(e *classicGame) error) { p.call(func(e *classicGame) { f(e) }) }
	switch k := a.Kind; {
	case p.inventoryUp():
		switch k {
		case front.ActClose:
			do(func(e *classicGame) error { return e.CloseItems(true) })
		case front.ActSheet:
			do(func(e *classicGame) error { return e.CloseItems(false) })
		case front.ActMember:
			do(func(e *classicGame) error { return e.ShowMember(n) })
		case front.ActReady:
			do(func(e *classicGame) error { return e.Ready(a.Item, a.X) })
		case front.ActUse:
			p.useCall(a.Item, 0, -1, false)
		case front.ActDrop:
			do(func(e *classicGame) error { return e.Drop(a.Item) })
		case front.ActSell:
			do(func(e *classicGame) error { return e.Sell(a.Item) })
		case front.ActBuy:
			do(func(e *classicGame) error { return e.Buy(a.Item) })
		case front.ActTakeAll:
			do(func(e *classicGame) error { return e.TakeAll() })
		case front.ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := slices.IndexFunc(v.Party, func(m front.Member) bool { return m.Selected })
			AskForMember(p, cmp.Or(Text("Trade to whom?"), "Trade to whom?"), sel, func(t int) {
				p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, Item: a.Item, X: t, Y: 1})
			})
		case front.ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case p.sheetShown():
		switch k {
		case front.ActClose:
			do(func(e *classicGame) error { return e.CloseSheet() })
		case front.ActMember:
			do(func(e *classicGame) error { return e.ShowMember(n) })
		case front.ActInventory, front.ActOverview:
			if k == front.ActOverview {
				n = -1
			}
			do(func(e *classicGame) error { return e.OpenItems(n, false) })
		case front.ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := slices.IndexFunc(v.Party, func(m front.Member) bool { return m.Selected })
			AskForMember(p, cmp.Or(Text("Trade to whom?"), "Trade to whom?"), sel, func(t int) {
				p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, Item: a.Item, X: t, Y: 1})
			})
		case front.ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case k == front.ActMember && (p.exploring() || v.Camp != nil || p.mainMenuUp() || turn), k == front.ActView && turn: // the bar's View in a turn: the active member's sheet, as a portrait's
		if turn {
			n = -1
		}
		do(func(e *classicGame) error { return e.OpenSheet(n) })
	case k == front.ActInventory && (p.exploring() || v.Camp != nil):
		do(func(e *classicGame) error { return e.OpenItems(n, true) })
	case k == front.ActItems && turn:
		do(func(e *classicGame) error { return e.OpenItems(-1, false) })
	default:
		return false
	}
	return true
}

// combatCast is the sheet's and the inventory's Cast in a fight: they
// close, and the turn's Cast takes the spell picked.
func (p *Presenter) combatCast() {
	c := p.board
	if c == nil || c.Turn == nil || len(c.Spells) == 0 {
		return
	}
	var spells []front.Spell
	for _, sp := range c.Spells {
		spells = append(spells, front.Spell{ID: sp.ID, Name: sp.Name})
	}
	p.call(func(e *classicGame) {
		e.CloseSheet()
		e.CloseItems(true)
		AskForSpell(p, spells, func(id int) { p.turnCall(front.Action{Kind: front.ActCast, N: id}, false) })
	})
}

// useCall calls UseItem on item i; a pick or question it is refused with
// is asked, and the call made again with the answer.
func (p *Presenter) useCall(i core.ItemHandle, spell, target int, sure bool) {
	p.call(func(e *classicGame) {
		var (
			who  front.PickMember
			pick front.PickSpell
		)
		switch err := e.UseItem(i, spell, target, sure); {
		case errors.As(err, &who):
			sel := slices.IndexFunc(e.View().Party, func(m front.Member) bool { return m.Selected })
			AskForMember(p, string(who), sel, func(t int) { p.useCall(i, spell, t, sure) })
		case errors.As(err, &pick):
			AskForSpell(p, pick, func(id int) { p.useCall(i, id, target, sure) })
		default:
			AskSure(p, err, func() { p.useCall(i, spell, target, true) })
		}
	})
}
