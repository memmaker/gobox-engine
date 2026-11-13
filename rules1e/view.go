package rules1e

import (
	"gobox/core/dnd"
	"slices"
)

// ovr20: the View screen (character sheet, items, money, spell lists).

// Paladin powers from CotAB (ovr020). Two port-only affect types time them
// (affLaidOnHands, affCuredDisease): one heal a day, one cure disease a week
// (one is all a paladin of PoR's level 8 cap gets).

// PaladinCan = CotAB CanCastHeal/CanCastCureDiseases: out of combat, okay,
// and the power not used yet.
func (e *RuleState) PaladinCan(c *Character, used AffectType) bool {
	return int8(c.Levels[ClsPaladin]) > 0 && e.BeginFlag != 5 && c.Status == StatusOkay && e.FindAffect(c, used) == nil
}

// CanLoseItem = ovr20:0D70 (item): readied items must be unreadied first.
// (Its question for a scroll someone was going to scribe from is the
// front end's: core.ItemCard.Scribing.)
func (e *RuleState) CanLoseItem(it *Item) bool {
	if it.Readied != 0 {
		e.notice("Must be unreadied")
		return false
	}
	return true
}

// Scribing: someone was going to scribe from the scroll (ovr20:0D70).
func (e *RuleState) Scribing(it *Item) bool {
	return e.IsScroll(it) && slices.ContainsFunc(it.Spells[:], func(p ItemPower) bool { return p&ScrollScribing != 0 })
}

// Usable: the item has a Use (a scroll, or a spell of its own).
func (e *RuleState) Usable(it *Item) bool {
	return e.IsScroll(it) || castsSpell(e, it)
}

// SwapReady readies it in place of the item in its slot (extra: a swap,
// not readyItem's "already using"), else readies or unreadies it.
func (e *RuleState) SwapReady(c *Character, it *Item) {
	if o := e.InTheWay(c, it); o != nil && e.Base(it.Type).Classes&c.ClassMask != 0 {
		e.ReadyItem(o)
		if o.Readied != 0 { // cursed, or no room for it
			return
		}
		e.ApplyItems(c) // the slot and hands free
	}
	// and whatever else holds the hands it needs: a two-handed weapon takes
	// both the weapon and the shield off
	if it.Readied == 0 && e.Base(it.Type).Classes&c.ClassMask != 0 {
		for _, s := range []int{SlotWeapon, SlotShield} {
			o := c.Slots[s]
			if o == nil || o == it || int(e.Base(it.Type).Hands)+int(c.Hands) <= 2 {
				continue
			}
			e.ReadyItem(o)
			if o.Readied != 0 {
				return
			}
			e.ApplyItems(c)
		}
	}
	e.ReadyItem(it)
}

// ReadySome (extra): readies or unreadies n of a stack (n 0 or the whole
// stack: all of it, swapReady), the rest staying where it was; a part
// readied beside a like readied stack joins it (arrows), so does one
// unreadied beside a like stack in the party inventory (addToPool).
func (e *RuleState) ReadySome(c *Character, it *Item, n int) {
	if n <= 0 || n >= int(it.Count) || it.Count == 0 || it.Spells[0] >= 2 {
		if o := e.InTheWay(c, it); it.Readied == 0 && o != nil && sameStack(o, it) && int(o.Count)+int(it.Count) <= 0xFF {
			o.Count += it.Count // a whole stack beside a like readied one: it joins
			e.RemoveItem(c, it)
			return
		}
		e.SwapReady(c, it)
		return
	}
	part := *it
	part.Count = byte(n)
	it.Count -= byte(n)
	if it.Readied != 0 { // n of them go back to the party inventory
		part.Readied = 0
		if e.poolFull(&part) {
			it.Count += byte(n)
			e.notice("No room in the party items")
			return
		}
		e.addToPool(&part)
		return
	}
	if o := e.InTheWay(c, &part); o != nil && sameStack(o, &part) && int(o.Count)+n <= 0xFF {
		o.Count += byte(n)
		return
	}
	e.Pool = append(e.Pool, &part) // n of them are readied: the part is an item of its own
	e.SwapReady(c, &part)
	if part.Readied == 0 { // refused: back into the stack
		e.Pool = slices.DeleteFunc(e.Pool, func(p *Item) bool { return p == &part })
		it.Count += byte(n)
	}
}

// InTheWay is the readied item of c whose slot it would take (readyItem's
// "already using"), nil for none.
func (e *RuleState) InTheWay(c *Character, it *Item) *Item {
	if it.Readied != 0 {
		return nil
	}
	slot := e.readySlot(it.Type)
	switch {
	case e.offHand(c, it):
		slot = SlotShield
	case e.isKind(it.Type, dnd.ItemArrow):
		slot = SlotArrows
	case e.isKind(it.Type, dnd.ItemQuarrel):
		slot = SlotBolts
	case slot == 9:
		slot = slotRing2
	case slot > 8:
		return nil
	}
	return c.Slots[slot]
}

// ReadyItem = ovr20:14D0 (item): ready or unready, with the reasons it
// cannot be readied and the item's own effect (+0x3E > 0x7F).
func (e *RuleState) ReadyItem(it *Item) {
	effect := it.Spells[2] > 0x7F
	c := e.Acting()
	if it.Readied != 0 {
		if it.Cursed != 0 {
			e.notice("It's Cursed")
			return
		}
		if e.partyMember(c) && e.poolFull(it) && !(e.BeginFlag == 5 && c.Auto != 0) { // extra: it goes back into the party inventory; the computer's turn never stalls on a full one (it overflows, addToPool)
			e.notice("No room in the party items")
			return
		}
		it.Readied = 0
		if effect {
			e.itemAffect(it, c, 1)
		}
		return
	}
	if e.DollSlot(c, it) < 0 { // extra: used straight from the party inventory
		e.notice("Nothing to ready")
		return
	}
	why := 0
	typ := it.Type
	if int(e.Base(typ).Hands)+int(c.Hands) > 2 {
		why = 3
	}
	slot := e.readySlot(typ)
	offHand := e.offHand(c, it) // extra: a second weapon goes into the off hand
	if offHand {
		slot = SlotShield
	}
	switch {
	case slot <= 8:
		if c.Slots[slot] != nil {
			why = 2
		}
	case slot == 9:
		if c.Slots[slotRing2] != nil {
			why = 2
		}
	}
	if e.isKind(typ, dnd.ItemArrow, dnd.ItemQuarrel) {
		for _, s := range []int{SlotArrows, SlotBolts} {
			if c.Slots[s] != nil {
				why, slot = 2, s
			}
		}
	}
	if e.Base(typ).Classes&c.ClassMask == 0 {
		why = 1
	}
	switch why {
	case 0:
		e.takeFromPool(c, it) // extra: out of the party inventory
		it.Readied = 1
		if offHand {
			it.Readied = ReadiedOffHand
		}
		if effect {
			e.itemAffect(it, c, 0)
		}
	case 1:
		e.notice("Wrong Class")
	case 2:
		// slot 9 (ring 2 taken) names Slots[9], which may be empty: the
		// original then reads through a nil pointer.
		name := ""
		if o := c.Slots[slot]; o != nil {
			e.ItemName(c, o, false)
			name = o.Name()
		}
		e.notice("already using " + name)
	case 3:
		if !(e.BeginFlag == 5 && c.Auto != 0) {
			e.notice("Your hands are full!")
		}
	}
}

// itemAffect calls ovr25:0000 with the item standing in for the affect
// record, as the original does.
func (e *RuleState) itemAffect(it *Item, c *Character, flag byte) {
	a := &Affect{Item: it}
	// the affect fields are the item's first bytes, as in the original
	n := it.PName
	a.Type, a.Time, a.Data, a.Call = AffectType(n[0]), uint16(n[1])|uint16(n[2])<<8, byte(n[3]), byte(n[4])
	e.callAffect(AffectType(it.Spells[2]), c, a, flag)
}

// joinItems = ovr20:189D (item): every matching stack joins it, up to 255
// a stack. (The original compares +0x3C..+0x3E of each item with itself.)
func (e *RuleState) joinItems(c *Character, it *Item) {
	total := int(it.Count)
	for _, p := range append([]*Item(nil), c.Items...) {
		if p == it || !sameStack(p, it) {
			continue
		}
		if int(p.Count)+total <= 0xFF {
			total += int(p.Count)
			e.RemoveItem(c, p)
		} else {
			it.Count = 0xFF
			p.Count -= byte(0xFF - total)
			total = int(p.Count)
			it = p
		}
	}
	it.Count = byte(total)
}

// UseItem is ovr20:1A8A once spell is picked (usingItem: the item's own):
// it casts it; true when the turn is used.
func (e *RuleState) UseItem(it *Item, spell SpellID, r CastReply) bool {
	quit := false
	c := e.Acting()
	if e.UsingItem {
		e.CombatSay(c, "uses an item", true)
		e.ItemName(c, it, false)
		e.CombatSay(nil, "Item: "+it.Name(), false)
	}
	e.UsingItem = true
	e.CastSpell(spell, c.Auto, 0, r, &quit) // a cancelled aim: nothing used, the inventory stays
	e.UsingItem = false
	if e.BeginFlag == 5 && e.SpellRec(spell, 0xB) != 0 {
		quit = e.ClearCombatInfo(c)
	}
	if quit {
		if e.IsScroll(it) {
			e.scrollUsed(c, it, SpellID(spell))
		} else if it.Spells[0] > 0 {
			if it.Count > 1 {
				it.Count--
			} else {
				it.Spells[0]--
				if it.Spells[0] == 0 {
					e.RemoveItem(c, it)
				}
			}
		}
	}
	return quit
}

// SpellRec = DS:31FA + 16*spell.
func (e *RuleState) SpellRec(spell SpellID, k int) byte {
	return byte(e.T().Spells[spell][k])
}

// Overloaded = ovr20:275B (item, char): no room for one more item.
// extra: a party member takes it into the party inventory, which has a
// slot count instead of the original's 16 items and weight allowance.
func (e *RuleState) Overloaded(it *Item, c *Character) bool {
	e.ApplyItems(c)
	return e.overloadedNow(it, c)
}

// overloadedNow is Overloaded with c's items as last applied: a pure query.
func (e *RuleState) overloadedNow(it *Item, c *Character) bool {
	if e.partyMember(c) {
		return e.poolFull(it)
	}
	return c.ItemCount > 0xF
}
