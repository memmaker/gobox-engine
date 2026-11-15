package game

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// engineSide are the packages behind the game contract: they decide what
// happens, never where or how it is shown.
var engineSide = []string{"game", "engine", "ecl", "adventure", "rules1e", "rules4e"}

// TestEngineDrawsNothing: no engine-side package reaches a drawing or
// presenting package, directly or not.
func TestEngineDrawsNothing(t *testing.T) {
	drawing := []string{"gobox/gpu", "gobox/ui", "image/draw", "golang.org/x/image/draw", "github.com/ebitengine"}
	for _, pkg := range engineSide {
		out, err := exec.Command("go", "list", "-deps", "gobox/"+pkg).Output()
		if err != nil {
			t.Fatal(pkg, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			for _, d := range drawing {
				if dep == d || strings.HasPrefix(dep, d+"/") {
					t.Errorf("%s depends on %s", pkg, dep)
				}
			}
		}
	}
}

// TestEnginePlacesNothing: the game contract cannot express a place on
// the screen. No field of package game holds an image.Point or an
// image.Rectangle (a square of the world is a Square); no field holds
// pre-wrapped lines ([][]TextRun) are allowed: they keep the original's design.
func TestEnginePlacesNothing(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			st, ok := n.(*ast.StructType)
			if !ok {
				return true
			}
			for _, fl := range st.Fields.List {
				if placing(fl.Type) {
					t.Errorf("%s: a field placing on the screen", fset.Position(fl.Pos()))
				}
			}
			return true
		})
	}
}

// placing: t is, or holds, a screen place.
func placing(t ast.Expr) bool {
	switch t := t.(type) {
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)
		return ok && x.Name == "image" && (t.Sel.Name == "Point" || t.Sel.Name == "Rectangle")
	case *ast.StarExpr:
		return placing(t.X)
	case *ast.ArrayType:
		return placing(t.Elt)
	case *ast.MapType:
		return placing(t.Key) || placing(t.Value)
	}
	return false
}

// TestEngineWritesNoCells: the helpers that wrote text into page cells
// and drew the classic view stay deleted from every engine-side package;
// text is lines by target, cleared on events (engine/text.go).
func TestEngineWritesNoCells(t *testing.T) {
	gone := map[string]bool{"char": true, "blank": true, "unwrite": true, "Draw3DView": true, "Flow": true, "textBlock": true,
		"classicView": true, "classicBlit": true, "classicClear": true, "TestClassicViews": true}
	fset := token.NewFileSet()
	for _, pkg := range engineSide {
		files, _ := filepath.Glob(filepath.Join("..", pkg, "*.go"))
		for _, f := range files {
			af, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range af.Decls {
				if fd, ok := d.(*ast.FuncDecl); ok && gone[fd.Name.Name] {
					t.Errorf("%s: %s is back", fset.Position(fd.Pos()), fd.Name.Name)
				}
			}
		}
	}
}
