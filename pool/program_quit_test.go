package pool

import (
	"gobox/ecl/script"
	"gobox/game"
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
	if !slices.Contains(e.events, play.Event(play.SaveOffered{})) || !slices.Contains(e.events, play.Event(play.Ended{})) || e.story != nil {
		t.Fatalf("SaveQuit: events %v, finale started %v", e.events, e.story != nil)
	}
}

// TestProgramSaveQuitEnding: Program 1 offers the save screen, then plays the end
// sequence (Continue ends it) or, with none, ends the run at once; never POR's finale.
func TestProgramSaveQuitEnding(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = script.V3
	e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: 1}}}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.title.seqs = []game.Sequence{{Name: seqEnding}}
	e.events = nil
	e.opProgram()
	if e.story != nil || e.sequence != seqEnding || slices.Contains(e.events, play.Event(play.Ended{})) {
		t.Fatalf("SaveQuit with an end: story %v, sequence %q, events %v: want the sequence up, the run on", e.story != nil, e.sequence, e.events)
	}
	o, err := e.Do(play.Continue{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(o.Events, play.Event(play.Ended{})) {
		t.Fatalf("Continue on the ending: events %v, want Ended", o.Events)
	}
}
