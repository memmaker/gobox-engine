package engine

import "testing"

// A note from before a battle (camp's "rudely interrupted") is gone once
// the battle is set up, as the original clears its message row (BUG-CLASSES 10).
func TestBattleClearsNote(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.say("The Party is rudely interrupted!")
	e.combatSetup()
	if n, _ := e.Note(); n != "" {
		t.Fatalf("note %q over the battle", n)
	}
}
