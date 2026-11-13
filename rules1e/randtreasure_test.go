package rules1e

import "testing"

// TestSotsbTreasure: every type SotSB's type roll gives is made by a kind,
// a scroll or a template, and every template lies within Templates.
func TestSotsbTreasure(t *testing.T) { treasureCovered(t, SotsbTreasure()) }

// TestCotabTreasure and TestPorTreasure: the same for cotab's and POR's.
func TestCotabTreasure(t *testing.T) {
	treasureCovered(t, CotabTreasure())
}

func treasureCovered(t *testing.T, tr *TreasureRolls) {
	var walk func(r Roll[ItemType])
	walk = func(r Roll[ItemType]) {
		for _, c := range r.Cases {
			if c.Reroll {
				continue
			}
			if c.Then != nil {
				walk(*c.Then)
				continue
			}
			_, k := tr.Kinds[c.Value]
			_, s := tr.Scrolls[c.Value]
			_, m := tr.Magic[c.Value]
			if !k && !s && !m {
				t.Errorf("type %d: nothing makes it", c.Value)
			}
		}
	}
	walk(tr.Type)
	in := func(r Roll[TemplateAt]) {
		for _, c := range r.Cases {
			if c.Value != NoTemplate && (c.Value < 1 || int(c.Value)+7 > len(tr.Templates)) {
				t.Errorf("template %d outside %d words", c.Value, len(tr.Templates))
			}
		}
	}
	for _, r := range tr.Magic {
		in(r)
	}
	for _, k := range tr.Kinds {
		if k.Magic != nil {
			in(*k.Magic)
		}
	}
}

func TestPorTreasure(t *testing.T) { treasureCovered(t, PorTreasure()) }
