// viewcalls.go: the sheet and the inventory are the presenter's: it opens
// them (a card's View and Items, the party menu's and camp's View, the
// turn's Items), and its decisions on them are rule calls (ShowMember,
// Ready, UseItem, Drop, Sell, Buy, TakeAll, CloseItems, CloseSheet).

package ui

import (
	"cmp"
	"errors"
	"gobox/game/play"
	"gobox/rules"
	"slices"
)

// viewPasses are the decisions that go on to the engine over the sheet
// and the inventory (every screen's).
var viewPasses = []ActKind{ActHead, ActPortrait, ActFont, ActSound, ActResize,
	ActMainMenu, ActNote, ActNoteRemove, ActIcon,
	ActTrade, ActHalve, ActJoin, ActTakeMoney, ActLootPool, ActLootShare} // the shop's Take, Pool, Share

// viewAct makes a decision on (or opening) the sheet or the inventory a
// rule call; false leaves it to the others.
func (p *Presenter) viewAct(a Action) bool {
	if p.askUp() != nil || p.powerPick() != nil {
		return false
	}
	turn := p.board != nil && p.board.Turn != nil
	n := a.N
	do := func(f func(e *classicGame) error) { p.call(func(e *classicGame) { f(e) }) }
	switch k := a.Kind; {
	case p.inventoryUp():
		switch k {
		case ActClose:
			do(func(e *classicGame) error { return e.CloseItems(true) })
		case ActSheet:
			do(func(e *classicGame) error { return e.CloseItems(false) })
		case ActMember:
			do(func(e *classicGame) error { return e.ShowMember(n) })
		case ActReady:
			do(func(e *classicGame) error { return e.do(play.Ready{Member: e.sel(-1), Item: a.Item, Count: a.X}) })
		case ActUse:
			p.useCall(a.Item, 0, -1, false)
		case ActDrop:
			do(func(e *classicGame) error { return e.do(play.Drop{Member: e.sel(-1), Item: a.Item}) })
		case ActSell:
			do(func(e *classicGame) error { return e.do(play.Sell{Member: e.sel(-1), Item: a.Item}) })
		case ActBuy:
			do(func(e *classicGame) error { return e.Buy(a.Item) })
		case ActTakeAll:
			do(func(e *classicGame) error { return e.do(play.Take{Item: rules.NoItem, All: true}) })
		case ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := p.selectedIndex()
			AskForMember(p, cmp.Or(Text("Trade to whom?"), "Trade to whom?"), sel, func(t int) {
				p.inv.acts = append(p.inv.acts, Action{Kind: ActTrade, Item: a.Item, X: t, Y: 1})
			})
		case ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case p.sheetShown():
		switch k {
		case ActClose:
			do(func(e *classicGame) error { return e.CloseSheet() })
		case ActMember:
			do(func(e *classicGame) error { return e.ShowMember(n) })
		case ActInventory, ActOverview:
			if k == ActOverview {
				n = -1
			}
			do(func(e *classicGame) error { return e.do(play.OpenItems{Member: n}) })
		case ActTrade: // to whom (Y 1: asked, X the member), then the game's Trade
			if a.Y == 1 {
				return false
			}
			sel := p.selectedIndex()
			AskForMember(p, cmp.Or(Text("Trade to whom?"), "Trade to whom?"), sel, func(t int) {
				p.inv.acts = append(p.inv.acts, Action{Kind: ActTrade, Item: a.Item, X: t, Y: 1})
			})
		case ActSpells:
			p.combatCast()
		default:
			return !slices.Contains(viewPasses, k)
		}
	case k == ActMember && (p.exploring() || p.campUp() || p.mainMenuUp() || turn), k == ActView && turn: // the bar's View in a turn: the active member's sheet, as a portrait's
		if turn {
			n = -1
		}
		do(func(e *classicGame) error { return e.OpenSheet(n) })
	case k == ActInventory && (p.exploring() || p.campUp()):
		do(func(e *classicGame) error { return e.do(play.OpenItems{Member: n}) })
	case k == ActItems && turn:
		do(func(e *classicGame) error { return e.do(play.OpenItems{Member: -1}) })
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
	var spells []Spell
	for _, sp := range c.Spells {
		spells = append(spells, Spell{ID: sp.ID, Name: sp.Name})
	}
	p.call(func(e *classicGame) {
		e.CloseSheet()
		e.CloseItems(true)
		AskForSpell(p, spells, func(id int) { p.turnCall(Action{Kind: ActCast, N: id}, false) })
	})
}

// useCall calls UseItem on item i; a pick or question it is refused with
// is asked, and the call made again with the answer.
func (p *Presenter) useCall(i rules.ItemHandle, spell, target int, sure bool) {
	p.call(func(e *classicGame) {
		var (
			who  PickMember
			pick PickSpell
		)
		switch err := e.UseItem(i, spell, target, sure); {
		case errors.As(err, &who):
			sel := p.selectedIndex()
			AskForMember(p, string(who), sel, func(t int) { p.useCall(i, spell, t, sure) })
		case errors.As(err, &pick):
			AskForSpell(p, pick, func(id int) { p.useCall(i, id, target, sure) })
		default:
			AskSure(p, err, func() { p.useCall(i, spell, target, true) })
		}
	})
}
