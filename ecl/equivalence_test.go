package ecl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gobox/dax"
	"gobox/ecl/script"
)

// noText answers text variables for both readers alike.
type noText struct{}

func (noText) Get(script.Ref) int64       { return 0 }
func (noText) Set(script.Ref, int64)      {}
func (noText) Text(script.Ref) string     { return "" }
func (noText) SetText(script.Ref, string) {}
func (noText) Random(int) int             { return 0 }
func (noText) Exit(*VM)                   {}

// TestScriptEqualsByteCode: for every game and every command of its ECL
// byte code, the command number, the operands and the next command the VM
// gets from the imported text are the byte code's, with jumps and table
// references at their new places (the tables' bytes equal). Unfolded: folds
// are compared in cotab's TestGateComparedWithFold. Without the game's data
// a game is skipped.
func TestScriptEqualsByteCode(t *testing.T) {
	for _, g := range []struct {
		dir string // the game's data: dos/ is its originals
		v   *script.Version
	}{{"../data/por/sources", script.V1}, {"../data/cotab/sources", script.V2}, {"../data/sotsb/sources", script.V2}, {"../data/pod/sources", script.V3}, {"../../buckrogers-go/data", script.V4}, {"../../buckrogers2-go/data", script.V4}} {
		t.Run(g.dir, func(t *testing.T) {
			dos := filepath.Join(g.dir, "dos")
			if _, err := os.Stat(dos); err != nil {
				t.Skip("no game data")
			}
			l, err := dax.NewLoader(dos)
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

func compareBlock(t *testing.T, v *script.Version, block int, code []byte) int {
	names := &script.Names{Version: v, Vars: map[int]string{}}
	d, err := script.Decode(code, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	text, err := script.Write(d, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	s, err := script.Parse(text, names)
	if err != nil {
		t.Fatalf("area %d: %v", block, err)
	}
	_, lines, _ := v.Walk(code)
	pcs := script.SortedPCs(lines)
	index := map[int]int{} // a command's number by its address
	i := 0
	for at, pc := range pcs {
		l := lines[pc]
		if i < len(s.Code) && s.Code[i].Op == script.OpExit && l.Cmd.Op != script.OpExit {
			i++ // the text's Exit where the original runs into data
		}
		if i == len(s.Code) || s.Code[i].Op != l.Cmd.Op {
			if l.Cmd.Op == script.OpGoto && len(l.Targets) == 1 && at+1 < len(pcs) && l.Targets[0] == pcs[at+1] {
				continue // a jump to the next command is left out
			}
			t.Fatalf("area %d: %04X %s is not in the script (command %d)", block, pc, script.OpName(l.Cmd.Op), i)
		}
		index[pc] = i
		i++
	}
	n := 0
	for pc, at := range index {
		l, in := lines[pc], s.Code[at]
		n++
		if len(l.Args) != len(in.Args) {
			if in.Op == script.OpLoadMap && v.LoadMapSecondUnused {
				continue
			}
			t.Errorf("area %d %04X %s: %d operands, %d", block, pc, script.OpName(in.Op), len(l.Args), len(in.Args))
			continue
		}
		for k, a := range l.Args {
			b := in.Args[k]
			if a.Type != b.Type || a.S != b.S {
				t.Errorf("area %d %04X %s operand %d: %+v, %+v", block, pc, script.OpName(in.Op), k, a, b)
				continue
			}
			switch {
			case b.Label:
				if want, ok := index[a.N]; ok && want != b.N {
					t.Errorf("area %d %04X %s jump to %04X: command %d, want %d", block, pc, script.OpName(in.Op), a.N, b.N, want)
				}
			case a.N == b.N || a.Type == script.ArgText:
			case a.N >= v.Base && a.N-v.Base < len(code) && b.N >= v.Base && b.N-v.Base < len(s.Data):
				if code[a.N-v.Base] != s.Data[b.N-v.Base] {
					t.Errorf("area %d %04X %s table %04X moved to %04X: its byte differs", block, pc, script.OpName(in.Op), a.N, b.N)
				}
			case in.Op == script.OpLoadMap && v.LoadMapSecondUnused && k == 1:
			default:
				t.Errorf("area %d %04X %s operand %d: %+v, %+v", block, pc, script.OpName(in.Op), k, a, b)
			}
		}
	}
	return n
}
