package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"gobox/rules1e"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/ecl"
)

// field is the loaded script's field called name, for tests on POR's scripts: a test that
// names a field the game lacks fails here. Shared code reads fields only through layoutField.
func (e *Engine) field(name string) ecl.Ref {
	r, ok := e.layoutField(name)
	if !ok {
		panic("no field " + name)
	}
	return r
}

// TestFieldReadsBound (bug class: shared pool code read POR-only script fields by name and
// panicked on other games): every field name pool's code reads (layoutField's constant
// argument) is in BoundFields, so each game's UnboundFields test lists it when its layout
// lacks it, and no code path can require a field a game does not have.
func TestFieldReadsBound(t *testing.T) {
	fset := token.NewFileSet()
	files, _ := filepath.Glob("*.go")
	consts := map[string]string{}
	var reads []ast.Expr
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.ValueSpec:
				for i, id := range n.Names {
					if i < len(n.Values) {
						if l, ok := n.Values[i].(*ast.BasicLit); ok && l.Kind == token.STRING {
							consts[id.Name], _ = strconv.Unquote(l.Value)
						}
					}
				}
			case *ast.CallExpr:
				if s, ok := n.Fun.(*ast.SelectorExpr); ok && (s.Sel.Name == "layoutField" || s.Sel.Name == "field") && len(n.Args) == 1 {
					reads = append(reads, n.Args[0])
				}
			}
			return true
		})
	}
	if len(reads) == 0 {
		t.Fatal("no layoutField reads found")
	}
	bound := BoundFields((&rules1e.Rules{}).Catalog())
	for _, a := range reads {
		id, ok := a.(*ast.Ident)
		if !ok {
			t.Errorf("%v: field read by a non-constant name", fset.Position(a.Pos()))
			continue
		}
		if name, ok := consts[id.Name]; !ok || !slices.Contains(bound, name) {
			t.Errorf("%v: field %s (%q) is read but not in BoundFields", fset.Position(a.Pos()), id.Name, name)
		}
	}
}
