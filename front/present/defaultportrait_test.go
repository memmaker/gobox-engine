package present

import (
	"strings"
	"testing"

	"gobox/game"
)

func TestDefaultPortraitAndIcon(t *testing.T) {
	m := defaultIcon(defaultPortrait(game.Member{Race: "Elf", Class: "Fighter"}))
	if !strings.HasPrefix(m.PortraitHeadID, "head_") || !strings.HasPrefix(m.PortraitBodyID, "body_") ||
		!strings.HasPrefix(m.IconHeadID, "combat_head_") || !strings.HasPrefix(m.IconBodyID, "combat_body_") {
		t.Fatalf("no defaults: %+v", m)
	}
	if n := defaultPortrait(game.Member{Race: "Elf", Class: "Fighter"}); n.PortraitHeadID != m.PortraitHeadID {
		t.Fatal("default not stable")
	}
	own := defaultIcon(defaultPortrait(game.Member{PortraitHeadID: "head_001", PortraitBodyID: "body_002", Icon: map[string]int{"head": 1}}))
	if own.PortraitHeadID != "head_001" || own.IconHeadID != "" {
		t.Fatalf("picks overridden: %+v", own)
	}
}
