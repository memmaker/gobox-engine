package engine

import (
	"bufio"
	"fmt"
	"gobox/rules"
	"os"
	"strings"
	"testing"
)

// TestWhoAgainstPascal compares WHO's answer (dWho) with EclVm.EclWho
// (rules1e testdata/por-who.txt, por-re vecwho): a party of three, member
// 1..3 current, the prompt answered with member 1..3 or backed out of (0).
// The purse pick (purseWho, an extra) answers without a prompt and is not here.
func TestWhoAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-who.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	if len(e.state().Party) < 3 {
		t.Fatal("the quick start has fewer than three members")
	}
	e.state().Party = e.state().Party[:3]
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		e.selectMember(k[0] - 1)
		start := e.selected()
		pick := rules.NoChar
		if k[1] > 0 {
			pick = e.partyIDs()[k[1]-1]
		}
		e.complete(askDone{kind: dWho, who: start}, ansv{who: pick})
		if got := fmt.Sprint(e.sel + 1); got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
