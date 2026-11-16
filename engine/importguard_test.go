package engine

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
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

// TestNoBlockLiteralsInEngine: a script block's number is a game's data
// (adventure.GameTables), never an engine literal. Bug class: POR's
// wilderness blocks 0x19..0x1B and block 10 compared in the engine.
func TestNoBlockLiteralsInEngine(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(b), "\n") {
			if c := strings.Index(line, "//"); c >= 0 {
				line = line[:c]
			}
			if regexp.MustCompile(`areaKind\s*(==|!=)\s*(0x|\d)`).MatchString(line) && !strings.Contains(line, "== 0 ") {
				t.Errorf("%s:%d compares areaKind with a block literal: %s", f, i+1, strings.TrimSpace(line))
			}
			if regexp.MustCompile(`case 0x1[9AB]:`).MatchString(line) {
				t.Errorf("%s:%d names a POR overland block: %s", f, i+1, strings.TrimSpace(line))
			}
		}
	}
}
