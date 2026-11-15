package rulesxxvc

import (
	"fmt"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// SpawnHP is eclvm.pas EclLoadMonster's difficulty scaling of a hostile's
// hit points: (difficulty + 2) / 4 of them, at least 1.
func SpawnHP(hp int, d Difficulty) int {
	if hp = (int(d) + 2) * hp / 4; hp == 0 {
		return 1
	}
	return hp
}

// Spawn is EclLoadMonster's one creature: the record InitCreature loads, a
// hostile's HP scaled by the difficulty. The script's count is the engine's
// loop; nothing is rolled.
func (r *Rules) Spawn(id rules.MonsterID, _ *rules.RNG) (rules.CharID, error) {
	m, ok := r.tables.Monsters[id]
	if !ok {
		return rules.NoChar, fmt.Errorf("Spawn: no monster %d", id)
	}
	c := Character{Name: m.Name, Race: m.Race, Career: m.Career, Level: int(m.Level),
		Scores: map[rules.Ability]int{}, HP: int(m.HP), MaxHP: int(m.HP), BaseCombat: int(m.BaseCombat),
		Enabled: true, Status: StOkay, Side: rules.SideAuto, XPAward: m.XPAward, Coins: m.Credits,
		AttackBonus: int(m.AttackBonus), ArmorClass: [2]int{int(m.ArmorClass[0]), int(m.ArmorClass[1])},
		Damage: m.Damage, BaseScale: int(m.BaseScale), Scale: int(m.Scale), Movement: int(m.Movement),
		Unarmed: m.Damage, BaseArmorClass: int(m.ArmorClass[0]), BaseMovement: int(m.Movement)}
	for i, ref := range m.Items {
		if ref.Count == 0 { // a stat block's item comes with the template's quantity
			ref.Count = r.tables.Items[ref.ID].Quantity
		}
		c.Items = append(c.Items, Carried{Slot: rules.ItemSlot(i + 1), Ref: ref})
	}
	for a, v := range m.Scores {
		c.Scores[a] = int(v)
	}
	for s, v := range m.Skills {
		if n, ok := skillNo(s); ok {
			c.Skills.Set(n, int8(v))
		}
	}
	for i, v := range m.Saves {
		c.Saves[i] = int(v)
	}
	if m.Hostile {
		c.Side = rules.SideMonsters
		c.HP = SpawnHP(c.HP, difficulty(r.options.Difficulty))
		c.MaxHP = c.HP
	}
	// A monster's credits are loot, never the party's purse.
	cid := r.next
	r.next++
	r.chars[cid] = &c
	r.handOut(&c)
	return cid, nil
}

// Item is a template's numbers (eclvm.pas: an item copies its template) with
// its type's (itemTypes); the placed item's own plus, charges and count win.
func (r *Rules) Item(ref rules.ItemRef) (rules.Item, error) {
	t, ok := r.tables.Items[ref.ID]
	if !ok || int(t.Type) >= len(r.tables.ItemTypes) || t.Type < 0 {
		return rules.Item{}, fmt.Errorf("Item: no template %d", ref.ID)
	}
	ty := r.tables.ItemTypes[t.Type]
	it := rules.Item{ID: ref.ID, Range: ty.Range, Damage: ty.Damage, Plus: t.Plus, Plus2: t.SaveBonus,
		Charges: ref.Charges, Count: t.Quantity, Price: t.Value, Cursed: ref.Cursed, Identified: true}
	switch ty.Slot {
	case xxvc.SlotWeapon:
		it.Kind = rules.KindWeapon
	case xxvc.SlotArmor:
		it.Kind, it.AC = rules.KindArmor, ty.ArmorClass
	default:
		it.Kind = rules.KindMisc
	}
	if ref.Plus != 0 {
		it.Plus = ref.Plus
	}
	if ref.Count != 0 {
		it.Count = ref.Count
	}
	for _, w := range t.Words {
		if w > 0 && int(w) < len(r.tables.ItemWords) {
			it.NameWords = append(it.NameWords, r.tables.ItemWords[w])
			it.Hidden = append(it.Hidden, false)
		}
	}
	return it, nil
}
