package engine

import (
	"gobox/adventure"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"slices"
	"testing"
)

// TestProgramQuit: pod's Program 2 tells the front to end the run
// (game.Ended); Program 1 offers the save screen, then the ending.
func TestProgramQuit(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = ecl.V3
	program := func(n int) {
		e.events = nil
		e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: []ecl.Arg{{N: n}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opProgram()
	}
	program(2)
	if !slices.Contains(e.events, game.Event(game.Ended{})) {
		t.Fatalf("Quit: events %v", e.events)
	}
	program(1)
	if !slices.Contains(e.events, game.Event(game.SaveOffered{})) || !slices.Contains(e.events, game.Event(game.Ended{})) {
		t.Fatalf("SaveQuit: events %v", e.events)
	}
}

// TestProgramSaveQuitEnding: Program 1 offers the save screen, then plays the end
// sequence (Continue ends it) or, with none, ends the run at once; never a finale of its own.
func TestProgramSaveQuitEnding(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = ecl.V3
	e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: []ecl.Arg{{N: 1}}}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.title.endAnim = true
	e.events = nil
	e.opProgram()
	if e.sequence != adventure.SeqEnding || slices.Contains(e.events, game.Event(game.Ended{})) {
		t.Fatalf("SaveQuit with an end: sequence %q, events %v: want the sequence up, the run on", e.sequence, e.events)
	}
	o, err := e.doNow(game.ContinueAction{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(o.Events, game.Event(game.Ended{})) {
		t.Fatalf("Continue on the ending: events %v, want Ended", o.Events)
	}
}

// TestProgramEnding: PROGRAM's ending (POR 8) plays the manifest's End and then the script goes
// on: no End, nothing waits; an End of pictures is a scene Continue ends, the run goes on.
func TestProgramEnding(t *testing.T) {
	e := quickStarted(t)
	program := func() {
		e.events = nil
		e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: []ecl.Arg{{N: 8}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opProgram()
	}
	program()
	if e.ext.ending || e.waiting(PendSequence) {
		t.Fatal("an ending with no End waits")
	}
	e.title.end = &game.Scene{Kind: game.SceneTitle, Intro: &game.Intro{Full: true}}
	program()
	if sc, ok := e.build().(game.Scene); !ok || sc.Kind != game.SceneTitle || !e.waiting(PendSequence) {
		t.Fatalf("the end is up as %T, waiting %v", e.build(), e.waiting(PendSequence))
	}
	o, err := e.doNow(game.ContinueAction{})
	if err != nil || e.ext.ending || e.waiting(PendSequence) || slices.Contains(o.Events, game.Event(game.Ended{})) {
		t.Fatalf("Continue: %v, ending %v, events %v: want the script on", err, e.ext.ending, o.Events)
	}
}

// TestEndScene: the end's scene states the party's members and the message delay gobox shows
// its pictures by (adventure.Layer.MinParty, MsgDelayed).
func TestEndScene(t *testing.T) {
	e := quickStarted(t)
	e.title.end = &game.Scene{Kind: game.SceneTitle, Intro: &game.Intro{Pictures: []int{1}, End: true}}
	e.msgDelay = 5
	e.setMemberCount(3)
	in := e.endScene(game.Common{}).Intro
	if in.Party != 3 || in.MsgDelay != 5 || !in.End || !slices.Equal(in.Pictures, []int{1}) {
		t.Fatalf("end intro %+v", in)
	}
	if e.title.end.Intro.Party != 0 {
		t.Fatal("the end's own intro was changed")
	}
}

// TestEndEffects: an End's After goes through the ruleset when it is over: the party is healed
// and the hall opens; the closing note is said.
func TestEndEffects(t *testing.T) {
	e := quickStarted(t)
	id := e.partyIDs()[0]
	e.do(rules.CloseHall{})
	before := e.ruleset().Sheet(id)
	if before.HP == 0 {
		t.Fatal("no member to heal")
	}
	e.title.after = []adventure.Effect{adventure.EffectHealParty, adventure.EffectOpenHall}
	e.title.note = "You've won."
	e.endEffects()
	if !e.trainingOpen() {
		t.Fatal("the hall is closed, want open")
	}
	if got := e.ruleset().Sheet(id); got.HP != got.MaxHP {
		t.Fatalf("HP %d of %d after the ending", got.HP, got.MaxHP)
	}
	if e.ext.note == "" {
		t.Fatal("no closing note")
	}
}
