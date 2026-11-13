package pool

import (
	"bufio"
	"fmt"
	"gobox/core/grid"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// TestDoorAgainstPascal compares lockedDoor and TryDoor with Explore.ResolveMovement
// (rules1e testdata/por-door.txt, por-re vecdoor): the Locked menu's offers, the wall
// when nothing is offered, Pick on a wizard-locked door, and Knock. An attempt not
// offered is Exit in the Pascal; Go refuses it, so the case is answered with Exit.
// Bash and kind 2's Pick roll dice and are not in the vectors.
func TestDoorAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-door.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "door ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		w := atois(t, strings.Fields(val))
		kind, choices, thief, mage, knock, attempt := k[0], k[1], k[2], k[3], k[4], k[5]
		e := quickStarted(t)
		c := e.state().Party[0]
		e.state().Party = e.state().Party[:1]
		c.Levels = [len(c.Levels)]rules1e.Level{}
		for i := range c.Memorized {
			c.Memorized[i] = 0
		}
		if thief == 1 {
			c.Levels[6] = 3
		}
		if mage == 1 {
			c.Levels[5] = 3
		}
		if knock >= 1 {
			c.Memorized[4] = rules1e.SpKnock
		}
		if knock == 2 { // a dead member's Knock
			c.Status = rules1e.StatusDead
		}
		e.run.doorTry = [3]byte{byte(choices & 1), byte(choices >> 1 & 1), byte(choices >> 2 & 1)}
		e.ext.door = nil
		e.ext.doorThrough = false
		e.lockedDoor(kind == 3)
		walled := e.pending != PendDoor
		if !walled {
			way := [3]DoorWay{DoorExit, DoorPick, DoorKnock}[attempt]
			if e.TryDoor(way) != nil {
				if err := e.TryDoor(DoorExit); err != nil {
					t.Fatalf("%s: Exit refused: %v", line, err)
				}
			}
		}
		through := e.ext.doorThrough
		if w[4] == 1 { // the Pascal's Knock opened the door ahead (1): so must Go's
			x, y, d := e.posX, e.posY, e.moveDir()
			if !grid.OnMap(x, y) || e.gs.geo.Plane[0x300+y<<4+x]>>uint(d)&3 != 1 {
				t.Errorf("%s: Go left the door shut", line)
			}
		}
		got := [4]bool{walled, e.run.doorTry[1] != 0, rules1e.SpellSlot(c, rules1e.SpKnock) != 0xFF, through}
		want := [4]bool{w[0] == 1, w[1] != 0, w[2] == 1, w[3] != 3}
		if through { // the Pascal's step reopens the choices; Go's step comes later
			got[1], want[1] = false, false
		}
		if got != want {
			t.Errorf("%s: Go wall, pickLeft, knockLeft, through %v", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestDoorDiceAgainstPascal compares bashDoor and pickLock with Explore.BashDoor and
// Explore.PickLock on a shared seed (rules1e testdata/por-door.txt, por-re vecdoor):
// opened, the choice left and the RNG state after.
func TestDoorDiceAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-door.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	party := e.state().Party[:2]
	e.ext.back = 0
	e.posX, e.posY, e.facing = 3, 2, 6 // as vecdoor: (3,2) facing west, a wall there (DoorAt reads a door only where one stands)
	e.gs.geo.Plane[0x100+2<<4+3] = 0x11
	e.gs.geo.Plane[0x100+2<<4+2] = 0x11
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		bash := strings.HasPrefix(line, "bash ")
		if !bash && !strings.HasPrefix(line, "pick ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		w := atois(t, strings.Fields(val))
		e.state().Party = party
		for _, c := range party {
			c.Status = rules1e.StatusOkay
		}
		a, b := party[0], party[1]
		e.run.doorTry = [3]byte{1, 1, 1}
		kind := 2
		if bash {
			kind = k[1]
		}
		e.setDoor(e.posX, e.posY, e.moveDir(), byte(kind))
		var opened bool
		var left byte
		if bash {
			a.Stats[rules1e.StatStr], a.Str00 = rules1e.Score(k[2]), byte(k[3])
			b.Stats[rules1e.StatStr], b.Str00 = 18, 100
			if k[4] == 1 {
				b.Status = rules1e.StatusDead
			}
			e.state().RandSeed = uint32(int32(k[0]))
			opened, left = e.bashDoor(), e.run.doorTry[0]
		} else {
			a.Thief[1], b.Thief[1] = rules1e.Percent(k[1]), rules1e.Percent(k[2])
			if k[3] == 1 {
				a.Status = rules1e.StatusDead
			}
			e.state().RandSeed = uint32(int32(k[0]))
			opened, left = e.pickLock(), e.run.doorTry[1]
		}
		got := fmt.Sprintf("%d %d %d", rules1e.B2i(opened), left, int32(e.state().RandSeed))
		if want := fmt.Sprintf("%d %d %d", w[0], w[1], w[2]); got != want {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
