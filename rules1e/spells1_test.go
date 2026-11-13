package rules1e

import (
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"
)

// TestSpellEffects: spellEffect has one case for every spell number
// 1..len(spellTable)-1, as the original's DS:6B84 table had a procedure.
func TestSpellEffects(t *testing.T) {
	fset := token.NewFileSet()
	names, _ := filepath.Glob("*.go")
	var files []*ast.File
	for _, n := range names {
		if len(n) > 8 && n[len(n)-8:] == "_test.go" {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil), Error: func(error) {}}
	conf.Check("rules1e", fset, files, info)
	seen := map[int64]bool{}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name.Name != "spellEffect" {
				return true
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if cc, ok := n.(*ast.CaseClause); ok {
					for _, x := range cc.List {
						v, _ := constant.Int64Val(info.Types[x].Value)
						if seen[v] {
							t.Errorf("spell %d twice", v)
						}
						seen[v] = true
					}
				}
				return true
			})
			return false
		})
	}
	for i := int64(1); i < int64(len(POR.Spells)); i++ {
		if !seen[i] && SpellIDs[i] != "" { // a removed spell (56 Restoration) has no id and no effect
			t.Errorf("spell %d has no effect", i)
		}
	}
}
