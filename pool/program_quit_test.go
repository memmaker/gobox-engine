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
	if !slices.Contains(e.events, play.Event(play.SaveOffered{})) || !slices.Contains(e.events, play.Event(play.Ended{})) {
		t.Fatalf("SaveQuit: events %v", e.events)
	}
}

// TestProgramSaveQuitEnding: Program 1 offers the save screen, then plays the end
// sequence (Continue ends it) or, with none, ends the run at once; never a finale of its own.
func TestProgramSaveQuitEnding(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = script.V3
	e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: 1}}}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.title.seqs = []game.Sequence{{Name: seqEnding}}
	e.events = nil
	e.opProgram()
	if e.sequence != seqEnding || slices.Contains(e.events, play.Event(play.Ended{})) {
		t.Fatalf("SaveQuit with an end: sequence %q, events %v: want the sequence up, the run on", e.sequence, e.events)
	}
	o, err := e.Do(play.Continue{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(o.Events, play.Event(play.Ended{})) {
		t.Fatalf("Continue on the ending: events %v, want Ended", o.Events)
	}
}

// TestProgramEnding: PROGRAM's ending (POR 8) plays the manifest's End and then the script goes
// on: no End, nothing waits; an End of pictures is a scene Continue ends, the run goes on.
func TestProgramEnding(t *testing.T) {
	e := quickStarted(t)
	program := func() {
		e.events = nil
		e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: 8}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opProgram()
	}
	program()
	if e.ext.ending || e.waiting(PendSequence) {
		t.Fatal("an ending with no End waits")
	}
	e.title.end = &play.Scene{Kind: play.SceneTitle, Intro: &game.Intro{Full: true}}
	program()
	if sc, ok := e.State().(play.Scene); !ok || sc.Kind != play.SceneTitle || !e.waiting(PendSequence) {
		t.Fatalf("the end is up as %T, waiting %v", e.State(), e.waiting(PendSequence))
	}
	o, err := e.Do(play.Continue{})
	if err != nil || e.ext.ending || e.waiting(PendSequence) || slices.Contains(o.Events, play.Event(play.Ended{})) {
		t.Fatalf("Continue: %v, ending %v, events %v: want the script on", err, e.ext.ending, o.Events)
	}
}
