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
// engine (REFACTOR.md). Production code only: ui's tests may drive a real engine.
func TestUINotEngine(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "gobox/ui").Output()
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

// TestEngineNotAudio: the engine plays through game.Sound, game.SoundPack
// and game.Voice; the host drives gobox/audio (REFACTOR.md E2).
func TestEngineNotAudio(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "gobox/engine").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/audio" || strings.HasPrefix(p, "gobox/audio/") {
			t.Errorf("gobox/engine depends on %s", p)
		}
	}
}

// TestGoboxNotRules1e: cmd/gobox reaches rules1e and rules4e only through
// engine (REFACTOR.md R1, R3); direct imports, tests included.
func TestGoboxNotRules1e(t *testing.T) {
	out, err := exec.Command("go", "list", "-test", "-f", `{{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}`, "gobox/cmd/gobox").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/rules1e" || p == "gobox/rules4e" {
			t.Errorf("gobox/cmd/gobox imports %s", p)
		}
	}
}

// TestRules4eNotEngine: a ruleset depends on rules (and its base rules1e),
// never on engine (REFACTOR.md R3).
func TestRules4eNotEngine(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "gobox/rules4e").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p == "gobox/engine" {
			t.Errorf("gobox/rules4e depends on %s", p)
		}
	}
}
