package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The parity guard, sketched: core imports nothing of ours, and each leaf
// (a ruleset or adventure package) imports only core and the standard
// library, directly or not. No leaf exists yet, so leaves is empty: a
// migrated package adds its module directory here.
var leaves = []string{
	// "../../rules1e", "../../por-go/adventure", ...
}

// foreign lists what the packages under dir depend on besides the standard
// library, gobox/core and the module's own packages.
func foreign(dir string) ([]string, error) {
	cmd := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}|{{.Module.Path}}|{{.Module.Main}}{{end}}", "./...")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	var bad []string
	for _, l := range strings.Fields(string(out)) {
		p := strings.Split(l, "|")
		if p[1] != "gobox/core" && p[2] != "true" {
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
