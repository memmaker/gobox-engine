package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// The menu names its question; View shows the own ship's sheet and takes no turn; the foe's sheet and the
// panel's mounts come from the ruleset; a nil generator draws from the ruleset's own stream.
func TestShipView(t *testing.T) {
	var r *Rules
	var sf rules.ShipFight
	var a rules.CharID
	for seed := uint32(1); ; seed++ { // a seed where the party moves first: its menu is up at once
		if r, sf, a, _ = shipFightOf(t, seed); !r.space.enemyFirst {
			break
		}
	}
	r.seed = rules.RNG{State: 5}
	sf, _ = r.Ships().ShipRound(sf, nil) // nil: the ruleset's own stream
	al := r.Allowed(sf)
	if al.Prompt != "what does Ace want to do?" {
		t.Fatalf("prompt %q", al.Prompt)
	}
	o, err := r.Resolve(sf, rules.ShipView{}, nil)
	if err != nil || len(o.Effects) != 1 || o.Effects[0] != (rules.ShipSheetShown{Known: true}) {
		t.Fatalf("view %+v %v", o, err)
	}
	if r.crewActor() != a {
		t.Fatal("View took the turn")
	}
	own, foe := r.Ships().ShipSheetOf(false), r.Ships().ShipSheetOf(true)
	if own.Name != "Maelstrom Rider" || foe.Name != "Hubble" || foe.Race != "Martian" || len(foe.Mounts) != 8 || foe.Mounts[1].Short != "Missile" {
		t.Fatalf("sheets %+v %+v", own, foe)
	}
	if st := r.Ships().ShipState(); st.Own.Mounts[0].Short != "K-cannon" {
		t.Fatalf("state %+v", st.Own.Mounts)
	}
}
