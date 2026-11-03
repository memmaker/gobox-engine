package tablespec

import (
	"encoding/json"
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
	if _, err := Generate("p", []Table{{Name: "x", At: 100, Len: 1}}, im); err == nil {
		t.Error("out of range accepted")
	}
	if String(7).Size() != 7 || Longs.Size() != 4 {
		t.Error("Size")
	}
}
