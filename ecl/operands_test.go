package ecl

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every imported area text of every game (data/<game>, its texts in its
// scripts folder or CacheDir) passes the load-time operand check: no byte
// operand past 0xFF, every Journal an entry of names/entries.json.
func TestGameOperands(t *testing.T) {
	for _, game := range []string{"por", "cotab", "pod", "sotsb"} {
		t.Run(game, func(t *testing.T) {
			dataDir := filepath.Join("..", "..", "data", game, "scripts")
			if _, err := os.Stat(filepath.Join(dataDir, "names")); err != nil {
				t.Skipf("no game data: %v", err)
			}
			texts, _ := filepath.Glob(filepath.Join(dataDir, "area*.txt"))
			if len(texts) == 0 {
				dir, err := CacheDir(game)
				if err != nil {
					t.Fatal(err)
				}
				texts, _ = filepath.Glob(filepath.Join(dir, "area*.txt"))
			}
			if len(texts) == 0 {
				t.Skip("scripts not imported")
			}
			data := os.DirFS(dataDir)
			journals := 0
			for _, f := range texts {
				area, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f), "area"), ".txt"))
				if err != nil {
					t.Fatal(f, err)
				}
				b, err := os.ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				id, _ := strconv.Atoi(strings.Fields(string(b))[1])
				v := Versions[VersionID(id)]
				if v == nil {
					t.Fatalf("%s: version %d", f, id)
				}
				n, err := ReadNames(data, area, false, v)
				if err != nil {
					t.Fatal(f, err)
				}
				s, err := Parse(string(b), n)
				if err != nil {
					t.Errorf("%s:%v", f, err)
					continue
				}
				if err := checkOperands(s, data); err != nil {
					t.Errorf("%s:%v", f, err)
				}
				for _, in := range s.Code {
					if in.Op == OpJournal {
						journals++
					}
				}
			}
			t.Logf("%d areas, %d Journal commands", len(texts), journals)
		})
	}
}
