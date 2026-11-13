package engine

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestPoolImportsNoRuleset: the engine reaches the rules only through core
// (fold E2): no non-test file imports a ruleset (rules1e, rulesxxvc), and no file, test files included, a
// game's package (por/...); the game hands its ruleset and data to New.
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
			p := strings.Trim(imp.Path.Value, `"`)
			isGame := p == "por" || strings.HasPrefix(p, "por/")
			isRuleset := p == "rules1e" || p == "rulesxxvc"
			// Tests build their rules from rules1e; no file may import a game.
			if isGame || isRuleset && !strings.HasSuffix(f, "_test.go") {
				t.Errorf("%s imports %s", f, p)
			}
		}
	}
}
