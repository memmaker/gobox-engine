package importer_test

import (
	"io/fs"
	"slices"
	"testing"

	"gobox/ecl"
	"gobox/engine"
	"gobox/importer"
)

// unusedMemberFields are the member fields of each game's layout that the engine does not bind, because no
// script of the game reads or writes them (grep the name in data/<game>/scripts). An unbound member field is
// one global cell, not the selected member's: a script that used one would read and write nothing of the member.
// Each entry names the field and says why no script uses it. All of today's layouts bind every member field.
var unusedMemberFields = map[string][]string{
	"por":   {},
	"cotab": {},
	"sotsb": {},
	"pod":   {},
	"br1":   {},
	"br2":   {},
}

// TestMemberFieldsBound (bug class: a member field named in a layout but bound to nothing is a plain stored
// cell, so SelHpCur = SelHpMax healed nobody): every member field of every game's layout is bound by the
// engine, or listed above as one no script uses.
func TestMemberFieldsBound(t *testing.T) {
	data := map[string]fs.FS{"por": importer.ScriptData(), "cotab": importer.EmbeddedScripts("cotab"), "sotsb": importer.EmbeddedScripts("sotsb"),
		"pod": importer.PodScriptData(), "br1": importer.EmbeddedScripts("br1"), "br2": importer.EmbeddedScripts("br2")}
	bound := engine.BoundFields()
	for game, fsys := range data {
		l, err := ecl.ReadLayout(fsys, 0)
		if err != nil {
			t.Fatal(game, err)
		}
		var members []string
		for _, f := range l.Fields {
			if f.Owner != ecl.OwnerMember {
				continue
			}
			members = append(members, f.Name)
			if !slices.Contains(bound, f.Name) && !slices.Contains(unusedMemberFields[game], f.Name) {
				t.Errorf("%s: member field %s is bound to nothing: bind it to the member's fact (engine/scriptvars_member.go), or list it in unusedMemberFields with why no script uses it", game, f.Name)
			}
		}
		for _, n := range unusedMemberFields[game] {
			if !slices.Contains(members, n) || slices.Contains(bound, n) {
				t.Errorf("%s: %s is listed as an unbound member field but is not one", game, n)
			}
		}
		if len(members) == 0 {
			t.Errorf("%s: no member fields read from its layout", game)
		}
	}
}
