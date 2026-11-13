package engine

import (
	"gobox/rules1e"
	"testing"
)

// spellName never panics: 57-67 are item and monster powers with no name.
func TestSpellNameRange(t *testing.T) {
	e := &Engine{}
	e.initVM()
	for id := 0; id <= 67; id++ {
		if n := e.state().SpellName(id); id >= len(rules1e.SpellNames) && n != "" {
			t.Errorf("power %d named %q", id, n)
		}
	}
}
