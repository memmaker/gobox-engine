// Command ecldecode decompiles a game's ECL*.DAX with no names, for
// surveying a dialect: `ecldecode VERSION DOSDIR OUTDIR` writes area<N>.txt
// and prints one line per block, its menus' exit hints and folded gates.
package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strconv"

	"gobox/dax"
	"gobox/ecl/script"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage: ecldecode VERSION DOSDIR OUTDIR")
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil || script.Versions[script.VersionID(id)] == nil {
		log.Fatalf("no version %q", os.Args[1])
	}
	v := script.Versions[script.VersionID(id)]
	l, err := dax.NewLoader(os.Args[2])
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
			names := &script.Names{Version: v, Vars: maps.Clone(v.Vars)}
			s, err := script.Decode(b[2:], names)
			if err != nil {
				fmt.Printf("area %d (ECL%d): DECODE %v\n", block, disk, err)
				bad++
				continue
			}
			s.DropNextGotos()
			script.MarkExits(s, block, v)
			folds := script.FoldVendorGates(s, block, names)
			for _, in := range s.Code {
				for k, a := range in.Args {
					if !a.Label && (a.Type == script.ArgVar || a.Type == script.ArgVarWord || a.Type == script.ArgTextVar || a.Type == script.ArgVarLong) && a.N >= v.Base && a.N < v.Base+len(b)-2 {
						fmt.Printf("  code area %d %s arg %d type %d addr %04X\n", block, script.OpName(in.Op), k, a.Type, a.N)
					}
				}
			}
			menus, hinted := 0, 0
			for _, in := range s.Code {
				if rows := script.MenuLabels(in); rows != nil {
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
			text, err := script.Write(s, names)
			if err != nil {
				fmt.Printf("area %d: WRITE %v\n", block, err)
				bad++
				continue
			}
			back, err := script.Parse(text, names)
			if err != nil {
				fmt.Printf("area %d: PARSE %v\n", block, err)
				bad++
			} else if d := differences(s, back); d > 0 {
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

// differences counts the commands whose byte-code command or operand count
// changed through Write and Parse.
func differences(a, b *script.Script) int {
	n := max(len(a.Code), len(b.Code)) - min(len(a.Code), len(b.Code))
	for i := range min(len(a.Code), len(b.Code)) {
		if a.Code[i].Op != b.Code[i].Op || len(a.Code[i].Args) != len(b.Code[i].Args) {
			n++
		}
	}
	return n
}
