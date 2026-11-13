package engine

import (
	"gobox/core"
	"gobox/ecl/script"
	"gobox/rules1e"
	"testing"
)

// The party questions through core.Rules Check (opPartyStrength,
// opCheckParty, opPartySurprise) answer as the Pascal's loops over the party.
func TestWorldChecks(t *testing.T) {
	e := quickStarted(t)
	out := [4]script.Ref{}
	for i, n := range []string{"Scratch1", "Scratch2", "Scratch3", "Scratch4"} {
		out[i] = testField(t, e, n)
	}
	v := func(r script.Ref) script.Arg { return script.Arg{Type: 1, Ref: r} }
	run := func(op func(), args ...script.Arg) {
		e.vm.Script = &script.Script{Code: []script.Instr{{Args: args}}, Layout: e.vm.Script.Layout}
		e.vm.PC = 0
		e.vm.Fetch()
		op()
	}
	get := func(i int) int64 { return e.host.Get(out[i]) }

	var want byte
	for _, c := range e.state().Party {
		ac, th := max(int(c.AC[0])-0x3C, 0), max(int(c.THAC0)-0x27, 0)
		want += byte((int(c.HP) + 5*ac + 5*th + 8*int(c.Levels[rules1e.ClsMage]) + 4*int(c.Levels[rules1e.ClsCleric])) / 10)
	}
	run(e.opPartyStrength, v(out[0]))
	if get(0) != int64(want) {
		t.Errorf("PartyStrength %d, want %d", get(0), want)
	}

	checkParty := func(what script.Arg) {
		run(e.opCheckParty, what, script.Arg{N: 19}, v(out[0]), v(out[1]), v(out[2]), v(out[3]))
	}
	checkParty(script.Arg{N: 0})
	if get(3) != 0 {
		t.Errorf("CheckParty find traps: found %d with no affect", get(3))
	}
	e.state().AddAffect(e.state().Party[len(e.state().Party)-1], 0x13, 0, 0, 0)
	checkParty(script.Arg{N: 0})
	if get(0) != 0 || get(1) != 0 || get(2) != 0 || get(3) != 1 {
		t.Errorf("CheckParty find traps: %d %d %d found %d", get(0), get(1), get(2), get(3))
	}
	// the fourth result is the count (S-diffs 12): area13's `Scratch1 != 0` branches as found did
	e.state().AddAffect(e.state().Party[0], 0x13, 0, 0, 0)
	checkParty(script.Arg{N: 0})
	if get(3) != 2 {
		t.Errorf("CheckParty find traps: count %d, want 2", get(3))
	}
	lo, hi, sum := byte(0xFF), byte(0), 0
	for _, c := range e.state().Party {
		lo, hi, sum = min(lo, c.Move), max(hi, c.Move), sum+int(c.Move)
	}
	checkParty(v(testField(t, e, fieldSelMove)))
	if get(0) != int64(lo) || get(1) != int64(hi) || get(2) != int64(byte(sum/len(e.state().Party))) {
		t.Errorf("CheckParty move: %d %d %d, want %d %d %d", get(0), get(1), get(2), lo, hi, sum/len(e.state().Party))
	}
	if get(3) != int64(len(e.state().Party)) { // area20's fix sets Scratch1 back to the original's 0
		t.Errorf("CheckParty move: count %d, want the party's %d", get(3), len(e.state().Party))
	}

	for _, c := range e.state().Party {
		if c.Class == rules1e.ClsRanger || c.Class == rules1e.ClsClericRanger {
			c.Class = rules1e.ClsFighter
		}
	}
	run(e.opPartySurprise, v(out[0]), v(out[1]))
	if get(0) != 0 {
		t.Errorf("PartySurprise %d with no ranger", get(0))
	}
	e.state().Party[1].Class = rules1e.ClsRanger
	run(e.opPartySurprise, v(out[0]), v(out[1]))
	if get(0) != 1 || get(1) != 0 {
		t.Errorf("PartySurprise %d %d with a ranger", get(0), get(1))
	}
}

// TestCheckPartyThiefSkills: opCheckParty averages every thief skill the
// layout names (cotab's area32 asks SelPickPockets), as the selected
// member's skill fact.
func TestCheckPartyThiefSkills(t *testing.T) {
	e := quickStarted(t)
	out := [4]script.Ref{}
	for i, n := range []string{"Scratch1", "Scratch2", "Scratch3", "Scratch4"} {
		out[i] = testField(t, e, n)
	}
	e.vm.Script = &script.Script{Code: []script.Instr{{Args: []script.Arg{{Type: 1, Ref: testField(t, e, "SelPickPockets")}, {N: 0},
		{Type: 1, Ref: out[0]}, {Type: 1, Ref: out[1]}, {Type: 1, Ref: out[2]}, {Type: 1, Ref: out[3]}}}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.opCheckParty()
	party := e.partyIDs()
	a := e.ruleset().Check(core.Census{Party: party, Fact: core.SkillFact{Skill: rules1e.PickPockets}}, nil)
	if got := [4]int64{e.host.Get(out[0]), e.host.Get(out[1]), e.host.Get(out[2]), e.host.Get(out[3])}; got != [4]int64{int64(byte(a.Low)), int64(byte(a.High)), int64(byte(a.Avg)), int64(len(party))} {
		t.Errorf("CheckParty pick pockets: %v, census %+v of %d", got, a, len(party))
	}
}
