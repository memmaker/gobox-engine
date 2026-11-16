package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/game"
)

// br1SpaceBlock writes a block of the test's own into the imported folder whose load entry is BR1's own
// SpaceFight line (area34 L_8A4D's operands), as the engine's tests do.
func br1SpaceBlock(t *testing.T, dir string, block string) {
	t.Helper()
	head, _, _ := strings.Cut(mustRead(t, filepath.Join(dir, "scripts", "area16.txt")), "\n")
	text := head + "\non load {\n    AreaKind(3, 255)\n    Word4C61 = 1\n    SpaceFight(Word4C61, 1, 50, 18)\n    exit\n}\n"
	for _, entry := range []string{"move", "step", "camp", "rest"} {
		text += "\non " + entry + " {\n    exit\n}\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "area"+block+".txt"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestBR1SpaceFight: the ship fight's first states and the ship sheet are drawn by the presenter, each with its
// question on a screen (no undrawn question), and the fight's turns go on from the front's keys.
func TestBR1SpaceFight(t *testing.T) {
	f := br1Folder(t)
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	br1SpaceBlock(t, f.Dir, "201")
	eng.DevAction = func() { eng.DevAt(201, 0, 0) }
	br1Begin(t, g, eng)
	br1Enter(t, g, eng, "space fight", game.ShipFight{}, 80)
	sf := eng.State().(game.ShipFight)
	if sf.Fight.Own.Max[0] == 0 || sf.Fight.Enemy.Max[0] == 0 || sf.Ask == nil {
		t.Fatalf("the fight's first state: %+v ask %+v", sf.Fight, sf.Ask)
	}
	flowKey(g, 'v') // View: the own ship's sheet over the fight
	br1Check(t, g, eng, "ship sheet", game.ShipFight{})
	if sf = eng.State().(game.ShipFight); sf.Sheet == nil || !sf.Sheet.Own {
		t.Fatalf("no ship sheet after View: %+v", sf.Sheet)
	}
	g.Key(keyEnter) // the key that puts it away
	br1Check(t, g, eng, "the menu again", game.ShipFight{})
	for range 12 {
		g.Key(keyEnter)
		br1Frames(g, 6)
		if _, ok := eng.State().(game.ShipFight); !ok {
			return
		}
		br1Check(t, g, eng, "fight turn", game.ShipFight{})
	}
}
