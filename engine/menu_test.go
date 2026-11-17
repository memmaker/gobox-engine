package engine

import (
	"gobox/rules1e"
	"testing"
)

func TestItemStats(t *testing.T) {
	e := &Engine{}
	e.state().Items = make([]rules1e.ItemBase, 5)
	e.initVM()
	e.state().Items[1].Dice, e.state().Items[1].Sides, e.state().Items[1].Bonus = 2, 4, 1 // type 1: 2d4+1
	e.state().Items[2].Protection = rules1e.AcGiven | 1                                   // type 2: shield 1
	e.state().Items[3].Protection = rules1e.AcGiven | (rules1e.AcTen + 7)                 // type 3: plate mail 7
	e.state().Items[4].Protection = rules1e.AcGiven | rules1e.AcTen                       // type 4: bracers 0
	for typ, want := range []string{"", "(2d4+1) ", "<1> ", "<7> ", "<0> "} {
		if s := e.state().ItemStats(&rules1e.Item{Type: rules1e.ItemType(typ)}); s != want {
			t.Errorf("type %d: %q, want %q", typ, s, want)
		}
	}
	for _, c := range []struct { // the known plus counts in; a hidden one does not
		typ    rules1e.ItemType
		plus   int8
		hidden byte
		want   string
	}{{4, 4, 0, "<4> "}, {4, 4, 1, "<0> "}, {3, 2, 0, "<9> "}, {2, 1, 0, "<2> "}, {1, 3, 0, "(2d4+1) "}} {
		if s := e.state().ItemStats(&rules1e.Item{Type: c.typ, Plus: byte(c.plus), Hidden: c.hidden}); s != c.want {
			t.Errorf("type %d %+d hidden %d: %q, want %q", c.typ, c.plus, c.hidden, s, c.want)
		}
	}
}

// An ECL picker list (opVMenu) has an empty bar: menuItems gives one item
// with no word, which ListMenu must not index.
func TestOptionText(t *testing.T) {
	for in, want := range map[string]string{"sWeep": "Sweep", "aiM": "Aim", "HEAL": "HEAL", "Exit": "Exit"} {
		if got := optionText(in); got != want {
			t.Errorf("optionText(%q) = %q, want %q", in, got, want)
		}
	}
}
