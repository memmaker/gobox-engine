package rules1e

import (
	"errors"
	"fmt"
	"slices"

	"gobox/core"
)

// ECL opcodes 0x25..0x3D, ported from ovr04 (continued from eclops.go).

// AddTreasureBlock = TREASURE's ITEM block: each 0x3F-byte record lands at the
// head of the treasure list in turn, so the last one leads.
func (e *RuleState) AddTreasureBlock(b []Item) {
	var add []*Item
	for _, it := range b {
		e.T().MendItem(&it)
		add = append(add, &it)
	}
	slices.Reverse(add)
	e.Treasure = append(add, e.Treasure...)
}

// errNoLibrary: a script asked for library items the ruleset's tables lack.
var errNoLibrary = errors.New("rules1e: no item library or random scroll in the tables (rules.json itemLibrary, randomScroll)")

// itemsKnown are Item.Hidden's name-word bits.
const itemsKnown = 7

// AddLibraryItems = pod TREASURE's and ITEM SET's items (O040:13E0,
// O040:251F): each library item at the head of the treasure in turn,
// core.RandomScrollItem a random scroll (O055:1ABE), empty entries none.
func (e *RuleState) AddLibraryItems(ids []int32, known bool) error {
	t := e.T()
	for _, id := range ids {
		var it *Item
		switch {
		case id == core.RandomScrollItem:
			if t.Scroll == nil {
				return errNoLibrary
			}
			var err error
			if it, err = e.RandomScroll(); err != nil {
				return err
			}
		case t.Library == nil:
			return errNoLibrary
		case id >= 1 && int(id) <= len(t.Library) && t.Library[id-1] != nil:
			one, err := e.gameItems([]Item{*t.Library[id-1]}) // the library is the game's
			if err != nil {
				return err
			}
			it = &one[0]
		default:
			continue
		}
		if known {
			it.Hidden &^= itemsKnown
		}
		e.Treasure = append([]*Item{it}, e.Treasure...)
	}
	return nil
}

// RandomScroll = O055:1ABE: Tables.Scroll's scroll, each spell rolled until
// it is a mage spell (the game's numbers), then held as unified ids.
func (e *RuleState) RandomScroll() (*Item, error) {
	s := e.T().Scroll
	it := &Item{Type: s.Type, Words: s.Words, Plus: s.Plus, Weight: s.Weight, Price: s.Price}
	for k := range it.Spells {
		for !slices.Contains(s.Mage, SpellID(it.Spells[k])) {
			it.Spells[k] = ItemPower(e.Dice(1, s.Last))
		}
	}
	for k, p := range it.Spells {
		id, err := e.T().spellNumber(SpellID(p))
		if err != nil {
			return nil, fmt.Errorf("RandomScroll: %w", err)
		}
		it.Spells[k] = ItemPower(id)
	}
	return it, nil
}

// AddRandomTreasure = TREASURE $80+n (ovr22:1010): n random items, each put
// at the head of the treasure list, made by the game's treasure rolls
// (rules.json randomTreasure; POR's, PorTreasure, when it has none).
func (e *RuleState) AddRandomTreasure(n int) error {
	t := e.treasureRolls()
	for range n {
		it, err := e.treasureItem(t, rollOf(e, t.Type))
		if err != nil {
			return err
		}
		e.Treasure = append([]*Item{it}, e.Treasure...)
	}
	return nil
}
