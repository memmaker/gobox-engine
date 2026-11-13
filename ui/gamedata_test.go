package ui

import (
	"testing"

	"gobox/game"
)

func TestGameData(t *testing.T) {
	old := spellRules
	defer func() { spellRules = old }()
	gameData(game.Data{SpellRules: map[int]string{3: "Cure Light Wounds: heals."}})
	if len(spellRules) != 4 || spellRules[3] != "Cure Light Wounds: heals." || spellRules[1] != "" {
		t.Errorf("spellRules %q", spellRules)
	}
}
