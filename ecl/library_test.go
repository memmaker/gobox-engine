package ecl

import (
	"testing"
	"testing/fstest"
)

// TestLibraryKey: an area's text goes stale when the game's names, layout
// or special cases change, and only then (a text beside them is not data).
func TestLibraryKey(t *testing.T) {
	data := fstest.MapFS{
		"layout/global.txt": {Data: []byte("0001 1 byte game A\n")},
		"names/areas.txt":   {Data: []byte("")},
		"area0.txt":         {Data: []byte("version 1 2\n")},
	}
	var fixes []Fix
	key := func() string {
		k, err := Key(V2, data, FoldGates, fixes)
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	k0 := key()
	data["area0.txt"] = &fstest.MapFile{Data: []byte("version 1 2 x\n")}
	if key() != k0 {
		t.Error("a text beside the data changed the key")
	}
	data["layout/global.txt"] = &fstest.MapFile{Data: []byte("0001 1 byte game B\n")}
	k1 := key()
	if k1 == k0 {
		t.Error("a layout change kept the key")
	}
	data["names/area3.txt"] = &fstest.MapFile{Data: []byte("x")}
	if key() == k1 {
		t.Error("a new names file kept the key")
	}
	fixes = []Fix{{Area: 1, Old: "a\n", New: "b\n"}}
	if key() == k1 {
		t.Error("a new fix kept the key")
	}
	if textKey(withKey("version 2 5\nrest", k1)) != k1 || textKey(withKey("version 2 5 old\nrest", k1)) != k1 {
		t.Error("withKey/textKey")
	}
}
