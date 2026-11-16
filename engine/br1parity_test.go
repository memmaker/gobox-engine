package engine

import (
	"encoding/csv"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBR1ParityTable keeps tracks/br1-parity/tests.csv (the audit of the old BR1 engine's tests against the shared
// engine) honest: every row has a known status, a covered or port-done row names the test functions that assert it
// and each exists in this repository, a port-open, n/a or differs row says why, and no row is listed twice.
func TestBR1ParityTable(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "tracks", "br1-parity", "tests.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 2 {
		t.Fatal("no rows")
	}
	have := testFunctions(t)
	seen := map[string]bool{}
	for _, r := range rows[1:] {
		file, test, status, named, note := r[0], r[1], r[3], r[4], r[5]
		key := file + "/" + test
		if seen[key] {
			t.Errorf("%s is listed twice", key)
		}
		seen[key] = true
		switch status {
		case "covered", "port-done":
			if named == "" {
				t.Errorf("%s: %s names no test", key, status)
			}
			for _, n := range strings.Split(named, ";") {
				if !have[n] {
					t.Errorf("%s: %s names %s, which is no test function in this repository", key, status, n)
				}
			}
		case "port-open", "n/a", "differs":
			if note == "" {
				t.Errorf("%s: %s with no reason", key, status)
			}
			if named != "" {
				t.Errorf("%s: %s names tests %q", key, status, named)
			}
		default:
			t.Errorf("%s: unknown status %q", key, status)
		}
	}
}

// testFunctions is the name of every Test function of every _test.go file under the repository.
func testFunctions(t *testing.T) map[string]bool {
	t.Helper()
	names := map[string]bool{}
	fset := token.NewFileSet()
	err := filepath.WalkDir("..", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "data" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(p, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Test") {
				names[fn.Name.Name] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return names
}
