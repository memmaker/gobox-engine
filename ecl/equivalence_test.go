package ecl_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl"
)

// TestScriptEqualsByteCode: for every game and every command of its ECL
// byte code, the command number, the operands and the next command the VM
// gets from the imported text are the byte code's, with jumps and table
// references at their new places (the tables' bytes equal). Unfolded: folds
// are compared in cotab's TestGateComparedWithFold. Without the game's data
// a game is skipped.
func TestScriptEqualsByteCode(t *testing.T) {
	for _, g := range []struct {
		dir string // the game's data: dos/ is its originals
		v   *ecl.Version
	}{{"../data/por/sources", ecl.V1}, {"../data/cotab/sources", ecl.V2}, {"../data/sotsb/sources", ecl.V2}, {"../data/pod/sources", ecl.V3}, {"../../buckrogers-go/data", ecl.V4}, {"../../buckrogers2-go/data", ecl.V4}} {
		t.Run(g.dir, func(t *testing.T) {
			dos := filepath.Join(g.dir, "dos")
			if _, err := os.Stat(dos); err != nil {
				t.Skip("no game data")
			}
			l, err := orig.NewLoader(dos)
			if err != nil {
				t.Fatal(err)
			}
			blocks, commands := 0, 0
			for disk := 1; disk < 9; disk++ {
				name := fmt.Sprintf("ECL%d", disk)
				if !l.Exists(name + ".DAX") {
					continue
				}
				for block := range 256 {
					b := l.BlockShared(name, block)
					if len(b) < 2 {
						continue
					}
					blocks++
					commands += compareBlock(t, g.v, block, b[2:])
				}
			}
			t.Logf("%d blocks, %d commands", blocks, commands)
		})
	}
}

func compareBlock(t *testing.T, v *ecl.Version, block int, code []byte) int {
	names := &ecl.Names{Version: v, Vars: map[int]string{}}
	d, err := ecl.Decode(code, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	text, err := ecl.Write(d, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	s, err := ecl.Parse(text, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	_, lines, _ := v.Walk(code)
	pcs := ecl.SortedPCs(lines)
	index := map[int]int{} // a command's number by its address
	i := 0
	for at, pc := range pcs {
		l := lines[pc]
		if i < len(s.Code) && s.Code[i].Op == ecl.OpExit && l.Cmd.Op != ecl.OpExit {
			i++ // the text's Exit where the original runs into data
		}
		if i == len(s.Code) || s.Code[i].Op != l.Cmd.Op {
			if l.Cmd.Op == ecl.OpGoto && len(l.Targets) == 1 && at+1 < len(pcs) && l.Targets[0] == pcs[at+1] {
				continue // a jump to the next command is left out
			}
			t.Fatalf("area %d: %04X %s is not in the script (command %d)", block, pc, ecl.OpName(l.Cmd.Op), i)
		}
		index[pc] = i
		i++
	}
	n := 0
	for pc, at := range index {
		l, in := lines[pc], s.Code[at]
		n++
		if len(l.Args) != len(in.Args) {
			if in.Op == ecl.OpLoadMap && v.LoadMapSecondUnused {
				continue
			}
			t.Errorf("area %d %04X %s: %d operands, %d", block, pc, ecl.OpName(in.Op), len(l.Args), len(in.Args))
			continue
		}
		for k, a := range l.Args {
			b := in.Args[k]
			if a.Type != b.Type || a.S != b.S {
				t.Errorf("area %d %04X %s operand %d: %+v, %+v", block, pc, ecl.OpName(in.Op), k, a, b)
				continue
			}
			switch {
			case b.Label:
				if want, ok := index[a.N]; ok && want != b.N {
					t.Errorf("area %d %04X %s jump to %04X: command %d, want %d", block, pc, ecl.OpName(in.Op), a.N, b.N, want)
				}
			case a.N == b.N || a.Type == ecl.ArgText:
			case a.N >= v.Base && a.N-v.Base < len(code) && b.N >= v.Base && b.N-v.Base < len(s.Data):
				if code[a.N-v.Base] != s.Data[b.N-v.Base] {
					t.Errorf("area %d %04X %s table %04X moved to %04X: its byte differs", block, pc, ecl.OpName(in.Op), a.N, b.N)
				}
			case in.Op == ecl.OpLoadMap && v.LoadMapSecondUnused && k == 1:
			default:
				t.Errorf("area %d %04X %s operand %d: %+v, %+v", block, pc, ecl.OpName(in.Op), k, a, b)
			}
		}
	}
	return n
}
