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
