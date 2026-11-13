package engine

import (
	"gobox/rules1e"
	"testing"
)

// TestEnginePictures: New reads the adventure's data/pictures.json.
func TestEnginePictures(t *testing.T) {
	e, err := New(rules1e.Rules1e, porTables, "../data/por/sources/assets", "../data/por/sources")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.pics) == 0 {
		t.Fatal("no picture table")
	}
}
