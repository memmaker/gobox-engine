package parity

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// plumbing are gobox's game↔UI types: a game's rules package that declares
// a type of one of these names keeps its own copy of gobox infrastructure
// (one canonical way: gobox/game/play).
var plumbing = []string{"Ask", "AskKind", "Question", "QKind", "Answer", "Decision", "Menu", "MenuItem",
	"ListLine", "Option", "Row", "Line", "Event", "EventKind", "EvKind", "Mode", "Refusal", "Note", "Result",
	"Choice", "Pick", "Entry", "Bar"}

// knownCopies are the copies still to remove (MIGRATION.md, phases 2-5).
// The list only shrinks: a new copy fails, and so does an entry whose copy
// is gone (delete it here in the same change).
var knownCopies = map[string][]string{}

// TestNoPrivatePlumbing: no game declares its own copy of gobox's game↔UI
// types beyond the known ones still being removed.
func TestNoPrivatePlumbing(t *testing.T) {
	games := []string{"por-go", "cotab-go", "sotsb-go", "pod-go", "buckrogers-go", "buckrogers2-go"}
	for _, g := range games {
		dir := filepath.Join("..", "..", g, "rules")
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Logf("%s: %v (not checked)", g, err)
			continue
		}
		var found []string
		fs := token.NewFileSet()
		for _, f := range files {
			if n := f.Name(); !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			af, err := parser.ParseFile(fs, filepath.Join(dir, f.Name()), nil, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range af.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, sp := range gd.Specs {
					if n := sp.(*ast.TypeSpec).Name.Name; slices.Contains(plumbing, n) && !slices.Contains(found, n) {
						found = append(found, n)
					}
				}
			}
		}
		slices.Sort(found)
		want := knownCopies[g]
		for _, n := range found {
			if !slices.Contains(want, n) {
				t.Errorf("%s/rules declares its own %s: use gobox/game/play's", g, n)
			}
		}
		for _, n := range want {
			if !slices.Contains(found, n) {
				t.Errorf("%s/rules no longer declares %s: remove it from knownCopies", g, n)
			}
		}
	}
}

// funcFields caps each game's function-typed struct fields in rules: a
// stored func is a continuation (the game paused mid-routine), and games
// are passive (MIGRATION.md, order 2). The caps only shrink.
var funcFields = map[string]int{"cotab-go": 2, "sotsb-go": 2, "pod-go": 1, "por-go": 8,
	"buckrogers-go": 1, "buckrogers2-go": 2}

// TestPassive: no game gains a stored callback; lower a cap when one goes.
func TestPassive(t *testing.T) {
	for g, max := range funcFields {
		dir := filepath.Join("..", "..", g, "rules")
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Logf("%s: %v (not checked)", g, err)
			continue
		}
		n := 0
		for _, f := range files {
			if nm := f.Name(); !strings.HasSuffix(nm, ".go") || strings.HasSuffix(nm, "_test.go") {
				continue
			}
			af, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, f.Name()), nil, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(af, func(x ast.Node) bool {
				if st, ok := x.(*ast.StructType); ok {
					for _, fl := range st.Fields.List {
						ty := fl.Type
						if s, ok := ty.(*ast.StarExpr); ok {
							ty = s.X
						}
						if a, ok := ty.(*ast.ArrayType); ok {
							ty = a.Elt
						}
						if _, ok := ty.(*ast.FuncType); ok {
							n += len(fl.Names) + min(1, 1-len(fl.Names)) // unnamed counts 1
						}
					}
				}
				return true
			})
		}
		switch {
		case n > max:
			t.Errorf("%s/rules stores %d funcs (cap %d): make the pending decision state, not a callback", g, n, max)
		case n < max:
			t.Errorf("%s/rules stores %d funcs: lower its funcFields cap from %d", g, n, max)
		}
	}
}

// oldPathImports and oldPathNames are the old path: no Go file of gobox or a
// game may import or name them.
var (
	oldPathImports = []string{"gobox/front/playrules"}
	oldPathNames   = map[string][]string{"game": {"Rules", "Mapped"}}
)

// TestNoOldPath fails on any reference to the removed rules path.
func TestNoOldPath(t *testing.T) {
	roots := []string{"..", "../../por-go", "../../cotab-go", "../../sotsb-go", "../../pod-go",
		"../../buckrogers-go", "../../buckrogers2-go"}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
				return err
			}
			af, err := parser.ParseFile(token.NewFileSet(), p, nil, parser.SkipObjectResolution)
			if err != nil {
				return nil // not this test's to judge: the build's
			}
			for _, im := range af.Imports {
				if slices.Contains(oldPathImports, strings.Trim(im.Path.Value, `"`)) {
					t.Errorf("%s imports %s", p, im.Path.Value)
				}
			}
			ast.Inspect(af, func(n ast.Node) bool {
				if s, ok := n.(*ast.SelectorExpr); ok {
					if x, ok := s.X.(*ast.Ident); ok && slices.Contains(oldPathNames[x.Name], s.Sel.Name) {
						t.Errorf("%s names %s.%s", p, x.Name, s.Sel.Name)
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
