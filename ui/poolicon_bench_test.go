package ui

import "testing"

// A party member's combat icon, as BattleOf builds it for every View of a fight.
func BenchmarkPoolIcon(b *testing.B) {
	for range b.N {
		if _, ok := poolIcon("combat_head_000", "combat_body_000", nil, 1, false); !ok {
			b.Fatal("no icon")
		}
	}
}
