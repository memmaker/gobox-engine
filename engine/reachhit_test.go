package engine

import (
	"image"
	"testing"
)

// TestReachAndHitChance: the party's turn shows its move reach with a walk
// to every square (c70cbee moveReach, walkSteps) and each target's chance to
// hit in the hover text (c70cbee hitChance into Unit.Hit).
func TestReachAndHitChance(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	b := e.battle()
	if b == nil || len(b.Reach) == 0 {
		t.Fatal("no move reach on the party's turn")
	}
	at := image.Pt(e.state().CombatX(c), e.state().CombatY(c))
	for _, sq := range b.Reach {
		p := b.Paths[sq]
		if len(p) < 2 || p[0].At.X != at.X || p[0].At.Y != at.Y || p[len(p)-1].At.X != sq.X || p[len(p)-1].At.Y != sq.Y {
			t.Fatalf("walk to %v: %v", sq, p)
		}
	}
	if len(b.Targets) == 0 {
		t.Skip("no target this turn")
	}
	for _, i := range b.Targets {
		for _, u := range b.Units {
			if u.ID == i && (u.Hit <= 0 || u.Hit > 100 || u.Hit != e.hitChance(e.charID(c), e.charID(e.state().Combatant(byte(i))))) {
				t.Fatalf("target %d: Hit %d", i, u.Hit)
			}
		}
	}
}
