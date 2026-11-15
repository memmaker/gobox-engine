package rulesxxvc

import (
	"image"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// SelectAreaTarget by hand: a hostile at (0,0) with an explosive grenade
// (range 5: outer 6, inner 5) against party members at (5,0) and (6,0).
// The blasts covering both score 2 on x = 5, rows -1..1; row -1 is off the
// board, so the best is (5,0) or, on the 25% tie roll, (5,1). The value is
// 40 * 2 = 80.
func TestSelectAreaTarget(t *testing.T) {
	r := New()
	tt := xxvc.ItemTemplate{Type: explosiveGrenadeType}
	ty := xxvc.ItemType{Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Grenade, Range: 5}
	a := r.Add(member("a", 10, 0, 0, 0))
	b := r.Add(member("b", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5})
	f := rules.Fight{Party: rules.Party{a, b}, Board: rules.Board{Size: image.Pt(20, 10), Units: []rules.BoardUnit{
		{ID: a, At: rules.BoardPos{X: 5}}, {ID: b, At: rules.BoardPos{X: 6}}, {ID: m}}}}
	rng := rules.RNG{State: 1}
	got, ok := r.selectAreaTarget(f, m, r.chars[m], &combatState{target: a}, tt, ty, &rng)
	if !ok || got.weight != 2 || got.at != (rules.BoardPos{X: 5}) && got.at != (rules.BoardPos{X: 5, Y: 1}) {
		t.Fatalf("%+v %v", got, ok)
	}
	if v := areaChoice(got, tt.Type, &rng); v != 80 {
		t.Fatal(v)
	}
	if _, ok := r.selectAreaTarget(f, m, r.chars[m], &combatState{}, xxvc.ItemTemplate{}, xxvc.ItemType{}, &rng); ok {
		t.Fatal("a melee weapon scored as an area weapon")
	}
}
