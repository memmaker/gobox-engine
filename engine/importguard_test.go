package engine

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestPoolImportsNoRuleset: the engine reaches the rules only through
// package rules (REFACTOR.md R2): no non-test file imports a ruleset
// (rules1e, rules4e, rulesxxvc), and no file, test files included, a
// game's package. No list of exceptions. Bug class: the guard compared the
// import with "rules1e" when the path is "gobox/rules1e", and passed on
// seventeen files.
func TestPoolImportsNoRuleset(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
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
				t.Errorf("%s imports the ruleset %s", f, p)
			}
		}
	}
}
