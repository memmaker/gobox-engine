package engine

import (
	"gobox/rules"
	"testing"

	"gobox/char"
)

type fixedDie int

func (d fixedDie) Dice(n, sides int) int { return int(d) }

type saveStats struct {
	char.Stats
	target int
}

func (s saveStats) SaveBonus() int { return 0 }
func (s saveStats) Save(int) int   { return s.target }

// POR's fix, now the one path: a penalty below 0 does not wrap to ~250 and
// save (roll 10 - 30 against a target of 12 fails).
func TestSavingThrowSigned(t *testing.T) {
	var rv rules.Roll
	if SavingThrow(&rv, fixedDie(10), saveStats{target: 12}, 4, -30, Hooks{}) {
		t.Error("a negative total saved")
	}
	if !SavingThrow(&rv, fixedDie(10), saveStats{target: 12}, 4, 5, Hooks{}) {
		t.Error("10+5 should reach 12")
	}
}
