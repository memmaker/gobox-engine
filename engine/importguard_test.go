package engine

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// rulesetFiles are the non-test files that still import a ruleset
// (REFACTOR.md R2). The list only shrinks: a file not on it may not import
// one, and a file that no longer does must leave it.
var rulesetFiles = []string{
	"bestiary.go", "combatrules_engine.go", "dataimport.go", "eclops2_engine.go", "eclops_engine.go",
	"files_engine.go", "game_engine.go", "init.go", "newpool.go", "present_engine.go", "roster.go",
	"rulesetseam.go", "savefile.go", "scriptvars_engine.go", "slots.go", "vault_engine.go", "view_engine.go",
}

// TestPoolImportsNoRuleset: the engine reaches the rules only through
// package rules: no non-test file imports a ruleset (rules1e, rules4e,
// rulesxxvc) but those still to fix (rulesetFiles), and no file, test
// files included, a game's package. Bug class: the guard compared the
// import with "rules1e" when the path is "gobox/rules1e", and passed on
// seventeen files.
func TestPoolImportsNoRuleset(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	imports := map[string]bool{}
	for _, f := range files {
		ast, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range ast.Imports {
			p := strings.TrimPrefix(strings.Trim(imp.Path.Value, `"`), "gobox/")
			isGame := p == "por" || strings.HasPrefix(p, "por/") || strings.HasPrefix(p, "games/") || strings.HasPrefix(p, "buckrogers")
			isRuleset := p == "rules1e" || p == "rules4e" || p == "rulesxxvc" || strings.HasPrefix(p, "rules1e/") || strings.HasPrefix(p, "rules4e/")
			if isGame {
				t.Errorf("%s imports the game %s", f, p)
			}
			if isRuleset && !strings.HasSuffix(f, "_test.go") { // tests build their rules from rules1e
				imports[f] = true
				if !slices.Contains(rulesetFiles, f) {
					t.Errorf("%s imports the ruleset %s", f, p)
				}
			}
		}
	}
	for _, f := range rulesetFiles {
		if !imports[f] {
			t.Errorf("%s imports no ruleset any more: take it off rulesetFiles", f)
		}
	}
}
