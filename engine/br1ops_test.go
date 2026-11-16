package engine

import (
	"slices"
	"testing"

	"gobox/ecl"
)

// br1Blocks are BR1's script blocks (ECL1..6).
var br1Blocks = []int{16, 17, 18, 32, 33, 34, 35, 48, 49, 50, 52, 64, 65, 66, 67, 80, 81, 82, 83, 94, 95, 96, 97, 98, 99}

// br1OpsNotPorted are the commands BR1's scripts use that the engine has no handler for yet: a script reaching
// one panics. The list must match what the scripts and the engine say: a command ported leaves it (REFACTOR.md B1f).
var br1OpsNotPorted = []string{"Damage6", "Duel", "HasItem", "NPCJoin", "Op64", "Op72", "Shop", "Skill", "SkillAll", "SpaceFight", "WallDistance"}

// TestBR1OpsPorted: every command BR1's scripts use has a handler, but the listed ones.
func TestBR1OpsPorted(t *testing.T) {
	e := br1Engine(t)
	missing := map[string]bool{}
	for _, block := range br1Blocks {
		s, err := e.parseScriptFile(block)
		if err != nil {
			t.Fatal(block, err)
		}
		for _, in := range s.Code {
			if e.vm.Ops[e.vm.Raw[in.Op]] == nil {
				missing[ecl.OpName(in.Op)] = true
			}
		}
	}
	var got []string
	for name := range missing {
		got = append(got, name)
	}
	slices.Sort(got)
	if !slices.Equal(got, br1OpsNotPorted) {
		t.Errorf("commands without a handler %v, listed %v", got, br1OpsNotPorted)
	}
}
