package rules1e

import "testing"

// TestWhoConItemSaveBonus: a readied COTAB constitution item gives who()
// the save bonus power (a human gets the dwarf's bonus); unreadied it does not.
func TestWhoConItemSaveBonus(t *testing.T) {
	it := &Item{Readied: 1}
	it.Spells[2] = ItemPower(itemPowerConstitution)
	c := &Character{Items: []*Item{it}}
	if !who(c).Powers.SaveBonus {
		t.Fatal("readied constitution item: no save bonus")
	}
	it.Readied = 0
	if who(c).Powers.SaveBonus {
		t.Fatal("unreadied constitution item gives a save bonus")
	}
}
