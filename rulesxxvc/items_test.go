package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// RecalcStats with WeaponAdjust: strength 18 gives a melee weapon +2/+2,
// its plus 1 and one specialization +2 more to both; dexterity 16 gives the
// front armour +2, the rear is the base less 2.
func TestRecalcStats(t *testing.T) {
	r := New()
	r.tables = xxvc.Tables{ItemTypes: []xxvc.ItemType{{}, {Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Melee, Damage: rules.Dice{N: 1, Sides: 8}}},
		Items: map[rules.ItemID]xxvc.ItemTemplate{3: {Type: 1, Plus: 1, Weight: 5}}}
	a := r.Add(Character{Scores: map[rules.Ability]int{rules.Strength: 18, rules.Dexterity: 16}, BaseCombat: 5,
		BaseArmorClass: 10, BaseMovement: 12, Unarmed: [2]rules.Dice{{N: 1, Sides: 2}}, Specializations: [SpecializationSlots]int32{1},
		Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 3}}}})
	if _, err := r.Resolve(rules.Camping{}, rules.Ready{Who: a, Item: r.chars[a].Items[0].Handle}, nil); err != nil {
		t.Fatal(err)
	}
	c := r.chars[a]
	if c.AttackBonus != 9 || c.Damage[0] != (rules.Dice{N: 1, Sides: 8, Bonus: 4}) || c.ArmorClass != [2]int{12, 8} ||
		c.Movement != 12 || c.Encumbrance != 5 {
		t.Errorf("%+v", c)
	}
}
