package engine

import (
	"slices"
	"strings"
	"testing"

	"gobox/game"
)

// TestBR1SpaceMenuEsc (old spacefight_e2e_test.go TestSpaceMenuEsc, shipcbt.pas 1130-1150): Esc on the crew menu
// changes nothing, the menu stands (MenuHorizontal answers C_Keys.NONE, no case arm, isActionTaken stays False);
// Esc on 'Target which system?' is no system chosen and not RIGHT_BRACKET, so the volley goes untargeted.
func TestBR1SpaceMenuEsc(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, spaceFightScript)
	for range fightSteps { // walk the crew until a menu offers Target (a member with a weapon)
		f, ok := shipFightOn(e)
		if !ok {
			t.Fatal("the fight ended before a menu offered Target")
		}
		i := indexOf(f.Ask, "Target")
		if i < 0 {
			var a game.Action = game.ContinueAction{}
			if !f.Ask.Continue {
				a = game.Reply{N: max(indexOf(f.Ask, "Quit"), 0)}
			}
			if _, err := e.doNow(a); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if _, err := e.doNow(game.Leave{}); err == nil {
			t.Fatal("Esc on the crew menu was taken")
		}
		if again, _ := shipFightOn(e); again.Ask == nil || again.Ask.Text != f.Ask.Text || !slices.Equal(again.Ask.Replies, f.Ask.Replies) {
			t.Fatalf("the crew menu after Esc: %+v, want %+v", again.Ask, f.Ask)
		}
		if _, err := e.doNow(game.Reply{N: i}); err != nil {
			t.Fatal(err)
		}
		if sub, _ := shipFightOn(e); sub.Ask == nil || sub.Ask.Text != "Target which system?" || sub.Ask.NoEsc {
			t.Fatalf("the target menu: %+v", sub.Ask)
		}
		if _, err := e.doNow(game.Leave{}); err != nil {
			t.Fatalf("Esc on the target menu: %v", err)
		}
		after, _ := shipFightOn(e)
		fired := false
		for _, l := range after.Fight.Log {
			fired = fired || strings.Contains(l.Text, " attacks with a ")
		}
		if !fired {
			t.Fatalf("Esc on the target menu did not fire: log %+v ask %+v", after.Fight.Log, after.Ask)
		}
		return
	}
	t.Fatal("no menu offered Target")
}
