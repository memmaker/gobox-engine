package engine

import (
	"fmt"
	"gobox/asset"
	"gobox/rules"
)

// The view side of the Engine's ruleset (fold step E-R6): one-line helpers
// that resolve the Engine's characters to ids, until the Engine holds ids.

// okay is whether member i's status is okay.
func (e *Engine) okay(i int) bool {
	return e.ruleset().Sheet(e.partyIDs()[i]).Status == rules.StatusOkay
}

// firstOkay is the first member whose status is okay, or -1.
func (e *Engine) firstOkay() int {
	for i := range e.partySize() {
		if e.okay(i) {
			return i
		}
	}
	return -1
}

// spellID is the ruleset's id of por spell number n, "" outside the table.
func (e *Engine) spellID(n int) rules.SpellID {
	return e.fact(rules.NoChar, rules.SpellByNumber{N: n}).(rules.SpellID)
}

// scriptSpell is the ruleset's id of a script's spell number n (the game's).
func (e *Engine) scriptSpell(n int) rules.SpellID {
	return e.fact(rules.NoChar, rules.SpellByNumber{N: n, Game: true}).(rules.SpellID)
}

// UnknownItemType is a script's item number with no ruleset id (ids.txt).
type UnknownItemType struct{ N int }

func (u UnknownItemType) Error() string {
	return fmt.Sprintf("por: item type %#x has no dnd id (rules1e ids.txt)", u.N)
}

// itemID is the ruleset's id of por item type n (a script's number);
// UnknownItemType when ids.txt has none.
func (e *Engine) itemID(n int) (rules.ItemID, error) {
	v, err := e.ruleset().Fact(rules.NoChar, rules.ItemByNumber{N: n})
	if err != nil {
		return 0, UnknownItemType{n}
	}
	return v.(rules.ItemID), nil
}

// weapon is the item cID has readied as its weapon, core.NoItem none.
func (e *Engine) weapon(cID rules.CharID) rules.ItemHandle {
	if w := e.fact(cID, rules.WeaponFact{}).(rules.Handles); len(w) > 0 {
		return w[0]
	}
	return rules.NoItem
}

// loadChar loads the saved character named, core.NoChar when its files are gone.
func (e *Engine) loadChar(name string) rules.CharID {
	id, _ := e.LoadCharacter(name, false)
	return id
}

// mod is the modifier of c's ability a (0 strength .. 5 charisma).
func (e *Engine) mod(cID rules.CharID, a int) int {
	return int(e.fact(cID, rules.ModFact{Ability: rules.Abilities[a]}).(rules.Number))
}

// paladinHeals and paladinCures are whether c may lay on hands, cure disease now.
func (e *Engine) paladinHeals(cID rules.CharID) bool {
	return e.ruleset().Check(rules.PaladinCan{Who: cID}, nil).Yes
}
func (e *Engine) paladinCures(cID rules.CharID) bool {
	return e.ruleset().Check(rules.PaladinCan{Who: cID, Cure: true}, nil).Yes
}

// nameColour = 00BA:1816 displayName's colour: red when not okay,
// yellow for monsters, cyan otherwise. (00BA:08EF displayAC and 00BA:0973
// displayHP are the presenter's: the roster, Party.)
func (e *Engine) nameColour(id rules.CharID) int {
	switch {
	case !bool(e.fact(id, rules.FactActive).(rules.Yes)):
		return asset.EGALightRed
	case e.fact(id, rules.FactSide).(rules.Side) == rules.SideMonsters:
		return asset.EGAYellow
	}
	return asset.EGALightCyan
}

// rosterColour is the roster's colour for id's name when it is not the one
// highlighted: magenta ready for a level (extra), else nameColour.
func (e *Engine) rosterColour(id rules.CharID) int {
	if e.fact(id, rules.FactActive).(rules.Yes) && e.fact(id, rules.FactSide).(rules.Side) != rules.SideMonsters &&
		e.fact(id, rules.FactLevelReady).(rules.Yes) {
		return asset.EGALightMagenta
	}
	return e.nameColour(id)
}

// name is id's name.
func (e *Engine) name(id rules.CharID) string { return e.ruleset().Sheet(id).Name }

// partySide is the party's members that fight on its side (not monsters), in marching order.
func (e *Engine) partySide() rules.Party {
	var out rules.Party
	for _, id := range e.partyIDs() {
		if e.fact(id, rules.FactSide).(rules.Side) != rules.SideMonsters {
			out = append(out, id)
		}
	}
	return out
}
