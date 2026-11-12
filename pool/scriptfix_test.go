package pool

import (
	"gobox/ecl/script"
	"rules1e"
	"testing"
)

// The cabinets' CAST KNOCK: the first okay member with Knock memorized, in
// any slot, casts it (the slot spent) and the script goes on at KnockOpens
// with that member's place; with none (the dead do not cast) at NoKnock.
func TestCastKnock(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(14)
	c := *e.state().Party[0]
	e.state().Party = []*rules1e.Character{&c, {}, {}}
	for _, p := range e.state().Party {
		p.Status, p.Memorized = rules1e.StatusOkay, [21]rules1e.SpellID{}
	}
	e.state().Party[0].Status = rules1e.StatusDead
	e.state().Party[0].Memorized[3] = rules1e.SpKnock
	e.state().Party[2].Memorized[5] = rules1e.SpKnock
	e.castKnock()
	if e.vm.PC != uint16(e.scriptLabel("KnockOpens")) || e.host.Get(e.field(fieldScratch6)) != 2 || e.state().Party[2].Memorized[5] != 0 || e.state().Party[0].Memorized[3] != rules1e.SpKnock {
		t.Errorf("cast: pc %d, Scratch6 %d, slots %v %v", e.vm.PC, e.host.Get(e.field(fieldScratch6)), e.state().Party[0].Memorized, e.state().Party[2].Memorized)
	}
	e.castKnock()
	if e.vm.PC != uint16(e.scriptLabel("NoKnock")) {
		t.Errorf("none: pc %d", e.vm.PC)
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
	e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{N: 1}, {N: 100}}}}}
	e.opAddNPC()
	if len(e.state().Party) != 8 || e.state().Party[3].Control != 0 {
		t.Errorf("%d members, member 3 control %#x", len(e.state().Party), e.state().Party[3].Control)
	}
}
