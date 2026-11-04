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
