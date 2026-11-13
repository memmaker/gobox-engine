package importpod

import (
	"encoding/json"
	eclscript "gobox/ecl/script"
	"gobox/pool"
	"os"
	"path/filepath"
	"rules1e"
	"testing"
)

// TestOriginalRules: a real pod save's record words reach rules1e by canonical name: the
// clock, 3D flag, text delay and party size are the save's own words at the layout's fields
// (file offset: the game block's bytes from its base).
func TestOriginalRules(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "data", "pod", "extras", "scenarios", "saves", "inside", "SAVGAMA.PTY"))
	if err != nil {
		t.Skip("no save:", err)
	}
	l, err := eclscript.ReadLayout(scriptData(), 0)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := pool.OriginalRules(b, pool.GameTables{OriginalSave: OriginalSave}, l)
	if err != nil {
		t.Fatal(err)
	}
	var s rules1e.SaveBlob
	if err := json.Unmarshal(blob, &s); err != nil {
		t.Fatal(err)
	}
	word := func(name string) uint16 {
		r, err := l.Named(name)
		if err != nil {
			t.Fatal(err)
		}
		return uint16(b[l.Field(r).From-OriginalSave.Records[0].Base]) // one record of byte cells
	}
	got := map[string]uint16{"Hours": s.Area.Clock[3], "TenMinutes": s.Area.Clock[2], "In3D": s.Area.In3D, "MsgDelay": s.Area.MsgDelay, "PartyCount": s.Game.PartyCount}
	set := 0
	for name, v := range got {
		if w := word(name); v != w {
			t.Errorf("%s: rules1e %d, the save's word %d", name, v, w)
		}
		if v != 0 {
			set++
		}
	}
	if set == 0 {
		t.Error("no word of the save reached rules1e")
	}
}
