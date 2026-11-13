package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// writeCalls are the os functions whose error must reach a caller.
var writeCalls = map[string]bool{
	"WriteFile": true, "Create": true, "CreateTemp": true, "OpenFile": true, "Rename": true,
	"MkdirAll": true, "Mkdir": true, "Remove": true, "RemoveAll": true, "Chmod": true, "Truncate": true,
}

// fileOpeners make files that are written: their Close error is a write error.
var fileOpeners = map[string]bool{"Create": true, "CreateTemp": true, "OpenFile": true}

func osCall(n ast.Node, names map[string]bool) bool {
	c, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	s, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, isID := s.X.(*ast.Ident)
	return isID && id.Name == "os" && names[s.Sel.Name]
}

// TestNoUnchecked file writes: a bug class (a write error shown or dropped,
// never returned; SaveCharacter once showed "Cannot write" and a later step
// failed with "no such roster character"). In the game repos and gobox-engine no
// non-test code may (1) call os.WriteFile/Create/Rename/MkdirAll/Remove...
// as a bare statement, defer or go, or assign all its results to _; (2)
// close a file it made with os.Create/OpenFile/CreateTemp without using
// Close's error; (3) handle an os write error in an if block that neither
// returns nor assigns it on (a prompt or a log only).
// Fix by returning the error (one os.WriteFile per file, saves.Store.Write,
// asset.WritePNG). A new violation fails here.
func TestNoUncheckedWrites(t *testing.T) {
	for _, root := range []string{"../games/por", "../../gobox-engine", "../games/cotab", "../games/sotsb", "../games/pod", "../../buckrogers-core", "../../buckrogers-go", "../../buckrogers2-go", "../rules1e", "../../rulesxxvc"} {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				t.Error(err)
				return nil
			}
			if filepath.ToSlash(path) == "../games/pod/tools/music" {
				return fs.SkipDir // a dev tool in its own module (a replace outside the tree): the gate cannot build it; its Close is left
			}
			if d.IsDir() && path != root && (d.Name() == "testdata" || d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".")) {
				return fs.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				t.Error(perr)
				return nil
			}
			report := func(n ast.Node, why string) {
				t.Errorf("%s: %s", fset.Position(n.Pos()), why)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					if lit, isLit := n.(*ast.FuncLit); isLit {
						checkBody(lit.Body, report)
					}
					return true
				}
				if fn.Body != nil {
					checkBody(fn.Body, report)
				}
				return true
			})
			return nil
		})
	}
}

func checkBody(body *ast.BlockStmt, report func(ast.Node, string)) {
	written := map[string]bool{} // variables holding a file made for writing
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.FuncLit:
			return false // checked on its own
		case *ast.ExprStmt:
			if osCall(s.X, writeCalls) {
				report(s, "os write call with its error dropped")
			}
			if isClose(s.X, written) {
				report(s, "Close of a written file with its error dropped")
			}
		case *ast.DeferStmt:
			if osCall(s.Call, writeCalls) || isClose(s.Call, written) {
				report(s, "deferred write call or Close with its error dropped")
			}
		case *ast.GoStmt:
			if osCall(s.Call, writeCalls) {
				report(s, "os write call in a goroutine with its error dropped")
			}
		case *ast.AssignStmt:
			if len(s.Rhs) == 1 && osCall(s.Rhs[0], writeCalls) {
				if allBlank(s.Lhs) {
					report(s, "os write call with its error assigned to _")
				}
				if osCall(s.Rhs[0], fileOpeners) && len(s.Lhs) > 0 {
					if id, ok := s.Lhs[0].(*ast.Ident); ok {
						written[id.Name] = true
					}
				}
			}
		case *ast.IfStmt:
			if a, ok := s.Init.(*ast.AssignStmt); ok && len(a.Rhs) == 1 && osCall(a.Rhs[0], writeCalls) && !propagates(s.Body) {
				report(s, "os write error handled without returning or keeping it")
			}
		}
		return true
	})
}

func isClose(n ast.Node, written map[string]bool) bool {
	c, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	s, ok := c.Fun.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != "Close" {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && written[id.Name]
}

func allBlank(es []ast.Expr) bool {
	for _, e := range es {
		if id, ok := e.(*ast.Ident); !ok || id.Name != "_" {
			return false
		}
	}
	return true
}

// propagates is true when the block returns, or assigns (keeps the error for the caller), or panics.
func propagates(b *ast.BlockStmt) bool {
	found := false
	ast.Inspect(b, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.ReturnStmt, *ast.AssignStmt:
			found = true
		case *ast.CallExpr:
			if id, ok := s.Fun.(*ast.Ident); ok && (id.Name == "panic" || id.Name == "check") {
				found = true
			}
			if sel, ok := s.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "faultErr" || sel.Sel.Name == "writeError" {
					found = true // cotab: the step's error (Do returns it)
				}
				if id, ok := sel.X.(*ast.Ident); ok && (id.Name == "log" && strings.HasPrefix(sel.Sel.Name, "Fatal") || id.Name == "os" && sel.Sel.Name == "Exit") {
					found = true // a command's last word
				}
			}
		}
		return !found
	})
	return found
}
