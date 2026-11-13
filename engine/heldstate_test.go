package engine

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestStateReadsNoDisk (bug class: per-frame disk reads): the party
// menu's State, read every frame, is the one built at the last change: with
// the save folder gone after the first State, State is the same.
func TestStateReadsNoDisk(t *testing.T) {
	e := quickStarted(t)
	before := e.State()
	if !strings.Contains(fmt.Sprintf("%+v", before), "Roster:[* ") {
		t.Fatalf("roster not in the state: %+v", before)
	}
	dir := e.Cfg.SaveDir
	if err := os.Rename(dir, dir+".gone"); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(dir+".gone", dir)
	if after := e.State(); !reflect.DeepEqual(before, after) {
		t.Errorf("State read the disk:\n%+v\n%+v", before, after)
	}
}
