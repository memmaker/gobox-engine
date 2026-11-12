package pool

import (
	"encoding/json"
	"gobox/ecl/script"
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenECLCommands: every command byte each game's own script machine
// dispatched (captured from sotsb, pod and cotab's rules/ before they went;
// each was equal to its dialect's Known set but 65 and 66, the games' own)
// runs in pool under that dialect.
func TestGoldenECLCommands(t *testing.T) {
	for game, v := range map[string]*script.Version{"sotsb": script.V2, "pod": script.V3, "cotab": script.V2} {
		b, err := os.ReadFile(filepath.Join("..", "..", "rules1e", "testdata", game, "eclcommands.golden.json"))
		if err != nil {
			t.Fatal(err)
		}
		var handled []int
		if err := json.Unmarshal(b, &handled); err != nil {
			t.Fatal(err)
		}
		e := &Engine{given: &GameTables{Script: v}}
		e.initVM()
		for _, c := range handled {
			if e.vm.Ops[c] == nil {
				t.Errorf("%s: command %d (%s) is dispatched by the game, not by pool", game, c, script.OpName(v.Cmds[c].Op))
			}
		}
		if e.vm.Args != v.Skip {
			t.Errorf("%s: the machine's skip counts are not the dialect's", game)
		}
	}
}
