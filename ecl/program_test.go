package ecl

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

// TestProgramTables pins each dialect's PROGRAM table to its Pascal (eclvm.pas EclProgram, the case's arms) and to
// the codes its scripts pass: a literal operand the table does not list is a silent nothing, the bug class.
func TestProgramTables(t *testing.T) {
	arms := map[string]struct {
		v    *Version
		want map[byte]ProgramFn // the arms (cotab 8's "play the ending, heal, quit" is ProgramEnding)
		none []byte             // listed codes whose case has no arm
	}{
		"por":   {V1, map[byte]ProgramFn{0: ProgramPartyMenu, 8: ProgramEnding, 9: ProgramCamp}, nil},
		"cotab": {V2, map[byte]ProgramFn{0: ProgramPartyMenu, 3: ProgramPartyDefeated, 8: ProgramEnding, 9: ProgramCamp}, nil},
		"pod":   {V3, map[byte]ProgramFn{0: ProgramPartyMenu, 1: ProgramSaveQuit, 2: ProgramQuit, 9: ProgramCamp}, []byte{3, 4, 5, 6, 7, 8, 10, 50, podDebugProgramMax}},
		"br1":   {V4, map[byte]ProgramFn{0: ProgramPartyMenu, 2: ProgramQuit, 9: ProgramRest}, []byte{3}},
	}
	for game, c := range arms {
		for code, fn := range c.want {
			if got, ok := c.v.Programs[code]; !ok || got != fn {
				t.Errorf("%s: program %d is %v (listed %v), want %v", game, code, got, ok, fn)
			}
		}
		for _, code := range c.none {
			if got, ok := c.v.Programs[code]; !ok || got != ProgramNone {
				t.Errorf("%s: program %d is %v (listed %v), want ProgramNone listed", game, code, got, ok)
			}
		}
	}
	if _, ok := V3.Programs[podDebugProgramMax+1]; ok {
		t.Error("pod: a code the debug menu cannot type is listed")
	}
	// sotsb shares V2: its Pascal has arms 0 and 9 only; its own codes are adventure data (importer GameTables.Programs: 3 and 8 do nothing)
	call := regexp.MustCompile(`\bProgram\((\d+)\)`)
	for _, g := range []struct {
		name string
		v    *Version
	}{{"por", V1}, {"cotab", V2}, {"sotsb", V2}, {"pod", V3}, {"br1", V4}, {"br2", V4}} {
		files, _ := filepath.Glob(filepath.Join("..", "data", g.name, "scripts", "area*.txt"))
		if len(files) == 0 {
			t.Logf("%s: scripts not imported, skipped", g.name)
			continue
		}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range call.FindAllStringSubmatch(string(b), -1) {
				n, _ := strconv.Atoi(m[1])
				if _, ok := g.v.Programs[byte(n)]; !ok {
					t.Errorf("%s: %s passes Program(%d), not in the dialect's table", g.name, filepath.Base(f), n)
				}
			}
		}
	}
}
