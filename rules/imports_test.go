package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The parity guard, sketched: core imports nothing of ours, and each leaf
// (a ruleset or adventure package) imports only core and the standard
// library, directly or not. No leaf exists yet, so leaves is empty: a
// migrated package adds its module directory here.
var leaves = []string{
	"../rules1e",
	"../rulesxxvc",
}

// foreign lists what the packages under dir depend on besides the standard
// library, gobox/rules and the packages under dir's own import path.
func foreign(dir string) ([]string, error) {
	rootCmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", ".")
	rootCmd.Dir = dir
	root, err := rootCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	own := strings.TrimSpace(string(root))
	cmd := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}|{{.Module.Path}}|{{.Module.Main}}{{end}}", "./...")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	var bad []string
	for _, l := range strings.Fields(string(out)) {
		p := strings.Split(l, "|")
		if p[0] != "gobox/rules" && !strings.HasPrefix(p[0], "gobox/rules/") && !(p[2] == "true" && strings.HasPrefix(p[0], own)) {
			bad = append(bad, p[0])
		}
	}
	return bad, nil
}

func TestLeavesImportOnlyCore(t *testing.T) {
	for _, dir := range append([]string{"."}, leaves...) {
		bad, err := foreign(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(bad) > 0 {
			t.Errorf("%s imports %v", dir, bad)
		}
	}
}

// diskCodecs are rules1e's files that serialize to disk: the characters'
// records and the original save image stay in the games' own byte formats
// (with sibling files beside them), so these, their tests and the id
// generator hold bytes. What is in memory stays typed.
var diskCodecs = []string{"char_record.go", "charfile.go", "later.go", "original.go", "idgen"}

// rawBytes lists what the Go files under dir use to reach raw memory: unsafe,
// encoding/binary, and any byte slice or array type. Scripts are the one byte
// passage (Adventure.Script's result); diskCodecs and tests are left out.
func rawBytes(dir string) ([]string, error) {
	var bad []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && slices.Contains(diskCodecs, d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			if v := strings.Trim(imp.Path.Value, `"`); v == "unsafe" || v == "encoding/binary" {
				bad = append(bad, fset.Position(imp.Pos()).String()+" imports "+v)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			// Script, and Rules.Save and Load (the ruleset's canonical save), are the byte passages.
			if fn, ok := n.(*ast.FuncDecl); ok && (fn.Name.Name == "Script" || fn.Name.Name == "Save" || fn.Name.Name == "Load") {
				return false
			}
			if a, ok := n.(*ast.ArrayType); ok {
				if e, ok := a.Elt.(*ast.Ident); ok && (e.Name == "byte" || e.Name == "uint8") {
					bad = append(bad, fset.Position(a.Pos()).String()+" raw bytes")
				}
			}
			return true
		})
		return nil
	})
	return bad, err
}

func TestLeavesNoRawBytes(t *testing.T) {
	for _, dir := range leaves {
		bad, err := rawBytes(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range bad {
			t.Error(b)
		}
	}
}

// TestRawBytesCatches plants each forbidden form.
func TestRawBytesCatches(t *testing.T) {
	dir := t.TempDir()
	src := "package l\n\nimport _ \"unsafe\"\n\nvar a [4]byte\nvar b []uint8\n\nfunc Script() []byte { return nil }\n"
	if err := os.WriteFile(filepath.Join(dir, "l.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bad, err := rawBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) != 3 {
		t.Fatalf("want the import and two types, got %v", bad)
	}
}

// TestGuardCatches plants a leaf importing another gobox module.
func TestGuardCatches(t *testing.T) {
	dir := t.TempDir()
	dep := filepath.Join(dir, "dep")
	leaf := filepath.Join(dir, "leaf")
	for f, s := range map[string]string{
		filepath.Join(dep, "go.mod"):  "module gobox/dep\n\ngo 1.26.2\n",
		filepath.Join(dep, "d.go"):    "package dep\n",
		filepath.Join(leaf, "go.mod"): "module leaf\n\ngo 1.26.2\n\nrequire gobox/dep v0.0.0\n\nreplace gobox/dep => ../dep\n",
		filepath.Join(leaf, "l.go"):   "package leaf\n\nimport _ \"gobox/dep\"\n",
	} {
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bad, err := foreign(leaf)
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) != 1 || bad[0] != "gobox/dep" {
		t.Fatalf("planted import not caught: %v", bad)
	}
}
