package engine

import (
	"testing"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules1e"
)

// The member's STATUS cell of the four Pool games, read and written through the game's encoding
// (what getCell/setVar do: the binding's set after adventure.Canonical, the binding's get then the codec's toGame).
// Expectations are the originals' Pascal:
//   POR   por-re/src/eclmenu.pas SpecialReadZ1 $100 (1 if enabled <> 0 else $80), SpecialWriteZ1 $100
//         (>= $80: enabled := 0, $87 also status := STONED; 0: dumpOnStatusClear only; else nothing),
//         eclvm.pas EclLoadCharacter (the removal needs dumpOnStatusClear and dumpOnNameClear).
//   COTAB cotab-re/src eclmenu.pas $100 (as POR's, 0 once after characterNotFound), SpecialWriteZ1 $100 as POR's.
//   SotSB sotsb-re/src eclmenu.pas SpecialReadZ1 $100 (the status 1..7, charNotFound 0), SpecialWriteZ1 $100 (1 only:
//         status := OKAY, enabled := 1).
//   POD   pod-re/src eclmenu.pas SpecialWriteZ1 $1ED (status := Byte(value); inPlay := Byte(status = OKAY)).

// selStatusState is the observable member: record status, enabled/in-play, the removal mark (memberDied).
type selStatusState struct {
	status, okay, mark byte
}

type selStatusWrite struct {
	value int64
	from  selStatusState
	want  selStatusState // the Pascal's
}

// selStatusEngine is an engine whose selected member is in state s; enc is the layout's encoding of SelStatus.
func selStatusEngine(s selStatusState) (*Engine, *rules1e.Character) {
	e := &Engine{}
	m := &rules1e.Character{}
	m.Status, m.Okay = s.status, s.okay
	e.state().Party = []*rules1e.Character{m}
	e.es.memberDied = s.mark
	return e, m
}

func selStatusRead(e *Engine, enc string) int64 {
	v, _ := e.newEngineFields()["SelStatus"].get(0)
	if c, ok := codecs[ecl.Encoding(enc)]; ok && c.toGame != nil {
		return c.toGame(e, v)
	}
	return v
}

func selStatusSet(e *Engine, enc string, v int64) {
	e.newEngineFields()["SelStatus"].set(0, adventure.Canonical(ecl.Encoding(enc), v, 0))
}

func TestSelStatusWrites(t *testing.T) {
	ok := selStatusState{rules1e.StatusOkay, 1, 0}
	dead := selStatusState{rules1e.StatusDead, 0, 0}
	out := selStatusState{rules1e.StatusOkay, 0, 0}
	stoned := selStatusState{rules1e.StatusStoned, 0, 0}
	marked := selStatusState{rules1e.StatusOkay, 1, 1}
	pool := func() []selStatusWrite { // POR and COTAB, SpecialWriteZ1 $100
		return []selStatusWrite{
			{0, ok, marked}, // dumpOnStatusClear only: the status is not changed
			{1, dead, dead}, // nothing
			{2, ok, ok}, {0x7F, ok, ok},
			{0x80, ok, out}, {0x81, ok, out}, {0xFF, ok, out},
			{0x87, ok, stoned},
		}
	}
	games := []struct {
		name, enc string
		rows      []selStatusWrite
	}{
		{"por", "PorSelStatus", pool()},
		{"cotab", "CotabSelStatus", pool()},
		{"sotsb", "SotsbSelStatus", []selStatusWrite{
			{1, dead, ok}, // status := OKAY; enabled := 1
			{0, ok, ok}, {2, ok, ok}, {3, dead, dead}, {7, ok, ok},
			{0x80, ok, ok}, {0x87, ok, ok},
		}},
		{"pod", "PodSelStatus", []selStatusWrite{
			{0, dead, ok}, // status := OKAY; inPlay := 1
			{1, ok, selStatusState{rules1e.StatusAnimated, 0, 0}},
			{3, ok, selStatusState{rules1e.StatusRunning, 0, 0}},
			{4, ok, selStatusState{rules1e.StatusUnconscious, 0, 0}},
			{5, ok, selStatusState{rules1e.StatusDying, 0, 0}},
			{6, ok, dead},
			{7, ok, stoned},
			{8, ok, selStatusState{rules1e.StatusGone, 0, 0}},
			{9, ok, selStatusState{rules1e.StatusReserve, 0, 0}},
		}},
	}
	for _, g := range games {
		for _, w := range g.rows {
			e, m := selStatusEngine(w.from)
			selStatusSet(e, g.enc, w.value)
			got := selStatusState{m.Status, m.Okay, e.es.memberDied}
			if got != w.want {
				t.Errorf("%s: write %#x on %+v: %+v, want %+v", g.name, w.value, w.from, got, w.want)
			}
		}
	}
}

// TestSelStatusPodUnnamed: a status number pod's C_Status has no name for (2, 10 up) is stored raw by the original;
// the rules have no status for it: the write fails loudly.
func TestSelStatusPodUnnamed(t *testing.T) {
	for _, v := range []int64{2, 10, 0xFF} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("pod write %d did not fail", v)
				}
			}()
			e, _ := selStatusEngine(selStatusState{})
			selStatusSet(e, "PodSelStatus", v)
		}()
	}
}

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func TestSelStatusReads(t *testing.T) {
	type row struct {
		from selStatusState
		want int64
	}
	var porRows []row
	for _, st := range []byte{rules1e.StatusOkay, rules1e.StatusAnimated, rules1e.StatusDead, rules1e.StatusStoned, rules1e.StatusGone} {
		porRows = append(porRows, row{selStatusState{st, 1, 0}, 1}, row{selStatusState{st, 0, 0}, 0x80}) // enabled <> 0
	}
	sotsb := map[byte]int64{rules1e.StatusOkay: 1, rules1e.StatusGone: 2, rules1e.StatusDead: 3, rules1e.StatusDying: 4,
		rules1e.StatusUnconscious: 5, rules1e.StatusRunning: 6, rules1e.StatusStoned: 7, rules1e.StatusAnimated: 0, rules1e.StatusReserve: 0}
	var sotsbRows, podRows []row
	for st, n := range sotsb {
		sotsbRows = append(sotsbRows, row{selStatusState{st, b2i(st == rules1e.StatusOkay), 0}, n})
	}
	for _, st := range []byte{0, 1, 3, 4, 5, 6, 7, 8, 9} { // the record's status byte
		podRows = append(podRows, row{selStatusState{st, b2i(st == rules1e.StatusOkay), 0}, int64(st)})
	}
	for _, g := range []struct {
		name, enc string
		rows      []row
	}{{"por", "PorSelStatus", porRows}, {"cotab", "CotabSelStatus", porRows}, {"sotsb", "SotsbSelStatus", sotsbRows}, {"pod", "PodSelStatus", podRows}} {
		for _, r := range g.rows {
			e, _ := selStatusEngine(r.from)
			if got := selStatusRead(e, g.enc); got != r.want {
				t.Errorf("%s: read of %+v is %d, want %d", g.name, r.from, got, r.want)
			}
		}
	}
	// SelectMember that matched nobody: cotab reads 0 once (characterNotFound), sotsb 0 (charNotFound).
	e, _ := selStatusEngine(selStatusState{rules1e.StatusOkay, 1, 0})
	e.es.memberMissing = true
	if got := selStatusRead(e, "CotabSelStatus"); got != 0 {
		t.Errorf("cotab no member: %d", got)
	}
	if got := selStatusRead(e, "CotabSelStatus"); got != 1 {
		t.Errorf("cotab second read: %d", got)
	}
	e.es.memberMissing = true
	if got := selStatusRead(e, "SotsbSelStatus"); got != 0 { // charNotFound; it stays until the next select that matches
		t.Errorf("sotsb no member: %d", got)
	}
	if got := selStatusRead(e, "SotsbSelStatus"); got != 0 {
		t.Errorf("sotsb no member, second read: %d", got)
	}
}
