package engine

import (
	"reflect"
	"strings"
	"testing"

	"gobox/ecl"
)

// TestFoldElements: a save written while QuestSokalKeep, Word7B91 were
// fields of their own loads into the arrays they are now elements of; a
// name that is still a field stays.
func TestFoldElements(t *testing.T) {
	l, err := ecl.ParseLayout(strings.NewReader(`4AA6 26 byte game QuestTable
4AA7 = QuestSokalKeep
4AB3 = Boss
7B90 16 text game Word7B90
7B91 = Word7B91
7B9A = Word7B9A
7C00 1 byte game Other
`))
	if err != nil {
		t.Fatal(err)
	}
	old := fieldCells{
		"QuestSokalKeep": {255},
		"Word7B90":       {'G', 'O', 0},
		"Word7B91":       {1, 2, 3, 4, 5, 6, 7, 8, 9},
		"Word7B9A":       {'S', 0},
		"Other":          {7},
	}
	foldElements(old, l.Fields)
	want := fieldCells{
		"QuestTable": {0, 255},
		"Word7B90":   {'G', 1, 2, 3, 4, 5, 6, 7, 8, 9, 'S', 0},
		"Other":      {7},
	}
	if !reflect.DeepEqual(old, want) {
		t.Errorf("got %v, want %v", old, want)
	}
}
