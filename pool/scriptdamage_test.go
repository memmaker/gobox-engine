package pool

import (
	"testing"

	"gobox/ecl/script"
)

// TestScriptDamage: DAMAGE goes through the rules' ScriptDamage. Everyone
// with a save at +16: without 0x10 (ruling 3) a made save spares, with it
// every living member takes the 1d1+2.
func TestScriptDamage(t *testing.T) {
	run := func(flags int) (hurt, spared int) {
		e := quickStarted(t)
		before := make([]int, len(e.state().Party))
		for i, c := range e.state().Party {
			before[i] = int(c.HP)
		}
		e.vm.Script = &script.Script{Code: []script.Instr{{Op: 0x2E, Args: []script.Arg{{N: flags}, {N: 1}, {N: 1}, {N: 2}, {N: 0}}}, {Op: script.OpExit}}}
		e.vm.PC, e.vm.Stop = 0, false
		e.eclRun()
		for i, c := range e.state().Party {
			switch before[i] - int(c.HP) {
			case 3:
				hurt++
			case 0:
				spared++
			default:
				t.Fatalf("flags %#x: member %d lost %d HP", flags, i, before[i]-int(c.HP))
			}
		}
		return hurt, spared
	}
	if hurt, spared := run(0xC0 | 0x0F); spared == 0 || hurt+spared == 0 {
		t.Errorf("save +15, spares: %d hurt, %d spared; want some spared", hurt, spared)
	}
	if hurt, spared := run(0xC0 | 0x10); spared != 0 || hurt == 0 {
		t.Errorf("0x10, a made save hurts: %d hurt, %d spared; want none spared", hurt, spared)
	}
}
