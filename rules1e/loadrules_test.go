package rules1e

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRules: a rules.json file replaces the tables it carries; the rest stay POR's.
func TestLoadRules(t *testing.T) {
	p := filepath.Join(t.TempDir(), "rules.json")
	thac0 := `[[1],[2],[3],[4],[5],[6],[7]]` // 7 classes: no monk
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"thac0Table":` + thac0 + `,"classHitDiceSides":[9,9,9,9,9,9,9]}`); err != nil {
		t.Fatal(err)
	}
	f.Close()
	tb, err := POR.LoadRules(p, Layout{SlotStride: 7})
	if err != nil {
		t.Fatal(err)
	}
	if tb.Thac0At(ClassFighter, 5) != 3 || tb.Thac0At(ClassMonk, 1) != 0 {
		t.Error("thac0")
	}
	if _, s := tb.TrainingDie(ClassMage); s != 9 || tb.XPForLevel(ClassFighter, 2) != POR.XPForLevel(ClassFighter, 2) {
		t.Error("hit dice or xp")
	}
	if _, err := POR.LoadRules(filepath.Join(t.TempDir(), "none.json"), Layout{SlotStride: 5}); err == nil {
		t.Error("a missing file loads")
	}
}
