//go:build devviews

package engine

import (
	"gobox/ecl"
	"gobox/game"
	"gobox/importer"
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
	e.title.seqs = []game.Sequence{{Name: seqEnding}}
	e.events = nil
	e.opProgram()
	if e.sequence != seqEnding || slices.Contains(e.events, game.Event(game.Ended{})) {
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

// TestEndScene: the end's pictures show only as many party pictures as the party has members,
// and an animation's frames last as the message delay says.
func TestEndScene(t *testing.T) {
	e := quickStarted(t)
	layer := func(min int, ms int) game.Layer {
		return game.Layer{MinParty: min, FrameMs: ms, MsgDelayed: ms != 0}
	}
	e.title.end = &game.Scene{Kind: game.SceneTitle, Steps: map[int]game.TitleStep{1: {Text: []string{"x"}, Layers: []game.Layer{layer(0, 0), layer(2, 0), layer(0, 160)}}}}
	e.msgDelay = 5 // (5+3)/(1+3) = twice the default
	for _, members := range []int{1, 3} {
		e.setWorldWord(rules.WordPartyCount, members)
		st := e.endScene(game.Common{}).Steps[1]
		want := 2
		if members >= 2 {
			want = 3
		}
		if len(st.Layers) != want || !slices.Equal(st.Text, []string{"x"}) {
			t.Fatalf("%d members: %d layers, want %d", members, len(st.Layers), want)
		}
		if l := st.Layers[len(st.Layers)-1]; l.FrameMs != 320 || l.MinParty != 0 || l.MsgDelayed {
			t.Fatalf("animation layer %+v: want 320 ms, filters cleared", l)
		}
	}
	if e.title.end.Steps[1].Layers[2].FrameMs != 160 {
		t.Fatal("the end's own layers were changed")
	}
}

// TestEndEffects: an End's After goes through the ruleset when it is over: the party is healed
// and the hall opens; the closing note is said.
func TestEndEffects(t *testing.T) {
	e := quickStarted(t)
	id := e.partyIDs()[0]
	e.ruleset().Resolve(nil, rules.SetWord{Word: rules.WordTraining, Value: 0}, nil)
	before := e.ruleset().Sheet(id)
	if before.HP == 0 {
		t.Fatal("no member to heal")
	}
	e.title.after = []importer.Effect{importer.EffectHealParty, importer.EffectOpenHall}
	e.title.note = "You've won."
	e.endEffects()
	if a := e.ruleset().Check(rules.WordIs{Word: rules.WordTraining}, nil); !a.Yes || a.Value == 0 {
		t.Fatalf("training word %+v: want open", a)
	}
	if got := e.ruleset().Sheet(id); got.HP != got.MaxHP {
		t.Fatalf("HP %d of %d after the ending", got.HP, got.MaxHP)
	}
	if e.ext.note == "" {
		t.Fatal("no closing note")
	}
}
