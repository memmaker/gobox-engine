package engine

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gobox/rules"
)

const moveText = "WHERE DO YOU WISH TO GO?"

// br1OwnBlock writes a block of the test's own into the imported folder (as TestScriptedPlace does) whose
// load entry is body, and enters it. Its move entry asks, the others just exit.
func br1OwnBlock(t *testing.T, e *Engine, block int, body string) {
	t.Helper()
	head, _, _ := strings.Cut(readFile(t, filepath.Join(e.scriptDir(), "area16.txt")), "\n") // the version line of this import's layout
	text := head + "\non load {\n    AreaKind(3, 255)\n" + body + "\n    exit\n}\n"
	text += "\non move {\n    Text(\"" + moveText + "\")\n    PressKey()\n    exit\n}\n" // a place waits for a move
	for _, entry := range []string{"step", "camp", "rest"} {
		text += "\non " + entry + " {\n    exit\n}\n"
	}
	name := filepath.Join(e.scriptDir(), "area"+strconv.Itoa(block)+".txt")
	if err := os.WriteFile(name, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	e.push(cont{kind: kStepEnd})
	e.DevAt(block, 0, 0)
}

// br1Cell is a layout cell by name.
func br1Cell(t *testing.T, e *Engine, name string) int64 {
	t.Helper()
	r, ok := e.layoutField(name)
	if !ok {
		t.Fatalf("no %s in the layout", name)
	}
	return e.host.Get(r)
}

// TestBR1StatusWrite: a script's SelStatusKind = 136 stores 8 (setChar keeps v & 127): the member is comatose
// and the script reads 8 back.
func TestBR1StatusWrite(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, "    SelStatusKind = 136\n    NewGameBytes = 63 & SelStatusKind")
	st, err := e.ruleset().Fact(e.partyIDs()[0], rules.FactStatus)
	if err != nil || st != rules.StatusComatose {
		t.Fatalf("status %v %v", st, err)
	}
	if got := br1Cell(t, e, "NewGameBytes"); got != 8 {
		t.Fatalf("read back %d", got)
	}
}

// TestBR1Credits: a script reads the party's credits as a dword, writes a higher value (added to the selected
// member) and a lower one (paid), and the ruleset's purse follows exactly (BR ecl.go charVar/setChar OffMoney).
func TestBR1Credits(t *testing.T) {
	e := br1Begun(t)
	purse := func() rules.Copper {
		v, err := e.ruleset().Fact(e.selected(), rules.PurseFact{})
		if err != nil {
			t.Fatal(err)
		}
		return rules.Copper(v.(rules.Purse).Copper)
	}
	start := purse()
	br1OwnBlock(t, e, 202, "    D_4CE2 = D_7C2B\n    D_7C2B = 60000\n    D_7C2B += 60000\n    D_4CE6 = D_7C2B\n    D_7C2B -= 1234")
	if got := rules.Copper(br1Cell(t, e, "Word4CE2")); got != start {
		t.Errorf("read %d, the purse held %d", got, start)
	}
	if got := br1Cell(t, e, "Word4CE6"); got != 120000 { // past 16 bits: the dword is carried whole
		t.Errorf("read back %d", got)
	}
	if got := purse(); got != 120000-1234 {
		t.Errorf("purse %d", got)
	}
}
