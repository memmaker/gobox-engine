package rules1e

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRules1eLoadsGameTables: a game that carries rules.json plays on those tables, laid out by
// the layout.json beside it; without that layout the build fails loudly, and a game with no
// rules.json (POR) builds on rules1e.POR.
func TestRules1eLoadsGameTables(t *testing.T) {
	dir := t.TempDir()
	rules := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(rules, []byte(`{"thac0Table":[[1],[1],[1],[1],[1],[1],[1],[1]]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	build := func(path string) (r *Rules, panicked any) {
		defer func() { panicked = recover() }()
		c, _ := Rules1e(path, nil, nil, nil, nil)
		return c.(*Rules), nil
	}
	if _, p := build(rules); p == nil {
		t.Fatal("rules.json without layout.json built")
	}
	if err := os.WriteFile(filepath.Join(dir, LayoutFile), []byte(`{"SlotStride":5}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, p := build(rules)
	if p != nil || r.S.Tables == nil || len(r.S.Tables.Thac0) != 8 || len(r.S.Tables.Thac0[0]) != 1 {
		t.Fatalf("with its layout: %v, tables %+v", p, r)
	}
	if r, p = build(""); p != nil || r.S.Tables != nil {
		t.Fatalf("no rules.json: panic %v, tables %v: want POR's (nil)", p, r.S.Tables)
	}
}
