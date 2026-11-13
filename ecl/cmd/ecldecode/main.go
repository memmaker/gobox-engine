// Command ecldecode decompiles a game's ECL*.DAX with no names, for
// surveying a dialect: `ecldecode VERSION DOSDIR OUTDIR` writes area<N>.txt
// and prints one line per block, its menus' exit hints and folded gates.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"gobox/asset/orig"
	"gobox/ecl"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage: ecldecode VERSION DOSDIR OUTDIR")
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil || ecl.Versions[ecl.VersionID(id)] == nil {
		log.Fatalf("no version %q", os.Args[1])
	}
	v := ecl.Versions[ecl.VersionID(id)]
	l, err := orig.NewLoader(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	out := os.Args[3]
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}
	done := map[int]bool{}
	bad := 0
	for disk := 1; disk < 9; disk++ {
		name := fmt.Sprintf("ECL%d", disk)
		if !l.Exists(name + ".DAX") {
			continue
		}
		for block := range 256 {
			b := l.BlockShared(name, block)
			if len(b) < 2 || done[block] {
				continue
			}
			done[block] = true
			names := &ecl.Names{Version: v, Vars: map[int]string{}}
			s, err := ecl.Decode(b[2:], names)
			if err != nil {
				fmt.Printf("area %d (ECL%d): DECODE %v\n", block, disk, err)
				bad++
				continue
			}
			s.DropNextGotos()
			ecl.MarkExits(s, block, v)
			folds := ecl.FoldVendorGates(s, block, names)
			for _, in := range s.Code {
				for k, a := range in.Args {
					if !a.Label && (a.Type == ecl.ArgVar || a.Type == ecl.ArgVarWord || a.Type == ecl.ArgTextVar || a.Type == ecl.ArgVarLong) && a.N >= v.Base && a.N < v.Base+len(b)-2 {
						fmt.Printf("  code area %d %s arg %d type %d addr %04X\n", block, ecl.OpName(in.Op), k, a.Type, a.N)
					}
				}
			}
			menus, hinted := 0, 0
			for _, in := range s.Code {
				if rows := ecl.MenuLabels(in); rows != nil {
					menus++
					if r, ok := in.Exit.Reply(); ok {
						hinted++
						fmt.Printf("  hint area %d %q -> %q\n", block, rows, rows[r])
					} else {
						fmt.Printf("  none area %d %q\n", block, rows)
					}
				}
			}
			for _, f := range folds {
				fmt.Printf("  fold area %d gate %d vendor %s\n", block, f.Gate, f.Vendor)
			}
			if ecl.NeedsPlainTests(s, names) {
				fmt.Printf("  plain area %d\n", block)
			}
			text, err := ecl.Write(s, names)
			if err != nil {
				fmt.Printf("area %d: WRITE %v\n", block, err)
				bad++
				continue
			}
			back, err := ecl.Parse(text, names)
			if err != nil {
				fmt.Printf("area %d: PARSE %v\n", block, err)
				bad++
			} else if d := differences(v, block, s, back); d > 0 {
				fmt.Printf("area %d: ROUNDTRIP %d commands differ (%d, %d)\n", block, d, len(s.Code), len(back.Code))
			}
			if err := os.WriteFile(filepath.Join(out, fmt.Sprintf("area%d.txt", block)), []byte(text), 0o644); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("area %d (ECL%d): %d commands, %d menus, %d hinted, %d folded\n", block, disk, len(s.Code), menus, hinted, len(folds))
		}
	}
	fmt.Printf("%d blocks, %d failed\n", len(done), bad)
}

// differences counts the commands that changed through Write and Parse,
// operands included, and prints the first few.
func differences(v *ecl.Version, block int, a, b *ecl.Script) int {
	n := max(len(a.Code), len(b.Code)) - min(len(a.Code), len(b.Code))
	for _, extra := range b.Code[min(len(a.Code), len(b.Code)):] {
		if extra.Op != ecl.OpGoto {
			fmt.Printf("  notstub area %d: %s\n", block, ecl.OpName(extra.Op))
		}
	}
	for e, at := range b.Events {
		if at >= len(a.Code) {
			at = b.Code[at].Args[0].N // the event shares a command: its stub jumps there
		}
		if at != a.Events[e] {
			fmt.Printf("  notstub area %d: event %d at %d, was %d\n", block, e, at, a.Events[e])
		}
	}
	shown := 0
	for i := range min(len(a.Code), len(b.Code)) {
		x, y := a.Code[i], b.Code[i]
		if x.Op == ecl.OpLoadMap && v.LoadMapSecondUnused && len(x.Args) == 3 && len(y.Args) == 3 {
			x.Args = []ecl.Arg{x.Args[0], y.Args[1], x.Args[2]}
		}
		if x.Op != y.Op || !slices.Equal(x.Args, y.Args) {
			n++
			if shown < 3 && x.Op != y.Op {
				shown++
				fmt.Printf("  diff area %d #%d: %s %+v | %s %+v\n", block, i, ecl.OpName(x.Op), x.Args, ecl.OpName(y.Op), y.Args)
			}
		}
	}
	return n
}
