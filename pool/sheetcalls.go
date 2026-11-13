package pool

import (
	"errors"
	"gobox/core"
	"slices"
)

// The sheet (Sheet) and the inventory (Inventory) are the
// presenter's screens: it opens and closes them (OpenSheet, OpenItems,
// CloseItems, CloseSheet) and its decisions on them are rule calls
// (ShowMember, Ready, UseItem, Drop, Sell, Buy, TakeAll; the powers:
// UsePower). The game opens them itself for the loot, the shop's Browse
// and the temple's View, and goes on when they close (view.chain).

// openSheet = ovr20:0AB3 (var quit), the View command: the sheet of the
// selected member up.
func (e *Engine) openSheet() {
	if e.view.on {
		return
	}
	if e.inFight() {
		e.combatViewEnter() // ovr31 (seg 03E0) +2F
		e.snap.sheetOn = true
	}
	e.view.on, e.view.leave = true, false
	e.pushRedraw(rdSheet)
	e.CharacterSheet()
}

// openItems = ovr20:0EF8 (ItemsMenu): the selected member's inventory up
// (over the sheet when it is up). extra: the remastered inventory screen
// draws it, with no Halve, Join, Trade or Id (unreadied items are the
// party's), Throw in combat; Drop and Sell come already confirmed.
func (e *Engine) openItems() {
	e.snap.invOn = true
	e.applyItems(e.selected()) // port: ItemCount may be stale (shop Inv for a buyer never viewed)
}

// openThen opens the inventory (items) or the sheet for the game, which
// goes on with then once they are closed.
func (e *Engine) openThen(items bool, then cont) {
	e.push(then)
	e.view.chain = true
	if items {
		e.openItems()
	} else {
		e.openSheet()
	}
}

// viewable: the sheet or the inventory may open now (a fight: on the
// member whose turn it is).
func (e *Engine) viewable() bool {
	if e.partySize() == 0 || e.snap.invOn || e.view.on && e.inFight() {
		return false
	}
	return !e.inFight() || e.waiting(PendFight) && e.cs.actor != core.NoChar && e.cs.aim == nil
}

// OpenSheet puts member n's (< 0: the selected one; in a fight the one
// whose turn it is) sheet up.
func (e *Engine) OpenSheet(n int) error {
	if !e.viewable() || e.view.on {
		return core.ErrNotNow
	}
	if !e.inFight() {
		e.actor(n)
	}
	e.openSheet()
	return nil
}

// OpenItems puts member n's (as OpenSheet) inventory up, over the sheet
// when it is up; with direct (F1-F8, a card's Items) the sheet opens
// under it and goes with it. Without items: "No items" (direct only).
func (e *Engine) OpenItems(n int, direct bool) error {
	if !e.viewable() {
		return core.ErrNotNow
	}
	if !e.inFight() {
		e.actor(n)
	}
	if direct && !e.view.on {
		if e.carryCount(e.selected()) == 0 {
			e.say("No items")
			return core.ErrNotNow
		}
		e.openSheet()
		e.view.leave = true
	}
	e.openItems()
	return nil
}

// CloseItems = the inventory's Exit: back to the sheet under it (all:
// that goes too, and so does one opened under it, direct).
func (e *Engine) CloseItems(all bool) error {
	if !e.snap.invOn {
		return core.ErrNotNow
	}
	e.snap.invOn = false
	switch {
	case e.view.on && (all || e.view.leave):
		e.CloseSheet()
	case e.view.on:
		e.CharacterSheet()
	default:
		e.viewClosed()
	}
	return nil
}

// CloseSheet = the sheet's Exit (the inventory over it goes too).
func (e *Engine) CloseSheet() error {
	if !e.view.on {
		return core.ErrNotNow
	}
	e.closeViews()
	e.viewClosed()
	return nil
}

// closeViews closes the sheet and the inventory without what follows
// them (viewClosed): UseItem in a fight goes on itself.
func (e *Engine) closeViews() {
	if !e.view.on {
		e.snap.invOn = false
		return
	}
	e.snap.invOn, e.view.on = false, false
	e.popRedraw()
	e.snap.sheetOn, e.snap.sheet = false, nil // the page goes with it (a fight's redraw clears no screen)
	if e.inFight() {
		e.combatViewLeave() // ovr31 (seg 03E0) +2A
	}
	e.snap.portrait = CellRect{} // extra: Shift-click and dropped PNGs go with the sheet (DrawMainScreen keeps the page)
	e.DrawMainScreen()
}

// viewClosed: the sheet and the inventory are both closed; the game goes
// on (view.chain), or what is under them draws itself again.
func (e *Engine) viewClosed() {
	if e.view.chain {
		e.view.chain = false
		e.ret()
		return
	}
	e.uncovered()
}

// ShowMember shows member n on the sheet and the inventory (not in a
// fight: the sheet is the fighter's).
func (e *Engine) ShowMember(n int) error {
	if !e.view.on && !e.snap.invOn || n < 0 || n >= e.partySize() {
		return core.ErrNotNow
	}
	if e.inFight() && (e.cs.actor == core.NoChar || e.partyIDs()[n] != e.cs.actor) { // a fight: only the member whose turn it is (BUG-CLASSES 40)
		return core.ErrNotNow
	}
	e.selectMember(n)
	if e.snap.invOn {
		e.applyItems(e.selected())
	} else {
		e.CharacterSheet()
	}
	return nil
}

// carriedBy is h when it is one of the selected member's items with the
// inventory up (NoItem: not).
func (e *Engine) carriedBy(h core.ItemHandle) (core.CharID, core.ItemHandle) {
	id := e.selected()
	if !e.snap.invOn || !slices.Contains(e.listedItems(id), h) {
		return id, core.NoItem
	}
	return id, h
}

// Ready readies or unreadies item h (x of the stack, 0: all of it).
func (e *Engine) Ready(h core.ItemHandle, x int) error {
	cID, h := e.carriedBy(h)
	if h == core.NoItem {
		return core.ErrNotNow
	}
	if _, err := e.ruleset().Resolve(nil, core.Ready{Who: cID, Item: h, Count: int32(x)}, nil); err != nil {
		return err
	}
	if t := e.snap.turn; t != nil && cID == e.cs.actor { // the turn's weapon follows (BUG-CLASSES 40), as turnChoices sets it
		t.Weapon = nil
		if w := e.weapon(cID); w != core.NoItem {
			t.Weapon = e.invItem(cID, w)
			t.Weapon.Rules = []string(e.fact(cID, core.WeaponRulesFact{}).(core.Lines))
		}
	}
	return nil
}

// Unready unreadies item h, refused when it is not readied.
func (e *Engine) Unready(h core.ItemHandle) error {
	if id, h := e.carriedBy(h); h != core.NoItem && !e.itemCard(id, h).Readied {
		return errors.New("not readied")
	}
	return e.Ready(h, 0)
}

// Drop drops item h (the front asked ovr20's "Drop It?").
func (e *Engine) Drop(h core.ItemHandle) error {
	cID, h := e.carriedBy(h)
	if h == core.NoItem {
		return core.ErrNotNow
	}
	if e.vault != nil { // Deposit: the item stays in the vault (pod's dropItem)
		_, err := e.ruleset().Resolve(nil, core.DepositItem{Who: cID, Item: h}, nil)
		return err
	}
	_, err := e.ruleset().Resolve(nil, core.DropItem{Who: cID, Item: h}, nil)
	return err
}

// Sell sells item h where the inventory sells (the shop's; confirmed).
func (e *Engine) Sell(h core.ItemHandle) error {
	cID, h := e.carriedBy(h)
	if h == core.NoItem || !e.inventoryView(cID).Sell {
		return core.ErrNotNow
	}
	o, err := e.ruleset().Resolve(nil, core.Sell{Who: cID, Item: h}, nil)
	e.traded(o, "Sold!")
	return err
}

// Buy buys ware h of the shop's (Inventory.Wares).
func (e *Engine) Buy(h core.ItemHandle) error {
	if !e.snap.invOn || !e.snap.shop || !slices.Contains(e.goods(), h) {
		return core.ErrNotNow
	}
	card := e.itemCard(core.NoChar, h)
	o, err := e.ruleset().Resolve(nil, core.Buy{Who: e.selected(), Item: h, Price: e.shopPrice(card.Price, card.Count)}, nil)
	e.traded(o, "Bought!")
	return err
}

// traded says done when the trade went through (core.Traded), or the
// refusal of a purse that could not pay.
func (e *Engine) traded(o core.Outcome, done string) {
	for _, f := range o.Effects {
		switch f := f.(type) {
		case core.Traded:
			e.say(done)
		case core.Refused:
			if f.Why == core.NoMoney {
				e.refuse("Not enough Money.")
			}
		}
	}
}

// UseItem = ovr20:1A8A (item): item h casts the scroll's spell or its own
// (+0x3D, charges +0x3C). spell, target, sure are the picks it asked
// for (a scroll's spell; the target and the yes/no as UsePower). In a fight the inventory and the sheet close, the spell aims
// (a phase of the turn) and a use ends the turn.
func (e *Engine) UseItem(h core.ItemHandle, spell, target int, sure bool) error {
	cID, h := e.carriedBy(h)
	uses := e.ruleset().Allowed(core.UsingItem{Who: cID, Item: h}).Offers
	if len(uses) == 1 && uses[0].Not != "" {
		e.refuse(uses[0].Not)
		return core.ErrNotNow
	}
	if len(uses) == 0 {
		return core.ErrNotNow
	}
	p := pendingCall{kind: callItem, member: e.sel, item: h, spell: spell, target: target}
	own := uses[0].Action.(core.UseItem).Spell == ""
	if own {
		spell = int(e.itemCard(cID, h).Casts)
	} else if !slices.ContainsFunc(uses, func(o core.Offer) bool { return o.Action.(core.UseItem).Spell == e.spellID(spell) }) {
		spells := make([]Spell, len(uses))
		for k, o := range uses {
			spells[k] = Spell{ID: e.spellNumber(o.Action.(core.UseItem).Spell), Name: o.Name}
		}
		return e.askSpell(spells, p)
	}
	s := byte(spell)
	if !e.inFight() {
		e.do(core.StopUsingItem{})
		p.spell = spell
		yes, on, err := e.castAsk(int(s), true, target, sure, p)
		if err != nil {
			return err
		}
		u := core.UseItem{User: cID, Item: h, Target: on, Sure: yes}
		if !own {
			u.Spell = e.spellID(spell)
		}
		e.as(e.sel) // the user acts (WordWho)
		e.do(u)
		return nil
	}
	e.closeViews() // the turn goes on once, after the aim (turnOn below)
	run := castWhat{kind: cwItem, c: cID, n: spell, it: h, own: own}
	if e.fact(cID, core.FactSide).(core.Side) == core.SideParty { // castAsk: the aims first
		e.aimSpell(cID, int(s), run)
	} else {
		e.castRun(run, nil)
	}
	if e.cs.aim == nil && e.cs.turn != nil && e.resumed() {
		e.turnDone()
	}
	e.turnOn()
	return nil
}

// uncovered: the sheet and the inventory went (opened by the presenter);
// what is under them draws itself again or goes on.
func (e *Engine) uncovered() {
	switch {
	case e.stateRedraw():
	case e.waiting(PendFight):
		if e.cs.actor != core.NoChar {
			e.setWorldWord(core.WordWho, e.sel)
			e.do(core.SetSpeed{Who: e.cs.actor})
		}
		e.turnOn()
	case e.top() == onWait:
		e.waitRedrawNow()
	case e.exploring():
		e.awaitCommand()
	}
}
