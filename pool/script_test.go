package pool

import (
	"fmt"
	"gobox/ecl"
	"gobox/ecl/script"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobox/dax"
	"por/importsrc"
)

// TestScriptRoundTrip: every area's byte code, written as a script and
// parsed back, is the same Script (cmd/import's output parses to the
// byte code's commands).
func TestScriptRoundTrip(t *testing.T) {
	l, err := dax.NewLoader("../../por-go/data/dos")
	if err != nil {
		t.Skip(err)
	}
	dir := "../../por-go/data/assets/scripts"
	out := os.Getenv("POR_SCRIPTS_OUT") // write the scripts there too
	for set := 1; set <= 8; set++ {
		for block := 0; block < 40; block++ {
			b := l.BlockShared(fmt.Sprintf("ECL%d", set), block)
			if len(b) < 2 {
				continue
			}
			names, err := script.ReadNames(os.DirFS(dir), block, true, script.V1)
			if err != nil {
				t.Fatal(err)
			}
			if names.Set != set {
				continue
			}
			want, err := script.Decode(b[2:], names)
			if err != nil {
				t.Fatalf("area %d: decode: %v", block, err)
			}
			want.DropNextGotos() // the text's goto before a label is no jump (a few originals have one: it does nothing)
			script.MarkExits(want, block, script.V1)
			if f := script.FoldVendorGates(want, block, names); len(f) > 0 {
				t.Logf("area %d folded %+v", block, f)
			}
			text, err := script.Write(want, names)
			if err != nil {
				t.Errorf("area %d: write: %v", block, err)
				continue
			}
			if out != "" {
				os.WriteFile(filepath.Join(out, fmt.Sprintf("area%d.txt", block)), []byte(text), 0o644)
			}
			got, err := script.Parse(text, names)
			if err != nil {
				t.Errorf("area %d: parse: %v", block, err)
				continue
			}
			for i := range got.Code {
				got.Code[i].Line, got.Code[i].Comment = 0, ""
				plainLoadMap(&got.Code[i])
			}
			for i := range want.Code {
				want.Code[i].Comment = ""
				plainLoadMap(&want.Code[i])
			}
			if out != "" && len(got.Code) != len(want.Code) {
				back, _ := script.Write(got, names)
				os.WriteFile(filepath.Join(out, fmt.Sprintf("area%d.got.txt", block)), []byte(back), 0o644)
			}
			for i := range want.Code {
				if i >= len(got.Code) || !reflect.DeepEqual(want.Code[i], got.Code[i]) {
					g := "none"
					if i < len(got.Code) {
						g = fmt.Sprint(got.Code[i])
					}
					t.Errorf("area %d: command %d (%s): want %v, got %s", block, i, where(want, i), want.Code[i], g)
					break
				}
			}
			if len(got.Code) != len(want.Code) {
				t.Errorf("area %d: %d commands, want %d", block, len(got.Code), len(want.Code))
			}
			if got.Events != want.Events {
				t.Errorf("area %d: events %v, want %v", block, got.Events, want.Events)
			}
			if !reflect.DeepEqual(namedLabels(got), namedLabels(want)) {
				t.Errorf("area %d: labels differ", block)
			}
			if !reflect.DeepEqual(got.Tables, want.Tables) || string(got.Data) != string(want.Data) {
				t.Errorf("area %d: tables differ", block)
			}
		}
	}
}

// plainLoadMap drops what the original never reads from a LoadMap: its
// second operand, and which value not 255 the third has (any loads the
// icons; the script writes it as one name).
func plainLoadMap(in *script.Instr) {
	if in.Op == script.OpLoadMap && len(in.Args) == 3 {
		in.Args[1] = script.Arg{}
		if in.Args[2] != (script.Arg{N: 0xFF}) {
			in.Args[2] = script.Arg{N: 0x7F}
		}
	}
}

// namedLabels are s's labels but the loops' (for X = A to B {} names
// its own).
func namedLabels(s *script.Script) map[string]int {
	out := map[string]int{}
	for l, i := range s.Labels {
		loop := false
		for g, in := range s.Code {
			if in.Op == script.OpGoto && in.Args[0].N == i && g >= 3 && script.IsIf(s.Code[g-1].Op) && s.Code[g-2].Op == script.OpCompare && s.Code[g-3].Op == script.OpAdd {
				loop = true
			}
		}
		if !loop {
			out[l] = i
		}
	}
	return out
}

func where(s *script.Script, i int) string {
	for j := i; j >= 0; j-- {
		if l := s.At[j]; l != "" {
			return fmt.Sprintf("%s+%d", l, i-j)
		}
	}
	return "?"
}

// TestParseForms: the forms the writer makes, by hand.
func TestParseForms(t *testing.T) {
	names, err := script.ReadNames(os.DirFS("../../por-go/data/assets/scripts"), 0, false, script.V1)
	if err != nil {
		t.Fatal(err)
	}
	names.Values = []script.Value{
		{Kind: "Facing", Set: 0, N: 2, Name: "South"}, {Kind: "CastleFlags", Set: 0, N: 1, Name: "Disguised"}, {Kind: "CastleFlags", Set: 0, N: 2, Name: "HallSeen"},
		{Kind: "Monster", Set: 0, N: 3, Name: "Kobold"}, {Kind: "Icon", Set: 0, N: 5, Name: "KoboldIcon"}, {Kind: "Item", Set: 0, N: 7, Name: "LongSword"},
		{Kind: "Map", Set: 0, N: 255, Name: "Keep"}, {Kind: "MapIcons", Set: 0, N: 127, Name: "Overland"},
	}
	names.Vars[0x4A65] = "CastleFlags"
	for i := 1; i <= 10; i++ {
		names.Vars[0x6E78+i] = fmt.Sprintf("Scratch%d", i)
	}
	vars := map[string]int{}
	for a, n := range names.Vars {
		vars[n] = a
	}
	v := func(n string) script.Arg { return script.Arg{Type: 1, N: vars[n]} }
	src := `
on load { exit }
on move { exit }
on camp { exit }
on rest { exit }
on step {
    if Facing == South { goto Ahead }
    if Facing < South {
        Sky = 3
        Sky += 1
    } else if Facing > South {
        Sky = 4
    } else {
        CastleFlags |= Disguised|HallSeen
    }
    if CastleFlags has Disguised { Done() }
    if Carried(LongSword) { Text("YES") }
    if Scratch1 = Scratch2 & 7; Scratch1 != 0 { Print(Scratch1) }
    if Done() { exit }
    if not Done() { Sky = 5 }
    if Count() >= 2 { exit }
    if Ask() == "NO" { exit }
    Compare(Facing, South)
    Sky = 6
    if (==) { exit }
    LoadMap(Keep, icons=Overland)
    for Scratch3 = 0 to 4 { Scratch4 = Kinds[Scratch3] }
    LoadMonster(Kobold, count=8, icon=KoboldIcon)
    goto [Ahead, Later][Scratch1]
}
Ahead:
    Treasure(gold=5, items=Random(2))
    exit
Later:
    exit
def Done() {
    return Facing == Scratch7
}
def Count() {
    Scratch5 = Random(3)
    return Scratch5
}
def Ask() {
    Scratch6 = Menu("YES", "NO")
    return Scratch6
}
# tables
Kinds = [1, 2, 3]
`
	s, err := script.Parse(script.V1.Header()+src, names)
	if err != nil {
		t.Fatal(err)
	}
	ahead, later, kinds := s.Labels["Ahead"], s.Labels["Later"], script.Arg{Type: 1, N: eclBase}
	done, count, ask := s.Labels["Done"], s.Labels["Count"], s.Labels["Ask"]
	call := func(l int) script.Instr {
		return script.Instr{Op: script.OpGosub, Args: []script.Arg{{Type: 1, N: l, Label: true}}}
	}
	want := []script.Instr{
		{Op: script.OpExit}, {Op: script.OpExit}, {Op: script.OpExit}, {Op: script.OpExit},
		{Op: script.OpCompare, Args: []script.Arg{v("Facing"), {N: 2}}}, {Op: script.OpIfEq}, {Op: script.OpGoto, Args: []script.Arg{{Type: 1, N: ahead, Label: true}}},
		{Op: script.OpCompare, Args: []script.Arg{v("Facing"), {N: 2}}},
		{Op: script.OpIfLt}, {Op: script.OpSave, Args: []script.Arg{{N: 3}, v("Sky")}}, {Op: script.OpIfLt}, {Op: script.OpAdd, Args: []script.Arg{{N: 1}, v("Sky"), v("Sky")}},
		{Op: script.OpIfGt}, {Op: script.OpSave, Args: []script.Arg{{N: 4}, v("Sky")}},
		{Op: script.OpIfLe}, {Op: script.OpOr, Args: []script.Arg{v("CastleFlags"), {N: 3}, v("CastleFlags")}},
		{Op: script.OpAnd, Args: []script.Arg{v("CastleFlags"), {N: 1}, v("Scratch10")}}, {Op: script.OpIfNe}, call(done),
		{Op: script.OpFindItem, Args: []script.Arg{{N: 7}}}, {Op: script.OpIfEq}, {Op: script.OpPrintClear, Args: []script.Arg{{Type: 0x80, N: 3, S: "YES"}}},
		{Op: script.OpAnd, Args: []script.Arg{v("Scratch2"), {N: 7}, v("Scratch1")}}, {Op: script.OpIfNe}, {Op: script.OpPrint, Args: []script.Arg{v("Scratch1")}},
		call(done), {Op: script.OpIfEq}, {Op: script.OpExit},
		call(done), {Op: script.OpIfNe}, {Op: script.OpSave, Args: []script.Arg{{N: 5}, v("Sky")}},
		call(count), {Op: script.OpIfGe}, {Op: script.OpExit},
		call(ask), {Op: script.OpIfEq}, {Op: script.OpExit},
		{Op: script.OpCompare, Args: []script.Arg{v("Facing"), {N: 2}}}, {Op: script.OpSave, Args: []script.Arg{{N: 6}, v("Sky")}}, {Op: script.OpIfEq}, {Op: script.OpExit},
		{Op: script.OpLoadMap, Args: []script.Arg{{N: 255}, {N: 255}, {N: 127}}},
		{Op: script.OpSave, Args: []script.Arg{{N: 0}, v("Scratch3")}},
		{Op: script.OpGetTable, Args: []script.Arg{kinds, v("Scratch3"), v("Scratch4")}},
		{Op: script.OpAdd, Args: []script.Arg{{N: 1}, v("Scratch3"), v("Scratch3")}}, {Op: script.OpCompare, Args: []script.Arg{v("Scratch3"), {N: 4}}}, {Op: script.OpIfLe},
		{Op: script.OpGoto, Args: []script.Arg{{Type: 1, N: 43, Label: true}}},
		{Op: 0x0B, Args: []script.Arg{{N: 3}, {N: 8}, {N: 5}}},
		{Op: script.OpOnGoto, Args: []script.Arg{v("Scratch1"), {N: 2}, {Type: 1, N: ahead, Label: true}, {Type: 1, N: later, Label: true}}},
		{Op: script.OpTreasure, Args: []script.Arg{{}, {}, {}, {N: 5}, {}, {}, {}, {N: 0x82}}}, {Op: script.OpExit},
		{Op: script.OpExit},
		{Op: script.OpCompare, Args: []script.Arg{v("Facing"), v("Scratch7")}}, {Op: script.OpReturn},
		{Op: script.OpRandom, Args: []script.Arg{{N: 3}, v("Scratch5")}}, {Op: script.OpCompare, Args: []script.Arg{v("Scratch5"), {N: 2}}}, {Op: script.OpReturn},
		{Op: script.OpHMenu, Args: []script.Arg{v("Scratch6"), {N: 2}, {Type: 0x80, N: 3, S: "YES"}, {Type: 0x80, N: 2, S: "NO"}}},
		{Op: script.OpCompare, Args: []script.Arg{v("Scratch6"), {N: 1}}}, {Op: script.OpReturn},
	}
	for i := range s.Code {
		s.Code[i].Line = 0
	}
	for i, w := range want {
		if i >= len(s.Code) || !reflect.DeepEqual(w, s.Code[i]) {
			g := "none"
			if i < len(s.Code) {
				g = fmt.Sprint(s.Code[i])
			}
			t.Fatalf("command %d: want %v, got %s", i, w, g)
		}
	}
	if len(s.Code) != len(want) || string(s.Data) != "\x01\x02\x03" {
		t.Fatalf("%d commands, want %d; data %v", len(s.Code), len(want), s.Data)
	}
	text, err := script.Write(s, names)
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range []string{"if Facing == South { goto Ahead }", "} else if Facing > South {", "} else {", "CastleFlags |= Disguised|HallSeen",
		"if CastleFlags has Disguised { Done() }", "if Carried(LongSword) { Text(\"YES\") }", "if Scratch1 = Scratch2 & 7; Scratch1 != 0 { Print(Scratch1) }",
		"if Done() { exit }", "if not Done() { Sky = 5 }", "if Count() >= 2 { exit }", "if Ask() == \"NO\" { exit }",
		"Compare(Facing, South)", "if (==) { exit } # Facing == South", "LoadMap(Keep, icons=Overland)",
		"for Scratch3 = 0 to 4 {", "LoadMonster(Kobold, count=8, icon=KoboldIcon)", "goto [Ahead, Later][Scratch1]",
		"Treasure(gold=5, items=Random(2))", "def Done() {", "return Facing == Scratch7", "return Scratch5", "return Scratch6", "Kinds = [1, 2, 3]"} {
		if !strings.Contains(text, form) {
			t.Errorf("the writer lost %q:\n%s", form, text)
		}
	}
}

// TestImportAppliesFixes: the game data decompiles and every special case (scriptFixes) applies.
func TestImportAppliesFixes(t *testing.T) {
	const scripts = "../../por-go/data/assets/scripts"
	l, err := dax.NewLoader("../../por-go/data/dos")
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	if _, err := script.Import(script.V1, importsrc.ImportSource(l), os.DirFS(scripts), dir, script.FoldGates, ScriptFixes); err != nil {
		t.Fatal(err)
	}
	if files, _ := filepath.Glob(filepath.Join(dir, "area*.txt")); len(files) == 0 {
		t.Fatal("no scripts")
	}
}

// Every monster a script names by number is in the MON file of its area's
// disk: LoadMonster aborts the game on a missing one (the set is the ECL file
// the area came from). Bug class: a script fix naming a monster of another disk.
func TestScriptMonstersInTheirSet(t *testing.T) {
	e := testEngine(t)
	e.Init()
	const sets, opLoadMonster = 8, 0x0B
	for area := 0; area <= 60; area++ {
		set := 0
		for s := 1; s <= sets; s++ {
			if refDAX(e).BlockShared(fmt.Sprintf("ECL%d", s), area) != nil {
				set = s
			}
		}
		if set == 0 {
			continue
		}
		e.LoadArea(area)
		for _, in := range e.vm.Script.Code {
			if in.Op != opLoadMonster || in.Args[0].Type != ecl.ArgImmediate {
				continue
			}
			if refDAX(e).BlockShared(fmt.Sprintf("MON%dCHA", set), in.Args[0].N) == nil {
				t.Errorf("area %d line %d: monster %d is not in MON%dCHA", area, in.Line, in.Args[0].N, set)
			}
		}
	}
}
