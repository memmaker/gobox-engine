package engine

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestActorSeam: the rules' actor (RuleState.who, DS:5D92) is set only from
// a call's argument. Every call outside RuleState's own methods into a
// RuleState method that reads the actor, directly or through what it calls
// (the affect handlers through callAffect's table), goes through Engine.as;
// and no function but as, follow and RuleState's methods names who. Detects
// a rule reading a selection the presenter changed without passing it.
func TestActorSeam(t *testing.T) {
	dir, err := exec.Command("go", "list", "-f", "{{.Dir}}", "rules1e").Output()
	if err != nil {
		t.Fatal(err)
	}
	tainted := actorReaders(t, strings.TrimSpace(string(dir)))
	if !tainted["Who"] || len(tainted) < 2 {
		t.Fatalf("no RuleState method reads the actor: %v", tainted)
	}
	t.Logf("%d RuleState names read the actor", len(tainted))
	fs, files, info := checkDir(t, ".")
	seam := map[string]bool{"as": true, "follow": true}
	for _, f := range files {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil || seam[fd.Name.Name] {
				continue
			}
			ast.Inspect(fd.Body, func(x ast.Node) bool {
				s, ok := x.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				sel := info.Selections[s]
				if sel == nil || !strings.HasSuffix(sel.Recv().String(), "rules1e.RuleState") && !strings.HasSuffix(sel.Recv().String(), "Engine") {
					return true
				}
				pos := fs.Position(s.Pos())
				switch o := sel.Obj().(type) {
				case *types.Var:
					if o.Name() == "Who" && o.Pkg().Path() == "rules1e" {
						t.Errorf("%s: %s names the rules' actor; select with sel, hand it with as", pos, fd.Name.Name)
					}
				case *types.Func:
					recv := o.Type().(*types.Signature).Recv()
					if recv == nil || !tainted[o.Name()] || !strings.HasSuffix(recv.Type().String(), "rules1e.RuleState") {
						return true
					}
					if c, ok := s.X.(*ast.CallExpr); ok {
						if f, ok := c.Fun.(*ast.SelectorExpr); ok && f.Sel.Name == "as" {
							return true
						}
					}
					t.Errorf("%s: %s calls %s, which reads the actor, without as", pos, fd.Name.Name, o.Name())
				}
				return true
			})
		}
	}
}

// checkDir type-checks the package in dir without its tests.
func checkDir(t *testing.T, dir string) (*token.FileSet, []*ast.File, *types.Info) {
	t.Helper()
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, func(fi os.FileInfo) bool {
		ok, _ := build.Default.MatchFile(dir, fi.Name())
		return ok && !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, p := range pkgs {
		for _, f := range p.Files {
			files = append(files, f)
		}
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	conf := types.Config{Importer: importer.ForCompiler(fs, "source", nil)}
	if _, err := conf.Check(files[0].Name.Name, fs, files, info); err != nil {
		t.Fatal(err)
	}
	return fs, files, info
}

// actorReaders is the names of the rules1e RuleState methods that read the
// actor (RuleState.Who), directly or through what they call (the affect
// handlers through affectHandler's users), and "Who" itself.
func actorReaders(t *testing.T, dir string) map[string]bool {
	_, files, info := checkDir(t, dir)
	var who types.Object
	for id, o := range info.Defs {
		if v, ok := o.(*types.Var); ok && v.IsField() && id.Name == "Who" {
			who = o
		}
	}
	var funcs []*ast.FuncDecl
	for _, f := range files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
				funcs = append(funcs, fd)
			}
		}
	}
	// refs: every function and method a function names; init's
	// handlers (affectHandler) count as named by the users of affectHandler.
	refs := map[types.Object][]types.Object{}
	reads := map[types.Object]bool{}
	var handlers []types.Object
	tableUsers := map[types.Object]bool{}
	for _, fd := range funcs {
		me := info.Defs[fd.Name]
		fills := false
		ast.Inspect(fd.Body, func(x ast.Node) bool {
			switch x := x.(type) {
			case *ast.AssignStmt:
				for _, l := range x.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name == "affectHandler" {
						fills = true
					}
				}
			case *ast.Ident:
				if x.Name == "affectHandler" {
					tableUsers[me] = true
				}
			case *ast.SelectorExpr:
				if sel := info.Selections[x]; sel != nil && sel.Obj() == who {
					reads[me] = true
				}
			}
			return true
		})
		var named []types.Object
		ast.Inspect(fd.Body, func(x ast.Node) bool {
			if id, ok := x.(*ast.Ident); ok {
				if fn, ok := info.Uses[id].(*types.Func); ok {
					named = append(named, fn)
				}
			}
			return true
		})
		if fills {
			handlers = append(handlers, named...)
			delete(tableUsers, me)
		} else {
			refs[me] = named
		}
	}
	for u := range tableUsers {
		refs[u] = append(refs[u], handlers...)
	}
	tainted := map[types.Object]bool{}
	for o := range reads {
		tainted[o] = true
	}
	for changed := true; changed; {
		changed = false
		for n, rs := range refs {
			for _, r := range rs {
				if !tainted[n] && tainted[r] {
					tainted[n], changed = true, true
				}
			}
		}
	}
	names := map[string]bool{"Who": true}
	for o := range tainted {
		if f, ok := o.(*types.Func); ok {
			if r := f.Type().(*types.Signature).Recv(); r != nil && strings.HasSuffix(r.Type().String(), ".RuleState") {
				names[f.Name()] = true
			}
		}
	}
	return names
}
