package present

import (
	"testing"

	"gobox/game"
)

func TestSetGameData(t *testing.T) {
	old := spellRules
	defer func() { spellRules = old }()
	SetGameData(game.Data{SpellRules: map[int]string{3: "Cure Light Wounds: heals."}})
	if len(spellRules) != 4 || spellRules[3] != "Cure Light Wounds: heals." || spellRules[1] != "" {
		t.Errorf("spellRules %q", spellRules)
	}
}
