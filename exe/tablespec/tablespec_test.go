package tablespec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/exe"
)

func TestGenerate(t *testing.T) {
	im := &exe.Image{Mem: []byte{1, 0xFF, 2, 0, 0xFE, 0xFF, 2, 'h', 'i', 9, 0, 0, 0, 0, 0, 0}}
	var tables []Table
	err := json.Unmarshal([]byte(`[
	 {"name":"a","at":0,"len":2,"kind":"signed","doc":"two"},
	 {"name":"w","at":2,"len":2,"kind":"swords","cols":1},
	 {"name":"s","at":6,"len":2,"kind":"string:4"}]`), &tables)
	if err != nil {
		t.Fatal(err)
	}
	src, err := Generate("p", tables, im)
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	for _, want := range []string{"package p", "// two", "var a = [2]int8{1, -1}", "[2][1]int16{", "{2}", "{-2}", `[2]string{"hi", ""`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in\n%s", want, s)
		}
	}
	js, err := JSON(tables, im)
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n \"a\": [1,-1],\n \"w\": [[2],[-2]],\n \"s\": [\"hi\",\"\"]\n}\n"; string(js) != want {
		t.Errorf("JSON %q, want %q", js, want)
	}
	if _, err := Generate("p", []Table{{Name: "x", At: 100, Len: 1}}, im); err == nil {
		t.Error("out of range accepted")
	}
	if String(7).Size() != 7 || Longs.Size() != 4 {
		t.Error("Size")
	}
}

// TestWriteRules: rules.json holds the tables of the images given; a table of another executable is left out.
func TestWriteRules(t *testing.T) {
	im := &exe.Image{Mem: []byte{1, 2}}
	dir := t.TempDir()
	if err := WriteRules([]byte(`[{"name":"a","at":0,"len":2,"kind":"bytes"},{"name":"o","at":0,"len":1,"kind":"bytes","img":1}]`), dir, im); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, RulesFile))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n \"a\": [1,2]\n}\n"; string(b) != want {
		t.Errorf("rules.json %q, want %q", b, want)
	}
}
