package engine

import (
	"errors"
	"gobox/rules"
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
	return !e.inFight() || e.waiting(PendFight) && e.cs.actor != rules.NoChar && e.cs.aim == nil
}

// OpenSheet puts member n's (< 0: the selected one; in a fight the one
// whose turn it is) sheet up.
func (e *Engine) OpenSheet(n int) error {
	if !e.viewable() || e.view.on {
		return rules.ErrNotNow
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
		return rules.ErrNotNow
	}
	if !e.inFight() {
		e.actor(n)
	}
	if direct && !e.view.on {
		if e.carryCount(e.selected()) == 0 {
			e.say("No items")
			return rules.ErrNotNow
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
		return rules.ErrNotNow
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
		return rules.ErrNotNow
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
		return rules.ErrNotNow
	}
	if e.inFight() && (e.cs.actor == rules.NoChar || e.partyIDs()[n] != e.cs.actor) { // a fight: only the member whose turn it is (BUG-CLASSES 40)
		return rules.ErrNotNow
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
func (e *Engine) carriedBy(h rules.ItemHandle) (rules.CharID, rules.ItemHandle) {
	id := e.selected()
	if !e.snap.invOn || !slices.Contains(e.listedItems(id), h) {
		return id, rules.NoItem
	}
	return id, h
}

// Ready readies or unreadies item h (x of the stack, 0: all of it); to is
// where a weapon with the choice goes (rules.ReadyTo).
func (e *Engine) Ready(h rules.ItemHandle, x int, to rules.ReadyTo) error {
	return e.readyAction(h, func(cID rules.CharID, h rules.ItemHandle) rules.Action {
		return rules.Ready{Who: cID, Item: h, Count: int32(x), To: to}
	})
}

// Unready unreadies item h, refused when it is not readied. It resolves rules.Unready: rules.Ready is a toggle
// in the Pool's ruleset but only readies in BR's (viewchar.pas Ready and Unready are two keys).
func (e *Engine) Unready(h rules.ItemHandle) error {
	if id, h := e.carriedBy(h); h != rules.NoItem && !e.itemCard(id, h).Readied {
		return errors.New("not readied")
	}
	return e.readyAction(h, func(cID rules.CharID, h rules.ItemHandle) rules.Action {
		return rules.Unready{Who: cID, Item: h}
	})
}

// readyAction resolves the ready action of the selected member's listed item h, then lets the turn's weapon follow.
func (e *Engine) readyAction(h rules.ItemHandle, act func(rules.CharID, rules.ItemHandle) rules.Action) error {
	cID, h := e.carriedBy(h)
	if h == rules.NoItem {
		return rules.ErrNotNow
	}
	if _, err := e.ruleset().Resolve(nil, act(cID, h), nil); err != nil {
		return err
	}
	if t := e.snap.turn; t != nil && cID == e.cs.actor { // the turn's weapon follows (BUG-CLASSES 40), as turnChoices sets it
		e.turnWeapons(cID, t)
	}
	return nil
}

// drop drops item h (the front asked ovr20's "drop It?").
func (e *Engine) drop(h rules.ItemHandle) error {
	cID, h := e.carriedBy(h)
	if h == rules.NoItem {
		return rules.ErrNotNow
	}
	if e.vault != nil { // Deposit: the item stays in the vault (pod's dropItem)
		_, err := e.ruleset().Resolve(nil, rules.DepositItem{Who: cID, Item: h}, nil)
		return err
	}
	_, err := e.ruleset().Resolve(nil, rules.DropItem{Who: cID, Item: h}, nil)
	return err
}

// sell sells item h where the inventory sells (the shop's). A ruleset that asks first (Outcome.Confirm: BR's
// "I'll give you N credits...") has the question asked (askSure: Yes sells, sure); sure: it was asked.
func (e *Engine) sell(h rules.ItemHandle) error {
	cID, h := e.carriedBy(h)
	if h == rules.NoItem || !e.inventoryView(cID).Sell {
		return rules.ErrNotNow
	}
	a := rules.Sell{Who: cID, Item: h}
	o, err := e.ruleset().Resolve(nil, a, nil)
	if err != nil {
		return err
	}
	if o.Confirm != "" { // the ruleset asks first (XXVc: "I'll give you ..."): the front has asked already, with the Sale it was shown (game.Data.Confirms)
		if o, err = e.ruleset().Resolve(nil, rules.Confirmed{Action: a}, nil); err != nil {
			return err
		}
	}
	e.traded(o, "Sold!")
	return nil
}

// Buy buys ware h of the shop's (Inventory.Wares).
func (e *Engine) Buy(h rules.ItemHandle) error {
	if !e.snap.invOn || !e.snap.shop || !slices.Contains(e.goods(), h) {
		return rules.ErrNotNow
	}
	card := e.itemCard(rules.NoChar, h)
	o, err := e.ruleset().Resolve(nil, rules.Buy{Who: e.selected(), Item: h, Price: e.shopPrice(card.Price, card.Count)}, nil)
	if err != nil {
		return err
	}
	return refuseIf(!e.traded(o, "Bought!"), e.ext.note)
}

// traded shows what a trade came to and says whether it went through (rules.Traded). The ruleset's own words
// (Narrated) are shown when it gives some: as the note, or, when the trade did not go through, as its refusal;
// else the engine's: done, or "Not enough Money." for a purse that could not pay.
func (e *Engine) traded(o rules.Outcome, done string) (through bool) {
	var told string
	noMoney := false
	for _, f := range o.Effects {
		switch f := f.(type) {
		case rules.Traded:
			through = true
		case rules.Narrated:
			told = f.Text
		case rules.Refused:
			noMoney = noMoney || f.Why == rules.NoMoney
		}
	}
	switch {
	case told != "" && through:
		e.say(told)
	case told != "":
		e.refuse(told)
	case through:
		e.say(done)
	case noMoney:
		e.refuse("Not enough Money.")
	}
	return through
}

// reload buys the ammunition of kind for member n's item h, as the ruleset offers it (Reloading): not offered, ErrNotNow.
func (e *Engine) reload(n int, h rules.ItemHandle, kind rules.ReloadKind) error {
	if !e.waiting(PendShop) || n < 0 || n >= e.partySize() {
		return rules.ErrNotNow
	}
	id := e.partyIDs()[n]
	want := rules.Reload{Who: id, Item: h, Kind: kind}
	if !slices.ContainsFunc(e.ruleset().Allowed(rules.Reloading{Who: id, Item: h}).Offers, func(o rules.Offer) bool { return o.Action == want }) {
		return rules.ErrNotNow
	}
	o, err := e.ruleset().Resolve(nil, want, nil)
	if err != nil {
		return err
	}
	for _, f := range o.Effects {
		if t, ok := f.(rules.Narrated); ok {
			e.say(t.Text)
		}
	}
	return nil
}

// UseItem = ovr20:1A8A (item): item h casts the scroll's spell or its own
// (+0x3D, charges +0x3C). spell, target, sure are the picks it asked
// for (a scroll's spell; the target and the yes/no as UsePower). In a fight the inventory and the sheet close, the spell aims
// (a phase of the turn) and a use ends the turn.
func (e *Engine) UseItem(h rules.ItemHandle, spell, target int, sure bool) error {
	cID, h := e.carriedBy(h)
	uses := e.ruleset().Allowed(rules.UsingItem{Who: cID, Item: h, NoMagic: e.noMagic()}).Offers
	if len(uses) == 1 && uses[0].Not != "" {
		e.refuse(uses[0].Not)
		return rules.ErrNotNow
	}
	if len(uses) == 0 {
		return rules.ErrNotNow
	}
	p := pendingCall{kind: callItem, member: e.sel, item: h, spell: spell, target: target}
	own := uses[0].Action.(rules.UseItem).Spell == ""
	if own {
		spell = int(e.itemCard(cID, h).Casts)
	} else if !slices.ContainsFunc(uses, func(o rules.Offer) bool { return o.Action.(rules.UseItem).Spell == e.spellID(spell) }) {
		spells := make([]Spell, len(uses))
		for k, o := range uses {
			spells[k] = Spell{ID: e.spellNumber(o.Action.(rules.UseItem).Spell), Name: o.Name}
		}
		return e.askSpell(spells, p)
	}
	s := byte(spell)
	if !e.inFight() {
		e.do(rules.StopUsingItem{})
		p.spell = spell
		yes, on, err := e.castAsk(int(s), true, target, sure, p)
		if err != nil {
			return err
		}
		u := rules.UseItem{User: cID, Item: h, Target: on, Sure: yes, In3D: e.in3D()}
		if !own {
			u.Spell = e.spellID(spell)
		}
		e.as(e.sel) // the user acts (rules.ActFor)
		e.do(u)
		return nil
	}
	e.closeViews() // the turn goes on once, after the aim (turnOn below)
	run := castWhat{kind: cwItem, c: cID, n: spell, it: h, own: own}
	if e.fact(cID, rules.FactSide).(rules.Side) == rules.SideParty { // castAsk: the aims first
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
		if e.cs.actor != rules.NoChar {
			e.as(e.sel)
			e.do(rules.SetSpeed{Who: e.cs.actor})
		}
		e.turnOn()
	case e.top() == onWait:
		e.waitRedrawNow()
	case e.exploring():
		e.awaitCommand()
	}
}
