package pool

import (
	"bufio"
	"fmt"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// TestRestAgainstPascal compares Rest and stepRest with GameTime.CampRest
// (rules1e testdata/por-rest.txt, por-re vecrest) on a shared seed: the rest
// menu's Rest or Exit (Exit rests nothing), Stop Resting after n ticks
// (StopRest's flag), a rest of 30 minutes or 2 hours, the area's safe-rest
// period (game.RestEvery, camp.interval) at 50 percent. The rest time is read
// before RestStep clears it.
func TestRestAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-rest.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := testEngine(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "rest ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		seed, rest, stopAfter, minutes, safe := k[0], k[1] == 1, k[2], k[3], k[4]
		e.state().Camp.Rest = [7]int{}
		if minutes == 120 {
			e.state().Camp.Rest[3] = 2
		} else {
			e.state().Camp.Rest[2] = 3
		}
		e.state().Camp.Interval = 0
		e.state().Game.RestEvery, e.state().Game.RestChance = uint16(safe), 50
		e.state().RandSeed = uint32(int32(seed))
		interrupted := false
		if rest {
			e.Rest(false, restThen{})
			for ticks := 0; ; ticks++ {
				if stopAfter != 0 && ticks == stopAfter {
					e.state().Camp.StopRest = true // StopRest: Yes to Stop Resting
				}
				if s := e.restTick(false); s.over {
					interrupted = s.interrupted
					break
				}
			}
			e.resting, e.restLine = nil, false
		}
		r := e.state().Camp.Rest
		got := fmt.Sprintf("%d %d %d %d %d %d %d", rules1e.B2i(interrupted), r[1], r[2], r[3], r[4], e.state().Camp.Interval, int32(e.state().RandSeed))
		if got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestRestTickAgainstPascal compares what a rest tick does with CampRest's
// (por-rest.txt "tick" lines): one member at 5 of 20 hp with spells 1 and 2
// pending and 0 or 4 study hours (Study), resting 30 minutes, 2 hours or a day
// and an hour from a zero clock: hp, the two slots, Study, the heal counter
// (camp.heal) and the clock.
func TestRestTickAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-rest.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "tick ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		c := e.state().Party[0]
		e.state().Party = e.state().Party[:1]
		c.HP, c.MaxHP, c.Status = 5, 20, rules1e.StatusOkay
		c.Memorized = [len(c.Memorized)]rules1e.SpellID{0x81, 0x82}
		c.Study = byte(k[0])
		clear(e.state().Area.Clock[:])
		e.state().Camp.Heal = 0
		e.state().Game.RestEvery = 0
		e.state().Camp.Rest = [7]int{}
		switch k[1] {
		case 0:
			e.state().Camp.Rest[2] = 3
		case 1:
			e.state().Camp.Rest[3] = 2
		case 2:
			e.state().Camp.Rest[4], e.state().Camp.Rest[3] = 1, 1
		}
		c.Surges = 1
		e.Rest(false, restThen{})
		for !e.restTick(false).over {
		}
		e.resting, e.restLine = nil, false
		// ruling 6: the original's rule set does nothing at a long rest; the daily uses are the 4e set's
		if c.Surges != 1 {
			t.Errorf("%s: Surges %d after the rest, the 1e set left them", line, c.Surges)
		}
		got := fmt.Sprintf("%d %d %d %d %d", c.HP, c.Memorized[0], c.Memorized[1], c.Study, e.state().Camp.Heal)
		for i := 1; i <= 6; i++ {
			got += fmt.Sprintf(" %d", e.state().Area.Clock[i])
		}
		if got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestRestStopHealed: the stopHealed rest (an extra, resting until healed; no
// Pascal counterpart) ends on the tick nobody is wounded, before its time.
func TestRestStopHealed(t *testing.T) {
	e := quickStarted(t)
	c := e.state().Party[0]
	e.state().Party = e.state().Party[:1]
	c.HP, c.MaxHP, c.Status = 19, 20, rules1e.StatusOkay
	e.state().Game.RestEvery = 0
	e.state().Camp.Heal = 0x120 - 1 // the next tick heals
	e.state().Camp.Rest = [7]int{}
	e.state().Camp.Rest[4] = 2
	e.Rest(true, restThen{})
	ticks := 0
	for !e.restTick(true).over {
		ticks++
	}
	e.resting, e.restLine = nil, false
	if c.HP != 20 || ticks != 1 {
		t.Errorf("hp %d after %d ticks, want 20 after 1", c.HP, ticks)
	}
}
