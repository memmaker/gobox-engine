package main

import (
	"path/filepath"
	"testing"
)

// TestKebab: a new spell's string follows the rule of every one already in core/dnd.
func TestKebab(t *testing.T) {
	_, ids, err := existing(filepath.Join("..", "spells.go"), spellLine)
	if err != nil || len(ids) == 0 {
		t.Fatalf("spells.go: %d ids, %v", len(ids), err)
	}
	for name, s := range ids {
		if k := kebab(name); k != s {
			t.Errorf("kebab(%s) = %q, the file has %q", name, k, s)
		}
	}
}
