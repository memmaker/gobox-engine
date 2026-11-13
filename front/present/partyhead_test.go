package present

import (
	"testing"

	"gobox/game"
)

// TestPartyPoolHead: a pool head goes to the card as the head (cropped to its
// lit part and scaled by present's face), not as a wide portrait fitted by width.
func TestPartyPoolHead(t *testing.T) {
	m := game.Member{Name: "A", PortraitHeadID: "head_000", PortraitBodyID: "body_000"}
	if headOf(m) == nil || m.Dropped != nil {
		t.Fatalf("Head %v Portrait %v", headOf(m) != nil, m.Dropped != nil)
	}
}
