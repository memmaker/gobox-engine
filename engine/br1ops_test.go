package engine

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/ecl"
)

// br1Blocks are BR1's script blocks (ECL1..6).
var br1Blocks = []int{16, 17, 18, 32, 33, 34, 35, 48, 49, 50, 52, 64, 65, 66, 67, 80, 81, 82, 83, 94, 95, 96, 97, 98, 99}

// br1OpsNotPorted are the commands BR1's scripts use that the engine has no handler for yet: a script reaching
// one panics. The list must match what the scripts and the engine say: a command ported leaves it (REFACTOR.md B1f).
var br1OpsNotPorted = []string{}

// brOpsMissing are the commands the game's scripts (every block of its folder) use that have no handler.
func brOpsMissing(t *testing.T, e *Engine) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(e.scriptDir(), "area*.txt"))
	if err != nil || len(files) == 0 {
		t.Fatal("no scripts", err)
	}
	missing := map[string]bool{}
	for _, f := range files {
		block, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f), "area"), ".txt"))
		if err != nil {
			t.Fatal(f, err)
		}
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
	return got
}

// TestBR1OpsPorted: every command BR1's scripts use has a handler, but the listed ones.
func TestBR1OpsPorted(t *testing.T) {
	if got := brOpsMissing(t, brEngine(t, gameBR1)); !slices.Equal(got, br1OpsNotPorted) {
		t.Errorf("commands without a handler %v, listed %v", got, br1OpsNotPorted)
	}
}

// br2OpsNotPorted are BR2's commands without a handler.
var br2OpsNotPorted = []string{}

// TestBR2OpsPorted: the same for BR2.
func TestBR2OpsPorted(t *testing.T) {
	if got := brOpsMissing(t, brEngine(t, gameBR2)); !slices.Equal(got, br2OpsNotPorted) {
		t.Errorf("commands without a handler %v, listed %v", got, br2OpsNotPorted)
	}
}
