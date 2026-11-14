package engine

import (
	"gobox/ecl"
	"gobox/rules1e"
	"testing"
)

// The cabinets' CastKnock(Scratch6): the first okay member with Knock
// memorized, in any slot, casts it (the slot spent) and Scratch6 is that
// member's place; with none (the dead do not cast) 255.
func TestCastKnock(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(14)
	c := *e.state().Party[0]
	e.state().Party = []*rules1e.Character{&c, {}, {}}
	for _, p := range e.state().Party {
		p.Status, p.Memorized = rules1e.StatusOkay, [rules1e.MemorySlots]rules1e.SpellID{}
	}
	e.state().Party[0].Status = rules1e.StatusDead
	e.state().Party[0].Memorized[3] = rules1e.SpKnock
	e.state().Party[2].Memorized[5] = rules1e.SpKnock
	knock := func() int64 { // CastKnock(Scratch6) at its label
		e.vm.PC = uint16(e.scriptLabel("CastKnock"))
		e.vm.Fetch()
		e.opCastKnock()
		return e.host.Get(e.field("Scratch6"))
	}
	if got := knock(); got != 2 || e.state().Party[2].Memorized[5] != 0 || e.state().Party[0].Memorized[3] != rules1e.SpKnock {
		t.Errorf("cast: Scratch6 %d, slots %v %v", got, e.state().Party[0].Memorized, e.state().Party[2].Memorized)
	}
	if got := knock(); got != noKnock {
		t.Errorf("none: Scratch6 %d", got)
	}
}

// A full party takes no NPC, and no member is made one.
func TestAddNPCFullParty(t *testing.T) {
	e := &Engine{}
	e.initVM()
	for range 8 {
		e.state().Party = append(e.state().Party, &rules1e.Character{})
	}
	e.sel = 3
	e.state().Game.PartyCount = 8
	e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: []ecl.Arg{{N: 1}, {N: 100}}}}}
	e.opAddNPC()
	if len(e.state().Party) != 8 || e.state().Party[3].Control != 0 {
		t.Errorf("%d members, member 3 control %#x", len(e.state().Party), e.state().Party[3].Control)
	}
}
