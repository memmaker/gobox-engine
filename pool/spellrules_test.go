package pool

import (
	"rules1e"
	"strings"
	"testing"
)

// TestSpellRulesAligned: every spell's rules start with its own name.
func TestSpellRulesAligned(t *testing.T) {
	for id := 1; id < len(rules1e.SpellNames); id++ {
		if !strings.HasPrefix(strings.ToLower(SpellRules[id]), strings.ToLower(rules1e.SpellNames[id])+":") {
			t.Errorf("spell %d %q: rules %q", id, rules1e.SpellNames[id], SpellRules[id])
		}
	}
}

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
