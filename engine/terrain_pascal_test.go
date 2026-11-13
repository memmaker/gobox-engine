package engine

import (
	"bufio"
	"os"
	"gobox/rules1e"
	"strconv"
	"strings"
	"testing"
)

// TestTerrainKindsAgainstPascal compares terrainFlags with CombatSt.TerrainClassOf
// (rules1e testdata/por-terrain.txt, por-re vecterr) for every value, before
// and after Yarash's mission.
func TestTerrainKindsAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-terrain.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := testEngine(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		var done, terr, want int
		key, val, _ := strings.Cut(line, " = ")
		k := strings.Fields(key)
		done, _ = strconv.Atoi(k[1])
		terr, _ = strconv.Atoi(k[2])
		want, _ = strconv.Atoi(val)
		e.state().Area.Boss = uint16(done)
		e.state().Terrain = rules1e.Terrain(byte(terr))
		if got := int(e.state().TerrainFlags()); got != want {
			t.Errorf("%s: pascal %d, ours %d", key, want, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestEveryWildSquareHasAKind: no wilderness square falls through the kinds.
func TestEveryWildSquareHasAKind(t *testing.T) {
	for y, row := range rules1e.TerrainMap {
		for x, v := range row {
			covered := false
			for _, k := range rules1e.TerrainKinds {
				if v >= k.Lo && v <= k.Hi {
					covered = true
					break
				}
			}
			if !covered {
				t.Errorf("square %d,%d: value %d has no kind", y, x, v)
			}
		}
	}
}
