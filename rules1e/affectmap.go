package rules1e

import (
	"errors"
	"fmt"
	"slices"
)

// AffectMap is a game's affect numbers as the unified affect ids (POR's
// numbers, and new ids for the affects only another game has): rules.json
// affectIds, written by the game's importer. Bytes enter as the game's
// numbers once, at the boundaries (Import, Spawn's monster records, the
// item library) and leave as them on Export; inside rules1e only ids exist.
type AffectMap map[byte]AffectType

// affectID is the game's affect number b as an id; nil (POR): b is the id.
func (t *Tables) affectID(b byte) (AffectType, error) {
	if t.AffectIDs == nil || b == byte(affAny) { // 0 is no affect in every game (an item granting none)
		return AffectType(b), nil
	}
	id, ok := t.AffectIDs[b]
	if !ok {
		return 0, fmt.Errorf("rules1e: affect number %d is not in the game's affectIds", b)
	}
	return id, nil
}

// gameAffect is affectID's inverse.
func (t *Tables) gameAffect(id AffectType) (byte, error) {
	if t.AffectIDs == nil || id == affAny {
		return byte(id), nil
	}
	for b, v := range t.AffectIDs {
		if v == id {
			return b, nil
		}
	}
	return 0, fmt.Errorf("rules1e: affect id %#x has no number in the game", id)
}

// affectsIn rewrites the affect numbers of affects and the power codes of
// items: each record's Type through affect; an item's power (+0x3E > 0x7F;
// scrolls hold spells there) through power, and its +0x3D through affect
// when that power grants the affect +0x3D names (grantsAffect).
func (e *RuleState) affectsIn(affects []AffectRecord, items []Item, affect func(AffectType) (AffectType, error), power func(ItemPower) (ItemPower, error), toIDs bool) error {
	for i := range affects {
		t, err := affect(affects[i].Type)
		if err != nil {
			return err
		}
		affects[i].Type = t
	}
	for i := range items {
		it := &items[i]
		if it.Spells[2] <= 0x7F || e.IsScroll(it) {
			continue
		}
		p, err := power(it.Spells[2])
		if err != nil {
			return err
		}
		unified := p // the unified power: after power on import, before it on export
		if !toIDs {
			unified = it.Spells[2]
		}
		if grantsAffect(unified) {
			g, err := affect(AffectType(it.Spells[1]))
			if err != nil {
				return err
			}
			it.Spells[1] = ItemPower(g)
		}
		it.Spells[2] = p
	}
	return nil
}

// grantsAffect tells whether the unified item power p grants the affect its
// item's +0x3D names (POR's itemGrantsAffect powers, the later games' 128).
func grantsAffect(p ItemPower) bool { return porAffects[AffectType(p)] == "itemGrantsAffect" }

// gameItems is a copy of a game's item records (the game's numbers) with
// unified ids: every item that enters the rules from the game's data comes
// through here or toIDs (TestItemsEnterConverted).
func (e *RuleState) gameItems(items []Item) ([]Item, error) {
	out := slices.Clone(items)
	return out, e.toIDs(nil, out)
}

// GameRecords is gameItems' inverse: a copy of items as the game's file
// has them (the treasure parity tests compare those bytes).
func (e *RuleState) GameRecords(items []Item) ([]Item, error) {
	out := slices.Clone(items)
	return out, e.toGame(nil, out)
}

// errNoItemTable: items cannot be converted before core.LoadItemTable
// (IsScroll reads the base records).
var errNoItemTable = errors.New("rules1e: items converted before the item table is loaded")

// toIDs and toGame are affectsIn's and scrollsIn's two directions.
func (e *RuleState) toIDs(affects []AffectRecord, items []Item) error {
	if len(items) > 0 && len(e.Items) == 0 {
		return errNoItemTable
	}
	if err := e.affectsIn(affects, items, func(t AffectType) (AffectType, error) {
		if t > 0xFF {
			return 0, fmt.Errorf("rules1e: affect number %d is not a byte", t)
		}
		return e.T().affectID(byte(t))
	}, func(p ItemPower) (ItemPower, error) {
		if p > 0xFF {
			return 0, fmt.Errorf("rules1e: item power %d is not a byte", p)
		}
		return e.T().powerID(byte(p))
	}, true); err != nil {
		return err
	}
	return e.scrollsIn(items, true)
}

func (e *RuleState) toGame(affects []AffectRecord, items []Item) error {
	if err := e.affectsIn(affects, items, func(t AffectType) (AffectType, error) {
		b, err := e.T().gameAffect(t)
		return AffectType(b), err
	}, func(p ItemPower) (ItemPower, error) {
		b, err := e.T().gamePower(p)
		return ItemPower(b), err
	}, false); err != nil {
		return err
	}
	return e.scrollsIn(items, false)
}

// GameAffect is an affect number as a game's data has it (before its map).
type GameAffect uint8

// Missing is the affect numbers of affects (a game's .spc and MON SPC
// records) that m lacks: a game's data must have none (each importer's test).
func (m AffectMap) Missing(affects []AffectFile) []GameAffect {
	var out []GameAffect
	for _, a := range affects {
		if _, ok := m[a.Type]; !ok && !slices.Contains(out, GameAffect(a.Type)) {
			out = append(out, GameAffect(a.Type))
		}
	}
	return out
}
