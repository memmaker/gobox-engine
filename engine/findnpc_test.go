package engine

import (
	"gobox/ecl/script"
	"testing"
)

// TestFindNPC: pod's FindNPC answers the hired NPC's place, +128 when not
// up, 255 for a member who is not an NPC.
func TestFindNPC(t *testing.T) {
	e := quickStarted(t)
	out := testField(t, e, "Scratch1")
	c := e.state().Party[1]
	find := func() int64 {
		e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{Type: 0x80, S: c.Name()}, {Type: 1, Ref: out}}}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opFindNPC()
		return e.host.Get(out)
	}
	if got := find(); got != npcNotFound {
		t.Fatalf("a party member: %d, want %d", got, npcNotFound)
	}
	c.Control = 0x80
	if got := find(); got != 1 {
		t.Fatalf("hired NPC: %d, want 1", got)
	}
	c.Okay = 0
	if got := find(); got != 1|npcNotOkay {
		t.Fatalf("NPC down: %d, want %d", got, 1|npcNotOkay)
	}
}
