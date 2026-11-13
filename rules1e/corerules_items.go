package rules1e

import (
	"errors"
	"fmt"
	"gobox/rules"
	"slices"
	"strconv"
)

// resolveItems resolves the inventory's item actions (por sheetcalls: Ready,
// Drop, Sell, Buy); ok: a was one of them. Who acts as the rules' actor.
func (r *Rules) resolveItems(sit rules.Situation, a rules.Action) (o rules.Outcome, ok bool, err error) {
	s := r.S
	switch a := a.(type) {
	case rules.StopUsingItem:
		s.UsingItem = false
		return rules.Outcome{}, true, nil
	case rules.UseItem:
		if _, ok := sit.(rules.Fight); ok {
			o, err := r.useItem(sit, a)
			return o, true, err
		}
		return o, false, nil
	case rules.DestroyItems:
		for _, c := range r.party(a.Whom) {
			s.DestroyItems(c, a.Item, a.MagicalOnly)
		}
		return o, true, nil
	case rules.ApplyItems:
		c := r.chars[a.Who]
		if c == nil {
			return o, true, errNoChar
		}
		s.ApplyItems(c)
		return o, true, nil
	case rules.Ready:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil {
			return o, true, ErrNotNow
		}
		return r.ready(c, it, int(a.Count)), true, nil
	case rules.Unready:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil || it.Readied == 0 {
			return o, true, ErrNotNow
		}
		return r.ready(c, it, 0), true, nil
	case rules.DropItem:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil {
			return o, true, ErrNotNow
		}
		if s.CanLoseItem(it) {
			s.Emit(rules.Sounded{Cue: rules.SoundCue("drop")})
			s.RemoveItem(c, it)
		}
		s.Who = s.PartyIndex(c)
		s.ApplyItems(c)
		return o, true, nil
	case rules.DepositItem:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil {
			return o, true, ErrNotNow
		}
		if s.CanLoseItem(it) {
			kept := *it
			s.Treasure = slices.Insert(s.Treasure, 0, &kept)
			s.RemoveItem(c, it)
		}
		s.Who = s.PartyIndex(c)
		s.ApplyItems(c)
		return o, true, nil
	case rules.Sell:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil {
			return o, true, ErrNotNow
		}
		if s.CanLoseItem(it) {
			s.RemoveItem(c, it)
			s.Purse.AddCopper(uint64(it.SellValue()), rules.GoldBoxRates)
			s.Emit(rules.Sounded{Cue: rules.SoundCue("store5")})
			o.Effects = append(o.Effects, rules.Traded{Who: a.Who, Item: a.Item, Sold: true})
		}
		s.Who = s.PartyIndex(c)
		s.ApplyItems(c)
		return o, true, nil
	case rules.Buy:
		c, it := r.chars[a.Who], r.items[a.Item]
		if c == nil || it == nil {
			return o, true, ErrNotNow
		}
		s.Who = s.PartyIndex(c)
		price := uint64(it.Price)
		if a.Price > 0 {
			price = uint64(a.Price)
		}
		switch {
		case price > s.Purse.Gold:
			o.Effects = append(o.Effects, rules.Refused{Who: a.Who, Why: rules.NoMoney})
		case !s.BuyItem(it):
			s.Purse.Gold -= price
			s.Emit(rules.Sounded{Cue: rules.SoundCue("store5")})
			o.Effects = append(o.Effects, rules.Traded{Who: a.Who, Item: a.Item})
		}
		return o, true, nil
	case rules.IdentifyItem:
		c, it := r.chars[a.Who], r.Held(a.Item)
		if c == nil || it == nil {
			return rules.Outcome{}, true, errNoItem
		}
		return r.identify(c, it), true, nil
	case rules.Valuate:
		it, ok := s.valuate(a.Jewel)
		if !ok {
			return rules.Outcome{}, true, ErrNotNow
		}
		r.valued = it
		return o, true, nil
	case rules.KeepValued:
		c, it := r.chars[a.Who], r.valued
		if c == nil || it == nil {
			return rules.Outcome{}, true, ErrNotNow
		}
		r.valued = nil
		s.Who = s.PartyIndex(c)
		if a.Keep && !s.Overloaded(it, c) {
			s.AddItemCopy(c, it)
		} else {
			s.Emit(rules.Sounded{Cue: rules.SoundCue("store5")})
			s.Purse.Gold += uint64(it.Price)
		}
		s.ApplyItems(c)
		return o, true, nil
	}
	return o, false, nil
}

// ready readies or unreadies n of it (0: all), as the player's wield: a
// sound when the slots changed (the computer readies in silence, aiPickWeapon).
func (r *Rules) ready(c *Character, it *Item, n int) rules.Outcome {
	s := r.S
	s.Who = s.PartyIndex(c)
	slots := c.Slots
	s.ReadySome(c, it, n)
	s.ApplyItems(c)
	if c.Slots != slots {
		s.Sound("wield")
	}
	return rules.Outcome{}
}

// useItemOutside = por's UseItem out of a fight, after its questions: the
// Item casts its own spell (UsingItem) or the scroll's Spell, on Target when
// the spell asks for one (Sure: yes to "Use it?"); the items are applied
// again. The acting member (WordWho) is the user, as the Engine set it.
func (r *Rules) useItemOutside(a rules.UseItem) (rules.Outcome, error) {
	c, it := r.chars[a.User], r.Held(a.Item)
	if c == nil || it == nil {
		return rules.Outcome{}, errNoChar
	}
	s := r.S
	own := !s.IsScroll(it)
	n := int(it.Spells[1])
	if !own {
		if n = slices.Index(SpellIDs[:], a.Spell); n <= 0 {
			return rules.Outcome{}, errBadValue
		}
		s.takeScroll(c, it)
	}
	s.UsingItem = own
	s.UseItem(it, SpellID(n), CastReply{Yes: a.Sure, On: r.chars[a.Target]})
	s.ApplyItems(c)
	return rules.Outcome{}, nil
}

// usingItem = the checks of por's UseItem (ovr20:1A8A): the inventory
// offers using items here (InvFlags), the item has a use (Usable), an item
// with a doll slot is readied (extra: items with no slot are used from the
// party inventory), and a scroll's spells as ovr23:044E lists them for the
// reader (Read Magic, or a cleric's cleric scroll, uncovers it). A pure
// query: UseItem's Resolve remembers and uncovers the scroll (takeScroll).
func (r *Rules) usingItem(u rules.UsingItem) rules.Allowed {
	s := r.S
	c, it := r.chars[u.Who], r.Held(u.Item)
	if c == nil || it == nil {
		return rules.Allowed{}
	}
	if use, _, _ := s.InvFlags(c); !use || !s.Usable(it) {
		return rules.Allowed{}
	}
	if it.Readied == 0 && s.DollSlot(c, it) >= 0 {
		return rules.Allowed{Offers: []rules.Offer{{Not: "Must be Readied"}}}
	}
	use := rules.UseItem{User: u.Who, Item: u.Item, Target: rules.NoChar}
	if !s.IsScroll(it) {
		return rules.Allowed{Offers: []rules.Offer{{Action: use}}}
	}
	var a rules.Allowed
	if it.Hidden != 0 && !s.readsScroll(c, it) {
		return a
	}
	for _, v := range it.Spells {
		if n := v.Spell(); v > 0 && s.T().hasSpell(n) {
			use.Spell = SpellIDs[n]
			a.Offers = append(a.Offers, rules.Offer{Action: use, Name: s.SpellName(int(n))})
		}
	}
	return a
}

// valuate = the Gems and Jewelry picks of ovr22:19B2: a gem or jewel of the
// purse is taken and valued (gold pieces) as the game's Appraisal says, its
// item made; false: none to value.
func (s *RuleState) valuate(jewel bool) (*Item, bool) {
	p := &s.Purse
	a := s.T().Appraisal
	if a == nil {
		a = &PorAppraisal
	}
	bands, word := a.Gems, a.GemWord
	switch {
	case !jewel && p.Gems > 0:
		p.Gems--
	case jewel && p.Jewelry > 0:
		p.Jewelry--
		bands, word = a.Jewels, a.JewelWord
	default:
		return nil, false
	}
	var v uint16
	d := s.Dice(1, 100)
	for _, b := range bands {
		if d <= b.Hi {
			v = b.Base
			if b.Span > 0 {
				v += uint16(s.Random(b.Span))
			}
			break
		}
	}
	it := &Item{Type: a.Type, Price: v, Weight: 1}
	it.Words[2] = word
	return it, true
}

// Appraisal is how a game values a gem or jewel (rules.json appraisal; nil:
// PorAppraisal): the item made and a d100's value bands.
type Appraisal struct {
	Type      ItemType
	GemWord   WordID
	JewelWord WordID
	Gems      []ValueBand
	Jewels    []ValueBand
}

// ValueBand is the value of a d100 up to Hi: Base, plus Random(Span) when
// Span is set.
type ValueBand struct {
	Hi   int
	Base uint16
	Span int `json:",omitempty"`
}

// PorAppraisal is POR's (ovr22:19B2).
var PorAppraisal = Appraisal{
	Type: ItemValuable, GemWord: rules.GemNameWord, JewelWord: rules.JewelNameWord,
	Gems: []ValueBand{{Hi: 25, Base: 10}, {Hi: 50, Base: 50}, {Hi: 70, Base: 100}, {Hi: 90, Base: 500}, {Hi: 99, Base: 1000}, {Hi: 100, Base: 5000}},
	Jewels: []ValueBand{{Hi: 10, Base: 100, Span: 900}, {Hi: 20, Base: 200, Span: 1000}, {Hi: 40, Base: 300, Span: 1500}, {Hi: 50, Base: 500, Span: 2500},
		{Hi: 70, Base: 1000, Span: 5000}, {Hi: 90, Base: 2000, Span: 6000}, {Hi: 100, Base: 2000, Span: 10000}},
}

// SotsbAppraisal is SotSB's and PoD's: half POR's values, type 47, words
// 122 and 123 (sotsb-re and pod-re money.pas AppraiseMenu).
var SotsbAppraisal = Appraisal{
	Type: 47, GemWord: 122, JewelWord: 123,
	Gems: []ValueBand{{Hi: 25, Base: 5}, {Hi: 50, Base: 25}, {Hi: 70, Base: 50}, {Hi: 90, Base: 250}, {Hi: 99, Base: 500}, {Hi: 100, Base: 2500}},
	Jewels: []ValueBand{{Hi: 10, Base: 50, Span: 450}, {Hi: 20, Base: 100, Span: 500}, {Hi: 40, Base: 150, Span: 750}, {Hi: 50, Base: 250, Span: 1250},
		{Hi: 70, Base: 500, Span: 2500}, {Hi: 90, Base: 1000, Span: 3000}, {Hi: 100, Base: 1000, Span: 5000}},
}

var errNoItem = errors.New("rules1e: no such item")

// handleList is items as the engine holds them.
func (r *Rules) handleList(l []*Item) rules.Handles {
	h := make(rules.Handles, len(l))
	for i, it := range l {
		h[i] = r.Handle(it)
	}
	return h
}

// itemCard is it as c (nil: no one) sees it (por itemState, invItem).
func (r *Rules) itemCard(c *Character, it *Item) rules.ItemCard {
	s := r.S
	s.ItemName(c, it, false)
	g := rules.ItemCard{Item: r.Handle(it), Name: it.Name(), Readied: it.Readied != 0, Identified: !it.Unknown(), Price: int32(it.Price), Weight: int32(it.Weight),
		Type: int32(it.Type), Count: int32(it.Count), Usable: s.Usable(it), Sale: rules.Copper(it.SellValue()), Words: s.wordNames(it), Casts: int32(it.Spells[1])}
	if c != nil {
		g.CannotUse = !s.CanUseItem(c, it)
		g.Look = s.ItemLines(it, s.ReadiedLike(c, it))
		g.Tab, g.Slot = s.ItemTab(c, it)
	} else {
		g.Look = s.ItemLines(it, nil)
	}
	g.LookPlain = s.ItemLines(it, nil)
	g.Scribing = s.Scribing(it)
	if it.Hidden != 0 {
		h := *it
		h.Hidden = 0
		s.ItemName(c, &h, false)
		g.Hidden = h.Name()
		s.ItemName(c, it, false)
	}
	for k, sp := range it.Spells {
		if s.IsScroll(it) && sp.Spell() != 0 {
			if len(g.Scroll) == 0 {
				g.Scribe = sp.Scribing() && c != nil && s.ScribesIt(c, it, k) && c.SpellBook[sp.Spell()-1] == 0
			}
			g.Scroll = append(g.Scroll, rules.ScrollSpell{Spell: SpellIDs[sp.Spell()], Number: int32(sp.Spell()), Scribing: sp.Scribing()})
		}
	}
	return g
}

// DamageText is c's weapon damage as the sheet words it (por combat_engine damageText).
func DamageText(c *Character) string {
	d := strconv.Itoa(int(c.Dice[0])) + "d" + strconv.Itoa(int(c.Sides[0]))
	if b := int8(c.Bonus[0]); b > 0 {
		d += "+" + strconv.Itoa(int(b))
	} else if b < 0 {
		d += "-" + strconv.Itoa(-int(b)) // abs in 16 bits: -128 shows 128
	}
	return d
}

// identify = ovr20:1F52 (item): 200 gp shows the hidden words (the player
// took the deal).
func (r *Rules) identify(c *Character, it *Item) (o rules.Outcome) {
	s := r.S
	if !s.Pay(200) {
		return o
	}
	s.Emit(rules.Sounded{Cue: rules.SoundCue("pseudo_id")})
	s.ItemName(c, it, false)
	done := rules.ItemIdentified{Item: r.Handle(it), Revealed: it.Hidden != 0}
	if done.Revealed {
		it.Hidden = 0
		s.ItemName(c, it, false)
	}
	o.Effects = append(o.Effects, done)
	return o
}

// checkItems answers the Checks of the items topic.
func (r *Rules) checkItems(k rules.Check) (rules.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case rules.Fit:
		c := r.chars[k.Who]
		return rules.Answer{Yes: c != nil && c.Okay != 0}, true
	case rules.PoolCap:
		return rules.Answer{Value: int32(s.PoolCap())}, true
	case rules.ItemTableMissing:
		return rules.Answer{Yes: !slices.ContainsFunc(s.Items, func(b ItemBase) bool { return b != ItemBase{} })}, true
	case rules.PoolCount:
		return rules.Answer{Value: int32(len(s.Pool))}, true
	case rules.CarryCount:
		if c := r.chars[k.Who]; c != nil {
			return rules.Answer{Value: int32(len(s.Carried(c)))}, true
		}
		return rules.Answer{}, true
	case rules.HasRoom:
		c, it := r.chars[k.Who], r.Held(k.Item)
		return rules.Answer{Yes: c != nil && it != nil && !r.S.overloadedNow(it, c)}, true
	case rules.ValuedItem:
		it := r.valued
		if it == nil {
			return rules.Answer{Value: int32(rules.NoItem)}, true
		}
		return rules.Answer{Value: int32(r.Handle(it)), High: int32(it.Price), Yes: it.Words[2] == WordID(rules.JewelNameWord)}, true
	default:
		return rules.Answer{}, false
	}
}

// factItems answers the Facts of the items topic.
func (r *Rules) factItems(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case rules.ItemByNumber:
		itemID := s.kind(ItemType(f.N))
		if itemID == 0 {
			return nil, true, fmt.Errorf("por: item type %#x has no dnd id (rules1e ids.txt)", f.N)
		}
		return itemID, true, nil
	case rules.ListedFact:
		if f.Goods {
			return r.handleList(s.Listed(nil, s.Treasure)), true, nil
		}
		if c == nil {
			return nil, true, errNoChar
		}
		return r.handleList(s.Listed(c, s.Carried(c))), true, nil
	case rules.TreasureRecordsFact:
		var items []Item
		for _, it := range slices.Backward(s.Treasure) {
			items = append(items, *it)
		}
		items, err := s.GameRecords(items)
		if err != nil {
			return nil, true, err
		}
		return r.ItemRecords(items), true, nil
	case rules.ItemCardFact:
		it := r.Held(f.Item)
		if it == nil {
			return nil, true, errNoItem
		}
		return r.itemCard(c, it), true, nil
	case rules.IdentifiableFact:
		if i := slices.Index(s.Party, c); c != nil && i >= 0 {
			return r.handleList(s.Identifiable(i)), true, nil
		}
		return nil, true, errNoChar
	case rules.DamageFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return rules.Text(DamageText(c)), true, nil
	case rules.InventoryFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.inventoryView(c), true, nil
	default:
		return nil, false, nil
	}
}

// inventoryView builds the snapshot of c's inventory screen.
func (r *Rules) inventoryView(c *Character) rules.InventoryView {
	s := r.S
	use, sell, switches := s.InvFlags(c)
	v := rules.InventoryView{Member: slices.Index(s.Party, c), Cap: s.PoolCap(), Doll: make([]*rules.ItemCard, rules.DollSlots), Use: use, Sell: sell, Switch: switches}
	for _, it := range s.Listed(c, s.Carried(c)) { // the inventory's order
		x := r.itemCard(c, it)
		if it.Readied != 0 {
			if d := s.DollSlot(c, it); d >= 0 && v.Doll[d] == nil {
				v.Doll[d] = &x
				continue
			}
		}
		v.Pool = append(v.Pool, x)
	}
	return v
}

// readsScroll: c can read scroll it (ovr23:044E): Read Magic, or a cleric's
// cleric scroll, uncovers it.
func (s *RuleState) readsScroll(c *Character, it *Item) bool {
	return s.FindAffect(c, AffReadMagic) != nil || c.Levels[ClsCleric] > 0 && c.Levels[ClsCleric] < 0x80 && s.Base(it.Type).Slot == 0xC
}

// takeScroll = the state change of ovr23:044E for c using scroll it: the
// scroll is remembered in DS:6B80 and uncovered when c reads it.
func (s *RuleState) takeScroll(c *Character, it *Item) {
	s.UsedScroll = it
	if s.readsScroll(c, it) {
		it.Hidden = 0
	}
}
