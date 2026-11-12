package pool

import (
	"gobox/ecl/script"
	"gobox/game/play"
	"slices"
	"testing"
)

// TestProgramQuit: pod's Program 2 tells the front to end the run
// (play.Ended); Program 1 offers the save screen, then the ending.
func TestProgramQuit(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = script.V3
	program := func(n int) {
		e.events = nil
		e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: n}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opProgram()
	}
	program(2)
	if !slices.Contains(e.events, play.Event(play.Ended{})) {
		t.Fatalf("Quit: events %v", e.events)
	}
	program(1)
	if !slices.Contains(e.events, play.Event(play.SaveOffered{})) || e.story == nil {
		t.Fatalf("SaveQuit: events %v, ending started %v", e.events, e.story != nil)
	}
}
