package pool

import (
	"gobox/core"
	"gobox/ecl/script"
	"gobox/game/play"
	"slices"
	"testing"
)

// TestSoundOp: Sound plays the game's cue for its number (ScriptSounds), nothing for an unlisted one.
func TestSoundOp(t *testing.T) {
	e := quickStarted(t)
	g := e.game()
	g.ScriptSounds = map[byte]ScriptSound{3: {Cue: core.SoundCue("3")}}
	e.given = &g
	sound := func(n int) {
		e.events = nil
		e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: n}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opSound()
	}
	sound(3)
	if !slices.Contains(e.events, play.Event(play.Sounded{Sound: "3"})) {
		t.Fatalf("sound 3: events %v", e.events)
	}
	sound(4)
	if len(e.events) != 0 {
		t.Fatalf("unlisted sound 4: events %v", e.events)
	}
}
