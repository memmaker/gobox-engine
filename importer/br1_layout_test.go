package importer

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl"
)

// Buck Rogers 1's full script layout (br1/scripts/layout/global.txt), proven against the byte code of every
// ECL block of the originals. Not wired into the import yet (WriteBRLayout still writes the partial one).

// The two scratch zones the original clears at each script load (types.pas entryBytes, newGameBytes): the
// only raw byte buffers of the layout.
const (
	entryBytesFrom, entryBytesLen = 0x4C00, 32
	newGameBytesFrom, newGameLen  = 0x7F79, 10
)

func readBR1Layout(t *testing.T) *ecl.Layout {
	t.Helper()
	l, err := ecl.ReadLayout(embeddedScripts("br1"), 0)
	if err != nil || l == nil {
		t.Fatal("br1 layout:", l, err)
	}
	if l.Partial {
		t.Fatal("br1 layout is partial")
	}
	return l
}

// br1Blocks are the ECL blocks of the originals by block number (the first disk that has one, as ecl.Import).
func br1Blocks(t *testing.T) (*orig.Loader, string, map[int][]byte) {
	t.Helper()
	dos := filepath.Join("..", "data", "br1", "sources", "dos")
	l, err := orig.NewLoader(dos)
	if err != nil || !l.Exists("ITEMS") {
		t.Skip("no originals:", err)
	}
	src := orig.DAXSource(l)
	blocks := map[int][]byte{}
	for disk := 1; disk < 9; disk++ {
		for block := range 256 {
			if b := src.Block(disk, block); len(b) >= 2 && blocks[block] == nil {
				blocks[block] = b[2:]
			}
		}
	}
	if len(blocks) == 0 {
		t.Fatal("no ECL blocks")
	}
	return l, dos, blocks
}

// operandWidth is the width of a variable operand type; ok false: not a variable.
func operandWidth(typ byte) (ecl.Width, int, bool) {
	switch typ {
	case ecl.ArgVar:
		return ecl.WidthByte, 1, true
	case ecl.ArgVarWord:
		return ecl.WidthWord, 2, true
	case ecl.ArgVarLong:
		return ecl.WidthDword, 4, true
	case ecl.ArgTextVar:
		return ecl.WidthText, 1, true
	}
	return 0, 0, false
}

func TestBR1LayoutShape(t *testing.T) {
	lay := readBR1Layout(t) // ParseLayout's check: no overlap, no name twice
	want := map[string][2]int{"EntryBytes": {entryBytesFrom, entryBytesLen}, "NewGameBytes": {newGameBytesFrom, newGameLen}}
	for _, f := range lay.Fields {
		if f.Width != ecl.WidthBytes {
			continue
		}
		w, ok := want[f.Name]
		if !ok || f.From != w[0] || f.Len != w[1] || f.Owner != ecl.OwnerLocal {
			t.Errorf("byte buffer %s %04X len %d owner %d: want exactly EntryBytes, NewGameBytes", f.Name, f.From, f.Len, f.Owner)
		}
		delete(want, f.Name)
	}
	for n := range want {
		t.Errorf("no byte buffer %s", n)
	}
	for name, addr := range map[string]int{"DivideRemainder": 0x7F3F, "LoopCounter": 0x4CF6, "LoopLimit": 0x4CF7} {
		r, err := lay.Named(name)
		if err != nil || lay.Address(r) != addr {
			t.Errorf("%s: %v at %04X, want %04X", name, err, lay.Address(r), addr)
		}
	}
	if r, _ := lay.Named("DivideRemainder"); lay.Field(r).Width != ecl.WidthDword {
		t.Error("DivideRemainder is not a dword")
	}
}

// TestBR1LayoutResolvesEveryOperand decodes every block of the originals and resolves every variable operand
// in the full layout (with the block's own const tables), and checks each operand against its field.
func TestBR1LayoutResolvesEveryOperand(t *testing.T) {
	lay := readBR1Layout(t)
	_, _, blocks := br1Blocks(t)
	data := embeddedScripts("br1")
	var unresolved, bad []string
	widths := map[int]map[ecl.Width]bool{} // per typed address, the widths the scripts use
	operands := 0
	for _, area := range slices.Sorted(maps.Keys(blocks)) {
		names, err := ecl.ReadNames(data, area, true, ecl.V4)
		if err != nil {
			t.Fatal(err)
		}
		s, err := ecl.Decode(blocks[area], names)
		if err != nil {
			t.Fatalf("area %d: %v", area, err)
		}
		full, err := lay.With(ecl.ConstLayout(s, ecl.V4, lay))
		if err != nil {
			t.Fatalf("area %d: %v", area, err)
		}
		if err := ecl.ResolveRefs(s, full); err != nil {
			t.Errorf("area %d: ResolveRefs: %v", area, err)
		}
		for i, in := range s.Code {
			for k, a := range in.Args {
				w, n, ok := operandWidth(a.Type)
				if a.Label || !ok || in.Op == ecl.OpCall {
					continue
				}
				operands++
				where := fmt.Sprintf("area %d instruction %d (%s) operand %d: %04X", area, i, ecl.OpName(in.Op), k, a.N)
				r, err := full.Resolve(a.N)
				if err != nil {
					unresolved = append(unresolved, where)
					continue
				}
				f := full.Field(r)
				switch {
				case f.Width == ecl.WidthBytes:
					if r.Index+n > f.Len { // a word at the last byte runs past the buffer
						bad = append(bad, where+" runs past "+f.Name)
					}
				case f.Owner == ecl.OwnerConst:
				default:
					if widths[a.N] == nil {
						widths[a.N] = map[ecl.Width]bool{}
					}
					widths[a.N][w] = true
					if f.Width != w && f.Name != "DivideRemainder" { // the engine writes DivideRemainder as a dword; scripts read its low byte
						bad = append(bad, where+fmt.Sprintf(" is a %d-wide operand of %s (width %d)", n, f.Name, f.Width))
					}
					for c := a.N + 1; c < a.N+n; c++ { // the cells the operand covers belong to no other field
						if o, err := full.Resolve(c); err == nil && full.Field(o) != f {
							bad = append(bad, where+" overlaps "+full.Field(o).Name)
						}
					}
				}
			}
		}
	}
	for a, ws := range widths {
		if len(ws) > 1 {
			bad = append(bad, fmt.Sprintf("%04X is used at %d widths", a, len(ws)))
		}
	}
	slices.Sort(unresolved)
	slices.Sort(bad)
	if len(unresolved) > 0 {
		t.Errorf("%d operands in no field:\n%s", len(unresolved), strings.Join(unresolved, "\n"))
	}
	if len(bad) > 0 {
		t.Errorf("%d operands that do not fit their field:\n%s", len(bad), strings.Join(bad, "\n"))
	}
	t.Logf("%d blocks, %d variable operands resolved", len(blocks), operands)
}

// TestBR1LayoutImportRoundTrip imports the originals' scripts with the full layout in place of the partial
// one: the importer can switch to it later. Every block's text then loads again (parse and resolve).
func TestBR1LayoutImportRoundTrip(t *testing.T) {
	l, _, blocks := br1Blocks(t)
	dir := t.TempDir()
	b, err := fs.ReadFile(embeddedScripts("br1"), ecl.GlobalLayout)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, ecl.GlobalLayout)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ecl.Import(ecl.V4, orig.DAXSource(l), os.DirFS(dir), dir, ecl.FoldGates, nil); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "area*.txt"))
	if len(files) != len(blocks) {
		t.Errorf("%d area files, %d blocks", len(files), len(blocks))
	}
	lib := &ecl.Library{Version: ecl.V4, Dir: dir, Data: os.DirFS(dir), Fold: ecl.FoldGates}
	for _, area := range slices.Sorted(maps.Keys(blocks)) {
		if _, err := lib.Area(area); err != nil {
			t.Errorf("area %d: %v", area, err)
		}
	}
}
