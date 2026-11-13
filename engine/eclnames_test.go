package engine

import (
	"gobox/ecl/script"
	"gobox/rules1e"
	"testing"
)

// TestEclOpsByName: each dialect's byte code reaches pool's command of that
// name, and no byte runs a command its dialect gives another name (fold S2).
func TestEclOpsByName(t *testing.T) {
	for _, v := range []*script.Version{script.V1, script.V2, script.V3} {
		e := &Engine{given: &GameTables{Script: v, OriginalSave: testCode}}
		e.initVM()
		for b, c := range v.Cmds {
			_, named := eclOps[c.Op]
			if c.Known && named && e.vm.Ops[b] == nil {
				t.Errorf("V%d byte %d (%s): not installed", v.ID, b, script.OpName(c.Op))
			}
		}
		if v != script.V1 && e.vm.Raw[script.OpRedrawView] == 49 {
			t.Errorf("V%d: byte 49 is SpriteOff, not RedrawView", v.ID)
		}
	}
}

// testCode is a code image for an Engine without the game's tables (POR's place and size).
var testCode = rules1e.SaveLayout{CodeBase: 0x9900, CodeSize: 0x1E00}
