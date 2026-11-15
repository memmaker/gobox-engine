package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// EquipWeaponAndArmor by hand from combatai.pas: the d10 sword (10) beats
// the d4 knife (4) and the unarmed 1d2 (2); a laser (2d6 = 12) against a
// laser-immune target shifts to 1; the better armour (class 5 + 1) wins; a
// heavy sword ends the turn.
func TestEquipWeaponAndArmor(t *testing.T) {
	r := New()
	r.tables = xxvc.Tables{
		ItemTypes: []xxvc.ItemType{
			{},
			{Slot: xxvc.SlotWeapon, Damage: rules.Dice{N: 1, Sides: 4}},
			{Slot: xxvc.SlotWeapon, Damage: rules.Dice{N: 1, Sides: 10}},
			{Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Laser, Damage: rules.Dice{N: 2, Sides: 6}},
			{Slot: xxvc.SlotArmor, ArmorClass: 5},
			{Slot: xxvc.SlotArmor, ArmorClass: 4},
			{Slot: xxvc.SlotWeapon, Heavy: true, Damage: rules.Dice{N: 1, Sides: 10}},
		},
		Items: map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 1}, 2: {Type: 2}, 3: {Type: 3, Quantity: 5}, 4: {Type: 4, Plus: 1}, 5: {Type: 5, Plus: 1}, 6: {Type: 6}},
	}
	c := &Character{Enabled: true, Unarmed: [2]rules.Dice{{N: 1, Sides: 2}}, Scores: map[rules.Ability]int{rules.Strength: 10, rules.Dexterity: 10},
		Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1}, Readied: true}, {Slot: 2, Ref: rules.ItemRef{ID: 2}}, {Slot: 3, Ref: rules.ItemRef{ID: 3, Count: 5}},
			{Slot: 4, Ref: rules.ItemRef{ID: 4}}, {Slot: 5, Ref: rules.ItemRef{ID: 5}, Readied: true}}}
	tgt := r.Add(Character{Enabled: true, Effects: []Effect{{ID: EffImmuneToLaser}}})
	s := &combatState{target: tgt}
	if r.equipWeaponAndArmor(rules.Fight{}, rules.NoChar, c, s, true, nil) || r.readiedIndex(c, xxvc.SlotWeapon) != 1 || r.readiedIndex(c, xxvc.SlotArmor) != 3 {
		t.Fatalf("weapon %d armour %d", r.readiedIndex(c, xxvc.SlotWeapon), r.readiedIndex(c, xxvc.SlotArmor))
	}
	r.chars[tgt].Effects = nil // the laser's 12 now beats the sword
	if r.equipWeaponAndArmor(rules.Fight{}, rules.NoChar, c, s, true, nil); r.readiedIndex(c, xxvc.SlotWeapon) != 2 {
		t.Fatalf("laser not readied: %d", r.readiedIndex(c, xxvc.SlotWeapon))
	}
	c.Items = append(c.Items, Carried{Slot: 6, Ref: rules.ItemRef{ID: 6, Plus: 1}}) // 10 + 3 = 13
	if !r.equipWeaponAndArmor(rules.Fight{}, rules.NoChar, c, s, true, nil) {
		t.Fatal("heavy weapon readied without ending the turn")
	}
}

// RecalcAttacksLeft by hand from cbtacts.pas: unarmed baseScale 4 in round 1
// gives (4+1)>>1 = 2, DIZZY halves the rate first ((2+1)>>1 = 1); once it
// has attacked a larger rate never raises the attacks left.
func TestRecalcAttacksLeft(t *testing.T) {
	r := New()
	c := &Character{Enabled: true, BaseScale: 4}
	s := &combatState{}
	f := rules.Fight{Round: 1}
	if r.recalcAttacksLeft(f, c, s); s.attacks[0] != 2 {
		t.Fatal(s.attacks[0])
	}
	c.Effects = []Effect{{ID: EffDizzy}}
	if r.recalcAttacksLeft(f, c, s); s.attacks[0] != 1 {
		t.Fatal(s.attacks[0])
	}
	c.Effects, s.hasAttacked = nil, true
	if r.recalcAttacksLeft(f, c, s); s.attacks[0] != 1 {
		t.Fatal("rose after attacking:", s.attacks[0])
	}
}
