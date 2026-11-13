package engine

import (
	"fmt"
	"gobox/ecl"
	"io/fs"
	"os"
	"slices"
	"testing"
)

// TestBoundFieldsPOR: POR's layouts name every field the engine binds but the VM words POR
// keeps in the ruleset, and the shops' price factor (POR's shops have none).
func TestBoundFieldsPOR(t *testing.T) {
	got, err := UnboundFields(os.DirFS("../data/por/sources/assets/scripts"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{fieldRemainder, fieldLastBlock, fieldPriceScale}; !slices.Equal(got, want) {
		t.Errorf("unbound %v, want %v", got, want)
	}
}

// BoundFields are the canonical script field names the engine gives a meaning: a game's
// layout names a field so when, and only when, it has that field (each game's layout test).
func BoundFields() []string {
	names := []string{fieldLastBlock, fieldRemainder, fieldPriceScale, fieldScratch6, fieldHours, tourDone}
	for n := range (&Engine{}).newEngineFields() {
		names = append(names, n)
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// UnboundFields are the BoundFields that no layout of data (a game's script data: layout/global.txt
// and every layout/area<N>.txt) names: the bindings the game does not have.
func UnboundFields(data fs.FS) ([]string, error) {
	names, err := fs.Glob(data, "layout/area*.txt")
	if err != nil {
		return nil, err
	}
	have := map[string]bool{}
	for _, n := range append(names, "layout/global.txt") {
		a := 0
		if n != "layout/global.txt" {
			if _, err := fmt.Sscanf(n, "layout/area%d.txt", &a); err != nil {
				return nil, err
			}
		}
		l, err := ecl.ReadLayout(data, a)
		if err != nil {
			return nil, err
		}
		for _, f := range l.Fields {
			have[f.Name] = true
			for _, e := range f.Elements {
				have[e] = true
			}
		}
	}
	var out []string
	for _, n := range BoundFields() {
		if !have[n] {
			out = append(out, n)
		}
	}
	return out, nil
}
