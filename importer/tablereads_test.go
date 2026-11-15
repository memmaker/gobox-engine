package importer

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"gobox/ecl"
)

// TestTableReadsFitTheirField: in every game's imported scripts, a table of
// the script's own that a variable of a declared range reads (Facing,
// range=4) has an entry for each value: the read never leaves its field
// (ecl.ConstLayout runs a short table on into the next; POR area 29's exits
// by facing were three fields, the west read past each). The detector of
// the bug class: a new variable-indexed table that is too short fails here.
func TestTableReadsFitTheirField(t *testing.T) {
	games := []struct {
		name  string
		reads func(dir string) ([]ecl.TableRead, error)
	}{
		{"por", func(dir string) ([]ecl.TableRead, error) {
			if err := ImportIfStale("../data/por/sources/dos", dir); err != nil {
				return nil, err
			}
			return ecl.TableReads(ecl.V1, dir, os.DirFS(dir), porScriptFixes)
		}},
		{"cotab", func(dir string) ([]ecl.TableRead, error) {
			return importedReads(dir, "cotab", CotabImportScriptsTo, ecl.V2, embeddedScripts("cotab"), cotabScriptFixes)
		}},
		{"sotsb", func(dir string) ([]ecl.TableRead, error) {
			return importedReads(dir, "sotsb", SotsbImportScriptsTo, ecl.V2, embeddedScripts("sotsb"), journalFixes)
		}},
		{"pod", func(dir string) ([]ecl.TableRead, error) {
			return importedReads(dir, "pod", PodImportScriptsTo, ecl.V3, podScriptData(), podScriptFixes)
		}},
	}
	for _, g := range games {
		t.Run(g.name, func(t *testing.T) {
			if _, err := os.Stat(filepath.Join("..", "data", g.name, "sources", "dos")); err != nil {
				t.Skip("no originals:", err)
			}
			reads, err := g.reads(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			ranged := 0
			for _, r := range reads {
				if r.Range == 0 {
					continue
				}
				ranged++
				if r.Offset+r.Range > r.Len {
					t.Errorf("area %d: %s[%s] reads %d entries from %d of a field of %d", r.Area, r.Table, r.Index, r.Range, r.Offset, r.Len)
				}
			}
			t.Logf("%d table reads by a variable, %d by one of a declared range", len(reads), ranged)
		})
	}
}

// importedReads imports game's scripts into dir and lists their table reads.
func importedReads(dir, game string, imp func(dos, dir string) error, v *ecl.Version, data fs.FS, fixes []ecl.Fix) ([]ecl.TableRead, error) {
	if err := imp(filepath.Join("..", "data", game, "sources", "dos"), dir); err != nil {
		return nil, err
	}
	return ecl.TableReads(v, dir, data, fixes)
}
