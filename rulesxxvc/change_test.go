package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// TestSetDifficulty: the Change menu's Level (explore.pas CampDifficultyMenu: zone0 difficulty := the level picked) changes
// the hit points a monster is made with (eclvm.pas:358, SpawnHP) and is saved with the game.
func TestSetDifficulty(t *testing.T) {
	r := New()
	if _, err := r.Resolve(nil, rules.SetDifficulty{Level: rules.DifficultyHardest}, nil); err != nil {
		t.Fatal(err)
	}
	if got := difficulty(r.options.Difficulty); got != Ace {
		t.Errorf("difficulty %d, want Ace", got)
	}
	b, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	l := New()
	if err := l.Load(b); err != nil {
		t.Fatal(err)
	}
	if l.options.Difficulty != rules.DifficultyHardest {
		t.Errorf("a loaded game has level %d", l.options.Difficulty)
	}
}

// TestDismissLeavesTheParty: a member dismissed is out of the marching order too (Remove, Purge and Drop use it).
func TestDismissLeavesTheParty(t *testing.T) {
	r := New()
	a, b := r.Add(Character{}), r.Add(Character{})
	r.party, r.acting = []rules.CharID{a, b}, 1
	r.Dismiss(b)
	if len(r.party) != 1 || r.party[0] != a || r.acting != 0 {
		t.Errorf("party %v acting %d", r.party, r.acting)
	}
	r.Dismiss(a)
	if len(r.party) != 0 || r.acting != 0 {
		t.Errorf("party %v acting %d", r.party, r.acting)
	}
}
