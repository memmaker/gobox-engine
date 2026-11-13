package engine

import (
	"encoding/binary"
	"fmt"
	"gobox/rules1e"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl/script"
)

// TestLayoutCoversScripts: every variable operand of every area's script
// (imported afresh, its special cases applied) resolves (script.ResolveRefs) in the
// game's layout, the area's over it and the area's constant tables.
func TestLayoutCoversScripts(t *testing.T) {
	const scripts = "../data/por/sources/assets/scripts"
	if _, err := orig.NewLoader("../data/por/sources/dos"); err != nil {
		t.Skip(err)
	}
	dir := scripts // imported, special cases applied, by the TestMain import (por-go cmd/importstale)
	files, err := filepath.Glob(filepath.Join(dir, "area*.txt"))
	if err != nil || len(files) == 0 {
		t.Fatal("no scripts", err)
	}
	for _, path := range files {
		var area int
		if _, err := fmt.Sscanf(filepath.Base(path), "area%d.txt", &area); err != nil {
			t.Fatal(err)
		}
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		names, err := script.ReadNames(os.DirFS(scripts), area, true, script.V1)
		if err != nil {
			t.Fatal(err)
		}
		s, err := script.Parse(string(text), names)
		if err != nil {
			t.Fatalf("area %d: parse: %v", area, err)
		}
		s.Area = area
		lay, err := script.ReadLayout(os.DirFS(scripts), area)
		if err != nil || lay == nil {
			t.Fatalf("area %d: %v %v", area, lay, err)
		}
		full, err := lay.With(script.ConstLayout(s, script.V1))
		if err != nil {
			t.Fatalf("area %d: %v", area, err)
		}
		if err := script.ResolveRefs(s, full); err != nil {
			t.Error(err)
		}
		for i, in := range s.Code {
			if in.Op == script.OpNamed("Spell") { // the engine has no Spell (Felix d3: area 14 casts by castKnock)
				t.Errorf("area %d instruction %d: Spell", area, i)
			}
		}
	}
}

// resolved is s with its operands resolved in area's layout (and s's
// tables), as Library.Area gives a script: for a test's own script.
func resolved(t *testing.T, s *script.Script, area int) *script.Script {
	t.Helper()
	return resolvedOver(t, s, area, "")
}

// resolvedOver is resolved with the test's own fields (layout text) over
// area's.
func resolvedOver(t *testing.T, s *script.Script, area int, over string) *script.Script {
	t.Helper()
	l, err := script.ReadLayout(os.DirFS("../data/por/sources/assets/scripts"), area)
	if err != nil {
		t.Fatal(err)
	}
	o, err := script.ParseLayout(strings.NewReader(over))
	if err != nil {
		t.Fatal(err)
	}
	if l, err = l.With(o); err != nil {
		t.Fatal(err)
	}
	full, err := l.With(script.ConstLayout(s, script.V1))
	if err != nil {
		t.Fatal(err)
	}
	if err := script.ResolveRefs(s, full); err != nil {
		t.Fatal(err)
	}
	return s
}

// refAt is the Ref of script address a in the running script's layout
// (area 0's when none is loaded): the Pascal reference tests give
// addresses. ok false: no field is there, so no script can reach it.
func refAt(t *testing.T, e *Engine, a int) (script.Ref, bool) {
	t.Helper()
	withLayout(t, e)
	r, err := e.vm.Script.Layout.Resolve(a)
	return r, err == nil
}

// testField is the field called name, in area 0's layout when no script is loaded.
func testField(t *testing.T, e *Engine, name string) script.Ref {
	t.Helper()
	withLayout(t, e)
	return e.field(name)
}

// withLayout gives e area 0's layout when it runs no script.
func withLayout(t *testing.T, e *Engine) {
	t.Helper()
	if e.vm.Script.Layout == nil {
		l, err := script.ReadLayout(os.DirFS(e.scriptDir()), 0)
		if err != nil {
			t.Fatal(err)
		}
		e.vm.Script = &script.Script{Layout: l}
	}
}

// TestSavedFields: the fields kept by name (scriptvars.go) stand for one
// thing each. Every area's layout names a game field with one place, no
// two names share a word but a Replaces field, the locals are exactly the
// words a fresh area clears, and every fact has its engine
// meaning. Found by: two areas naming PatrolsBeaten at 4A34 and
// 4A80, which the address had kept apart.
func TestSavedFields(t *testing.T) {
	e := &Engine{}
	e.initVM()
	data := os.DirFS(filepath.Join("..", "data", "por", "sources", AssetsDir, "scripts"))
	names, err := fs.Glob(data, "layout/area*.txt")
	if err != nil {
		t.Fatal(err)
	}
	place := map[string]script.Field{}
	owner := map[int]string{}
	for _, n := range append(names, "layout/area0.txt") {
		var a int
		if _, err := fmt.Sscanf(n, "layout/area%d.txt", &a); err != nil {
			t.Fatal(err)
		}
		l, err := script.ReadLayout(data, a)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range l.Fields {
			if f.Owner == script.OwnerCall {
				continue
			}
			local := f.From >= 0x4A00 && f.From+f.Len <= 0x4A20 || f.From >= 0x6E79 && f.From+f.Len <= 0x6E83
			if local != (f.Owner == script.OwnerLocal) {
				t.Errorf("area %d: %s at %04X is %v; the words a fresh area clears are the locals", a, f.Name, f.From, f.Owner)
			}
			if _, ok := e.engineFields[f.Name]; !ok && f.Owner == script.OwnerFact {
				t.Errorf("%s: a fact the engine gives no meaning", f.Name)
			}
			if f.Owner == script.OwnerLocal || f.Replaces {
				continue
			}
			if p, ok := place[f.Name]; ok && (p.From != f.From || p.Len != f.Len) {
				t.Errorf("%s at %04X (area %d) and at %04X", f.Name, f.From, a, p.From)
			}
			place[f.Name] = f
			for i := range f.Len {
				if o, ok := owner[f.From+i]; ok && o != f.Name {
					t.Errorf("word %04X: %s and %s (area %d)", f.From+i, o, f.Name, a)
				}
				owner[f.From+i] = f.Name
			}
		}
	}
}

// TestSaveKeepsFields: a save and load keeps the game's fields and the
// area's locals, and the fields' words leave the engine's own words in
// the records alone. A text keeps to the words before the next field.
func TestSaveKeepsFields(t *testing.T) {
	e := quickStarted(t)
	withLayout(t, e)
	lay := e.vm.Script.Layout
	var saved []script.Field
	for i, fl := range lay.Fields {
		if !e.saved(fl) {
			continue
		}
		saved = append(saved, fl)
		r := script.Ref{Field: script.FieldID(i + 1)}
		if fl.Width == script.WidthText {
			if i+1 < len(lay.Fields) && lay.Fields[i+1].From-fl.From > 3 {
				e.host.SetText(r, "ABC")
			}
			continue
		}
		for k := range fl.Len {
			if !e.boundElement(&fl, k) {
				e.storeCells(&fl, k, []int64{0xBEEF - int64(k)})
			}
		}
	}
	values := func() []string {
		var out []string
		for i, fl := range lay.Fields {
			if !slices.ContainsFunc(saved, func(s script.Field) bool { return s.Name == fl.Name }) {
				continue
			}
			r := script.Ref{Field: script.FieldID(i + 1)}
			if fl.Width == script.WidthText {
				out = append(out, fl.Name+"="+e.host.Text(r))
				continue
			}
			for k := range fl.Len {
				if !e.boundElement(&fl, k) {
					out = append(out, fmt.Sprint(fl.Name, k, "=", e.getVar(r.Plus(k))))
				}
			}
		}
		return out
	}
	want := values()
	check := func(how string) {
		t.Helper()
		if got := values(); !slices.Equal(got, want) {
			for i := range want {
				if i < len(got) && got[i] != want[i] {
					t.Errorf("%s, back %s", want[i], got[i])
				}
			}
			t.Fatalf("the fields did not come back from %s", how)
		}
	}
	// the canonical save, by name
	path := filepath.Join(t.TempDir(), gameFile)
	if err := writeGameSave(path, gameSave{Fields: e.vars, Locals: e.locals}); err != nil {
		t.Fatal(err)
	}
	g, err := readGameSave(path)
	if err != nil {
		t.Fatal(err)
	}
	vars, locals := e.vars, e.locals
	e.vars, e.locals = g.Fields, g.Locals
	check("the canonical save")
	// an original save's records, by layout address (texts first, so a text
	// longer than its field gives way to the field after it)
	f := rules1e.OriginalSave{Area: e.state().Area, Game: e.state().Game}
	f.Area.AreaKind = uint16(e.vm.Script.Area)
	slices.SortStableFunc(saved, func(a, b script.Field) int {
		return rules1e.B2i(a.Width != script.WidthText) - rules1e.B2i(b.Width != script.WidthText)
	})
	for _, fl := range saved {
		cells := vars
		if fl.Owner == script.OwnerLocal {
			cells = locals
		}
		for i, v := range cells[fl.Name] {
			if p := testRecordCell(&f, fl.From+i); p != nil {
				*p = uint16(v)
			}
		}
	}
	b, err := binary.Append(nil, binary.LittleEndian, &f)
	if err != nil {
		t.Fatal(err)
	}
	if g, err = e.importSave(b); err != nil {
		t.Fatal(err)
	}
	r := blobOf(t, g)
	if r.Area.AreaKind = e.state().Area.AreaKind; r.Area != e.state().Area || r.Game != e.state().Game {
		t.Error("a field's words overwrote an engine word of the records")
	}
	e.vars, e.locals = g.Fields, g.Locals
	check("an original save's records")
}

// testRecordCell is the word of an original save's records at script address a.
func testRecordCell(f *rules1e.OriginalSave, a int) *uint16 {
	switch {
	case a >= areaVarBase && a < areaVarBase+rules1e.AreaWords:
		return f.Area.Word(a - areaVarBase)
	case a >= gameVarBase && a < gameVarBase+rules1e.GameWords:
		return f.Game.Word(a - gameVarBase)
	case a >= auxVarBase && a < auxVarBase+len(f.Aux):
		return &f.Aux[a-auxVarBase]
	}
	return nil
}

// POR's record bases in its script addresses (importpor's OriginalSave layout).
const (
	areaVarBase = 0x4900
	gameVarBase = 0x6B00
	auxVarBase  = 0x9700
)
