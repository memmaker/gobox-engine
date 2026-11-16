package engine

import (
	"gobox/ecl"
	"testing"
)

// TestProgramPartyDefeated: cotab's Program 3 is partyDefeated := True (eclvm.pas EclProgram),
// the flag a lost fight sets: the script stops (run.quit), it is no run-end event.
func TestProgramPartyDefeated(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = ecl.V2
	e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: []ecl.Arg{{N: 3}}}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.run.quit = false
	e.opProgram()
	if !e.run.quit {
		t.Fatal("Program 3 did not defeat the party")
	}
}
