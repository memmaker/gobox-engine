package classic

import (
	"testing"

	"gobox/game"
)

// TestPartyPoolHead: a pool head goes to the card as Head (cropped to its lit
// part and scaled by present's face), not as a wide Portrait fitted by width.
func TestPartyPoolHead(t *testing.T) {
	ms := Party([]game.Member{{Name: "A", PortraitHeadID: "head_000", PortraitBodyID: "body_000"}}, 0)
	if len(ms) != 1 || ms[0].Head == nil || ms[0].Portrait != nil {
		t.Fatalf("Head %v Portrait %v", ms[0].Head != nil, ms[0].Portrait != nil)
	}
}
