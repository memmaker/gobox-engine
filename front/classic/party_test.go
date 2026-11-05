package classic

import (
	"testing"

	"gobox/front"
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

// TestInventoryListsReadied: a readied item is on the doll and in the
// list, marked readied, so the keys reach it.
func TestInventoryListsReadied(t *testing.T) {
	s := game.State{Party: []game.Member{{Items: []game.Item{{Name: "Gun", Kind: "weapon", Slot: "weapon", Readied: true}}}}}
	in := Inventory(&front.Inventory{}, s)
	if in.Doll[0] == nil || len(in.Pool) != 1 || !in.Pool[0].Readied {
		t.Fatalf("doll %v pool %+v", in.Doll[0], in.Pool)
	}
}

// A yes/no over the 3D view with its question in the message is still asked.
func TestYesNoOverExplore(t *testing.T) {
	q, ok := Ask(game.State{Mode: game.ModeExplore, Pending: &game.Ask{Kind: game.YesNo}})
	if !ok || q == nil || q.Kind != front.AskConfirm || q.Prompt == "" {
		t.Fatalf("%+v", q)
	}
}
