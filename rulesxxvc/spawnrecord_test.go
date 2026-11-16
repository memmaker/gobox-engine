package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// EclLoadMonster keeps what the monster's record holds: isCompanion and the npc byte (CheckMorale tests them), the
// items it has READIED (RecalcStats counts them from the first round); the clones of a group get the item list
// reversed (the copy prepends each item), the template its own order.
func TestSpawnRecordAndClones(t *testing.T) {
	tb := tables()
	tb.Items[8] = xxvc.ItemTemplate{Type: 0, Words: [3]int32{1, 2, 0}, Quantity: 1}
	tb.Monsters[3] = xxvc.Monster{Name: "RAM trooper", HP: 9, Hostile: true, Companion: true, NonPlayer: true, Morale: 40,
		Items: []xxvc.MonsterItem{{ItemRef: rules.ItemRef{ID: 7}, Readied: true}, {ItemRef: rules.ItemRef{ID: 8}}}}
	r := New()
	r.Begin(rules.Campaign{Tables: tb}, rules.Options{})
	order := func(id rules.CharID) []rules.ItemID {
		var out []rules.ItemID
		for _, it := range r.chars[id].Items {
			out = append(out, it.Ref.ID)
		}
		return out
	}
	r.Resolve(nil, rules.NewGroup{}, nil)
	first, _ := r.Spawn(3, nil)
	clone, _ := r.Spawn(3, nil)
	c := r.chars[first]
	if !c.Companion || !c.NonPlayer || c.Morale != 40 || !c.Items[0].Readied || c.Items[1].Readied {
		t.Errorf("%+v", c)
	}
	if !slices.Equal(order(first), []rules.ItemID{7, 8}) || !slices.Equal(order(clone), []rules.ItemID{8, 7}) {
		t.Errorf("template %v, clone %v", order(first), order(clone))
	}
	r.Resolve(nil, rules.NewGroup{}, nil)
	again, _ := r.Spawn(3, nil)
	if !slices.Equal(order(again), []rules.ItemID{7, 8}) {
		t.Errorf("the next group's template %v", order(again))
	}
}

// combat.pas CreatureAction clears attacksDrawn as the creature's turn begins: an attack drawn on it before its own
// turn does not make a later one a rear attack.
func TestBeginTurnClearsAttacksDrawn(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	rng := rules.RNG{State: 1}
	f := rules.Fight{Party: rules.Party{a}}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[a].attacksDrawn = 2
	r.BeginTurn(f, a, &rng)
	if r.fight[a].attacksDrawn != 0 {
		t.Errorf("attacksDrawn %d", r.fight[a].attacksDrawn)
	}
}
