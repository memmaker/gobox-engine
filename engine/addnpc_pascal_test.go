package engine

import (
	"bufio"
	"fmt"
	"gobox/rules1e"
	"os"
	"strings"
	"testing"

	"gobox/ecl/script"
	"gobox/rules"
)

// TestAddNPCAgainstPascal compares ADD NPC (opAddNPC) with EclVm.EclAddNpc
// (rules1e testdata/por-addnpc.txt, por-re vecnpc): a party of 1, 7 or 8
// with member 1 selected, monster 5 or 24, morale 100. The Go port always
// has the monster's record, so the cases with none are not here.
func TestAddNPCAgainstPascal(t *testing.T) {
	f, err := os.Open("../rules1e/testdata/por-addnpc.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	base := quickStarted(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, want, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		if k[1] == 0 {
			continue
		}
		e := quickStarted(t)
		for d := rules.Set(1); d <= 8 && refDAX(e).BlockShared(e.set.File("MON")+"CHA", k[2]) == nil; d++ {
			e.set = d // a set that has monster k[2]
		}
		members := e.state().Party
		for len(members) < k[0] {
			c := *base.state().Party[0]
			members = append(members, &c)
		}
		e.state().Party = members[:k[0]]
		e.setWorldWord(rules.WordPartyCount, k[0])
		e.selectMember(0)
		first := e.member(e.selected())
		e.vm.Script = &script.Script{Code: []script.Instr{{Op: 0x36, Args: []script.Arg{{Type: 0, N: k[2]}, {Type: 0, N: 100}}}}}
		e.vm.PC = 0
		e.vm.Fetch()
		e.opAddNPC()
		c := e.member(e.selected())
		got := fmt.Sprint(e.state().PartyCount(), rules1e.B2i(c == first), c.Control, c.Side)
		if got != want {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}
