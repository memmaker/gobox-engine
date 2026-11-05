package classic

import (
	"testing"

	"gobox/game"
)


// TestPartyIconHead: a member with no portrait (Buck Rogers) shows their
// combat icon on the card, never an empty frame.
func TestPartyIconHead(t *testing.T) {
	old := Series
	defer func() { Series = old }()
	Series = "buck-rogers"
	ms := Party(game.State{Party: []game.Member{{Name: "Buck", IconBodyID: "combat_icon_000"}}}, nil)
	if len(ms) != 1 || ms[0].Head == nil {
		t.Fatal("no face for an icon-only member")
	}
}
