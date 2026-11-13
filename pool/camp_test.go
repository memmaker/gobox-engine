package pool

import (
	"gobox/ecl/script"
	"gobox/rules1e"
	"testing"
)

// The scripts' "hour < 14" sky check is the original's; under the DaySky
// option it follows day 6:00-20:00 instead.
func TestSkyHour(t *testing.T) {
	e := &Engine{}
	e.initVM()
	// Compare [49C9] 14
	e.vm.Script = resolved(t, &script.Script{Code: []script.Instr{{Op: 0x03, Args: []script.Arg{{Type: 1, N: 0x49C9}, {Type: 0, N: 14}}}}}, 0)
	for _, c := range []struct {
		hour                 uint16
		original, byTheClock bool
	}{{1, true, false}, {6, true, true}, {13, true, true}, {15, false, true}, {19, false, true}, {20, false, false}} {
		for _, option := range []bool{false, true} {
			e.DaySky.Store(option)
			e.state().Area.Clock[3] = c.hour
			e.vm.PC = 0
			e.vm.Fetch()
			e.opCompare()
			want := c.original
			if option {
				want = c.byTheClock
			}
			if lt := e.vm.Flags[2]; lt != want {
				t.Errorf("hour %d, option %v: day sky %v, want %v", c.hour, option, lt, want)
			}
		}
	}
}

// The planned spells c does not hold are set to be memorized on the next
// rest, as far as the free slots go; what c holds stays.
func TestPlanSpells(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	c.SpellSlots[rules1e.SpellClassCleric][0] = 3 // three 1st-level cleric spells
	c.Plan = [rules1e.MemorySlots]rules1e.SpellID{rules1e.SpCureLightWounds, rules1e.SpCureLightWounds, rules1e.SpCureLightWounds}
	c.Memorized = [rules1e.MemorySlots]rules1e.SpellID{rules1e.SpBless, rules1e.SpCureLightWounds | rules1e.SpellPending}
	e.state().PlanSpells(c)
	want := [rules1e.MemorySlots]rules1e.SpellID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, rules1e.SpBless, rules1e.SpCureLightWounds | rules1e.SpellPending, rules1e.SpCureLightWounds | rules1e.SpellPending}
	if c.Memorized != want {
		t.Fatalf("got %x, want %x", c.Memorized, want)
	}
}
