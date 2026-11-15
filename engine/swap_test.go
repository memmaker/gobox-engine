package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"testing"
)

// Readying an item whose slot is taken swaps the two.
func TestSwapReady(t *testing.T) {
	e := &Engine{}
	e.state().Items = make([]rules1e.ItemBase, rules1e.ItemTypes)
	e.initVM()
	e.state().Items[rules1e.ItemLongSword] = rules1e.ItemBase{Hands: 1, Classes: 0xFF}
	e.state().Items[rules1e.ItemDagger] = rules1e.ItemBase{Hands: 1, Classes: 0xFF}
	c := &rules1e.Character{}
	c.ClassMask = 1
	e.state().Party = []*rules1e.Character{c}
	sword := &rules1e.Item{Type: rules1e.ItemLongSword, Readied: 1}
	c.Items = []*rules1e.Item{sword}
	e.state().ApplyItems(c)
	e.state().AddItemCopy(c, &rules1e.Item{Type: rules1e.ItemDagger})
	d := e.state().Pool[0]
	e.state().SwapReady(c, d)
	e.state().ApplyItems(c)
	if d.Readied == 0 || sword.Readied != 0 || c.Slots[rules1e.SlotWeapon] != d {
		t.Fatalf("dagger %d sword %d", d.Readied, sword.Readied)
	}
}

// Bracers take the armour's slot: readying them swaps the armour off, and
// an older save with both readied keeps the armour.
func TestBracersSwapArmor(t *testing.T) {
	e := &Engine{}
	e.state().Items = make([]rules1e.ItemBase, rules1e.ItemTypes)
	e.initVM()
	e.state().Items[rules1e.ItemPlateMail] = rules1e.ItemBase{Slot: rules1e.SlotArmor, Protection: rules1e.AcGiven | (rules1e.AcTen + 7), Classes: 0xFF}
	e.state().Items[rules1e.ItemBracers] = rules1e.ItemBase{Slot: byte(rules1e.ItemBracers), Protection: rules1e.AcGiven | rules1e.AcTen, Classes: 0xFF}
	c := &rules1e.Character{}
	c.ClassMask = 1
	e.state().Party = []*rules1e.Character{c}
	plate := &rules1e.Item{Type: rules1e.ItemPlateMail, Readied: 1}
	c.Items = []*rules1e.Item{plate}
	e.state().ApplyItems(c)
	e.state().AddItemCopy(c, &rules1e.Item{Type: rules1e.ItemBracers, Plus: 4})
	b := e.state().Pool[0]
	if e.state().DollSlot(c, b) != rules.DollArmor || e.state().InTheWay(c, b) != plate {
		t.Fatalf("bracers: doll slot %d, in the way %v", e.state().DollSlot(c, b), e.state().InTheWay(c, b))
	}
	e.state().SwapReady(c, b)
	e.state().ApplyItems(c)
	if b.Readied == 0 || plate.Readied != 0 || c.Slots[rules1e.SlotArmor] != b {
		t.Fatalf("bracers %d plate %d", b.Readied, plate.Readied)
	}
	plate.Readied = 1 // as an older save: both readied, mended by its import
	c.Items = []*rules1e.Item{b, plate}
	e.state().ImportCharacter(c)
	e.state().ApplyItems(c)
	if c.Slots[rules1e.SlotArmor] != plate || b.Readied != 0 {
		t.Fatalf("both readied: slot holds %v, bracers readied %d", c.Slots[rules1e.SlotArmor], b.Readied)
	}
}

// Readying a two-handed weapon takes the weapon and the shield off.
func TestSwapReadyTwoHanded(t *testing.T) {
	e := &Engine{}
	e.state().Items = make([]rules1e.ItemBase, rules1e.ItemTypes)
	e.initVM()
	e.state().Items[rules1e.ItemLongSword] = rules1e.ItemBase{Hands: 1, Classes: 0xFF}
	e.state().Items[rules1e.ItemShield] = rules1e.ItemBase{Slot: rules1e.SlotShield, Hands: 1, Classes: 0xFF}
	e.state().Items[rules1e.ItemTwoHandedSword] = rules1e.ItemBase{Hands: 2, Classes: 0xFF}
	c := &rules1e.Character{}
	c.ClassMask = 1
	e.state().Party = []*rules1e.Character{c}
	sword := &rules1e.Item{Type: rules1e.ItemLongSword, Readied: 1}
	shield := &rules1e.Item{Type: rules1e.ItemShield, Readied: 1}
	c.Items = []*rules1e.Item{sword, shield}
	e.state().ApplyItems(c)
	e.state().AddItemCopy(c, &rules1e.Item{Type: rules1e.ItemTwoHandedSword})
	th := e.state().Pool[0]
	e.state().SwapReady(c, th)
	e.state().ApplyItems(c)
	if th.Readied == 0 || sword.Readied != 0 || shield.Readied != 0 {
		t.Fatalf("two-handed %d sword %d shield %d", th.Readied, sword.Readied, shield.Readied)
	}
}
