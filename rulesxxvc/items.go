package rulesxxvc

import (
	"fmt"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// Carried is one item on a character's itemList: Slot is its id on the
// character, Readied TItem's C_Item.READIED flag.
type Carried struct {
	Handle  rules.ItemHandle // the actions' name for it, the ruleset's own (handOut); 0 until it holds it
	Slot    rules.ItemSlot
	Ref     rules.ItemRef
	Readied bool
}

// ArmorFlags is C_Armor's armorFlags, set by AdjustEquippedItem.
type ArmorFlags struct{ Body, Breathing, Fields, Goggles, ECM bool }

// The item kinds AdjustEquippedItem knows beyond the armour (charhlp.pas).
const (
	equipGoggles          xxvc.EquipSlot = 3
	equipBreathingMask    xxvc.EquipSlot = 4
	battleArmorFieldsType                = 24 // Battle armor w/fields
)

// sonicStunnerType is C_Item.SONIC_STUNNER_TYPE: its plus adds to the
// to-hit, never to the damage (its hit stuns).
const sonicStunnerType = 15

// The gdconsts.pas tables by strength or dexterity 1..22 (BR1 = BR2); index 0 unused.
var (
	dexArmorAdjustments  = [23]int8{0, -5, -5, -4, -3, -2, -1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 4, 4, 5, 5}
	strToHitAdjustments  = [23]int8{0, -5, -4, -3, -2, -2, -1, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 3, 4, 4}
	strDamageAdjustments = [23]int8{0, -4, -3, -2, -1, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 4, 6, 8, 10}
	weightAllowances     = [23]int{0, 2, 4, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 170, 220, 400, 600, 900, 1200}
)

// score is an ability as the tables index it, clamped to 1..22.
func score(c *Character, a rules.Ability) int { return min(max(c.Scores[a], 1), 22) }

// readied is RecalcStats' readiedItems[slot]: the last readied item of that
// slot in list order, with its template and type.
func (r *Rules) readied(c *Character, slot xxvc.EquipSlot) (Carried, xxvc.ItemTemplate, xxvc.ItemType, bool) {
	var (
		it    Carried
		t     xxvc.ItemTemplate
		ty    xxvc.ItemType
		found bool
	)
	for _, ci := range c.Items {
		if tt, tty, ok := r.itemType(ci); ok && ci.Readied && tty.Slot == slot {
			it, t, ty, found = ci, tt, tty, true
		}
	}
	return it, t, ty, found
}

func (r *Rules) itemType(ci Carried) (xxvc.ItemTemplate, xxvc.ItemType, bool) {
	t, ok := r.tables.Items[ci.Ref.ID]
	if !ok || t.Type < 0 || int(t.Type) >= len(r.tables.ItemTypes) {
		return t, xxvc.ItemType{}, false
	}
	return t, r.tables.ItemTypes[t.Type], true
}

// plus is the item's own plus over its template's.
func plus(ci Carried, t xxvc.ItemTemplate) int {
	if ci.Ref.Plus != 0 {
		return int(ci.Ref.Plus)
	}
	return int(t.Plus)
}

// RecalcStats is charhlp.pas RecalcStats with WeaponAdjust and
// AdjustEquippedItem: encumbrance (readied armour weighs nothing), the
// unarmed or readied weapon's to-hit and damage, the armour class from the
// readied armour, movement, the dodge bonus and the rear armour.
// A side that made its tactics roll this round hits at +1.
func (r *Rules) RecalcStats(c *Character) {
	c.Encumbrance = 0
	for _, ci := range c.Items {
		t, ty, ok := r.itemType(ci)
		if !ok || ty.Slot == xxvc.SlotArmor && ci.Readied {
			continue
		}
		w := int(t.Weight)
		if q := quantity(ci, t); q > 0 && !t.NotJoinable {
			w *= q
		}
		c.Encumbrance += w
	}
	c.Damage = c.Unarmed
	c.AttackBonus = c.BaseCombat
	if wi, wt, wty, ok := r.readied(c, xxvc.SlotWeapon); ok { // WeaponAdjust
		c.Damage[0] = rules.Dice{N: wty.Damage.N, Sides: wty.Damage.Sides, Bonus: wty.Damage.Bonus}
		if wty.WeaponClass == xxvc.Melee {
			c.AttackBonus += int(strToHitAdjustments[score(c, rules.Strength)])
			c.Damage[0].Bonus += int32(strDamageAdjustments[score(c, rules.Strength)])
		} else {
			c.AttackBonus += DexMissileAdjust(score(c, rules.Dexterity))
		}
		adj := c.Specialized(wt.Type) + plus(wi, wt)
		if wt.Type != sonicStunnerType {
			c.Damage[0].Bonus += int32(adj)
		}
		c.AttackBonus += adj
	} else {
		c.AttackBonus += int(strToHitAdjustments[score(c, rules.Strength)])
		c.Damage[0].Bonus += int32(strDamageAdjustments[score(c, rules.Strength)])
	}
	if c.Status == StFailedZeroG {
		c.AttackBonus -= 2
	}
	if r.tactics[side(c)] {
		c.AttackBonus++
	}
	c.ArmorClass[0], c.Armor = c.BaseArmorClass, ArmorFlags{}
	for _, ci := range c.Items { // AdjustEquippedItem
		t, ty, ok := r.itemType(ci)
		if !ok || !ci.Readied {
			continue
		}
		switch ty.Slot {
		case xxvc.SlotArmor:
			c.ArmorClass[0] += int(ty.ArmorClass) + plus(ci, t) // weaponClass - 50: xxvc.ItemType's ArmorClass
			c.Armor.Body, c.Armor.Breathing = true, true
			c.Armor.Fields = c.Armor.Fields || t.Type == battleArmorFieldsType
		case equipGoggles:
			c.Armor.Goggles = true
		case equipBreathingMask:
			c.Armor.Breathing = true
		case xxvc.SlotECM:
			c.Armor.ECM = true
		}
	}
	c.Movement = r.RecalcMovement(c) / 2
	if c.Status == StDodging {
		if c.Level < 4 {
			c.ArmorClass[0] += 2
		} else {
			c.ArmorClass[0] += c.Level / 2
		}
	}
	c.ArmorClass[1] = c.ArmorClass[0] - 2
	c.ArmorClass[0] += int(dexArmorAdjustments[score(c, rules.Dexterity)])
}

func quantity(ci Carried, _ xxvc.ItemTemplate) int { return int(ci.Ref.Count) }

// SpecializationSlots is TCharacter.weaponSpecializations' size; 0 is an
// empty slot (viewchar.pas training fills the first 0).
const SpecializationSlots = 10

// Specialized is CharHlp.SpecializationLevel: how many of the
// weaponSpecializations name the item type. As in the reference, item type
// 0 (ITEMS row 0: slot WEAPON, no dice) counts the empty slots.
func (c *Character) Specialized(itemType int32) int {
	n := 0
	for _, s := range c.Specializations {
		if s == itemType {
			n++
		}
	}
	return n
}

// RecalcMovement is charhlp.pas RecalcMovement: the base movement, halved
// when the encumbrance is over half the weight allowance, plus the zone's
// movementBonus for the team (core.FightTerms.PartyMove, kept from
// BeginFight to EndFight: postcbt.pas clears it), doubled.
func (r *Rules) RecalcMovement(c *Character) int {
	m := c.BaseMovement
	if weightAllowances[score(c, rules.Strength)]/2 < c.Encumbrance {
		m /= 2
	}
	if c.team() {
		m += r.moveBonus
	}
	return m * 2
}

// ready is items.pas Ready/Unready's rule part: the item's READIED flag, and
// for a weapon or armour the slot's other readied item laid down, then
// RecalcStats. ponytail: heavy items ending the turn and the class/strength
// checks of readying are not ported.
func (r *Rules) ready(who rules.CharID, h rules.ItemHandle, on bool) (rules.Outcome, error) {
	c := r.chars[who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("Ready: no character %d", who)
	}
	i := heldAt(c, h)
	if i < 0 {
		return rules.Outcome{}, fmt.Errorf("Ready: %d carries no item %d", who, h)
	}
	if _, ty, ok := r.itemType(c.Items[i]); ok && on && (ty.Slot == xxvc.SlotWeapon || ty.Slot == xxvc.SlotArmor) {
		for j := range c.Items {
			if _, oty, ok := r.itemType(c.Items[j]); ok && oty.Slot == ty.Slot {
				c.Items[j].Readied = false
			}
		}
	}
	c.Items[i].Readied = on
	r.RecalcStats(c)
	return rules.Outcome{}, nil
}

// dropItem is viewchar.pas's gear menu 'D' after YesNoMenu 'Drop It?': CheckItemAvail refuses a readied item ("Must be
// unreadied"), else CharHlp.RemoveItem takes it off the list and RecalcStats follows.
func (r *Rules) dropItem(who rules.CharID, h rules.ItemHandle) (rules.Outcome, error) {
	c := r.chars[who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("DropItem: no character %d", who)
	}
	i := heldAt(c, h)
	if i < 0 {
		return rules.Outcome{}, fmt.Errorf("DropItem: %d carries no item %d", who, h)
	}
	if c.Items[i].Readied {
		return rules.Outcome{}, rules.ErrNotNow
	}
	c.Items = slices.Delete(c.Items, i, i+1)
	r.RecalcStats(c)
	return rules.Outcome{}, nil
}
