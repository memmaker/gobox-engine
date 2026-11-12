package pool

import (
	"testing"
)

// TestEnginePictures: New reads the adventure's data/pictures.json.
func TestEnginePictures(t *testing.T) {
	e, err := New(porRuleset, porTables, "../../por-go/data/assets", "../../por-go/data")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.pics) == 0 {
		t.Fatal("no picture table")
	}
}
