package game

import (
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// TestGameImports runs scripts/check-game-imports.sh and compares with
// testdata/violations.txt: the list may only shrink.
func TestGameImports(t *testing.T) {
	out, err := exec.Command("bash", "../scripts/check-game-imports.sh").Output()
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/violations.txt")
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, l := range strings.Split(string(want), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			known[l] = true
		}
	}
	var bad []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l == "" {
			continue
		}
		if !known[l] {
			bad = append(bad, "new violation: "+l)
		}
		delete(known, l)
	}
	for l := range known {
		bad = append(bad, "fixed, delete the line from testdata/violations.txt: "+l)
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Fatal("\n" + strings.Join(bad, "\n"))
	}
}

// TestUINotEngine: ui reads games only through this contract, never the pool
// engine (REFACTOR.md).
func TestUINotEngine(t *testing.T) {
	out, err := exec.Command("go", "list", "-test", "-deps", "gobox/ui").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/engine" || strings.HasPrefix(p, "gobox/engine/") {
			t.Errorf("gobox/ui depends on %s", p)
		}
	}
}

// TestEngineNotImporter: the engine reads an adventure folder (package
// adventure), never the importer that writes it (REFACTOR.md).
func TestEngineNotImporter(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "gobox/engine").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/importer" || strings.HasPrefix(p, "gobox/importer/") {
			t.Errorf("gobox/engine depends on %s", p)
		}
	}
}

// TestGoboxNotRules1e: cmd/gobox reaches rules1e only through engine
// (REFACTOR.md R1); direct imports, tests included.
func TestGoboxNotRules1e(t *testing.T) {
	out, err := exec.Command("go", "list", "-test", "-f", `{{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}`, "gobox/cmd/gobox").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/rules1e" {
			t.Errorf("gobox/cmd/gobox imports %s", p)
		}
	}
}
