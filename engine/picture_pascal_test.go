package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"gobox/ecl"
)

// TestPictureAgainstPascal compares opPicture's flags with EclVm.EclPicture
// (rules1e testdata/por-picture.txt, por-re vecpic): eclGraphics is es.picShown,
// sprShown es.spriteLoaded, picDrawn es.picAnimating, monsterFlags
// es.encounterShown, locTypePrev and locType prevViewMode and viewMode.
func TestPictureAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-picture.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := testEngine(t)
	e.state().Area.In3D = 1 // no window blanking: drawing only
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Op: 0x0F, Args: []ecl.Arg{{Type: 0, N: k[0]}}}}}
		e.vm.PC = 0
		e.vm.Fetch()
		e.prevViewMode, e.viewMode = k[1], k[2]
		e.es.picShown, e.es.spriteLoaded, e.es.picAnimating = byte(k[3]), byte(k[4]), 0
		e.es.encounterShown = [2]byte{1, 0}
		e.opPicture()
		got := fmt.Sprintf("%d %d %d %d %d", e.es.picShown, e.es.spriteLoaded, e.es.picAnimating, e.es.encounterShown[0], e.es.encounterShown[1])
		if got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
