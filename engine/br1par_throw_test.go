package engine

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// brBuyGrenade buys a grenade at one of the shop lists the game's scripts open and readies it for the first member;
// a game whose lists hold none skips.
func brBuyGrenade(t *testing.T, e *Engine) {
	t.Helper()
	tab := br1Tables(t, e)
	who := e.partyIDs()[0]
	for list := 1; list <= 16; list++ {
		br1OwnBlock(t, e, 201+list, "    PriceMultiplier = 16\n    Shop("+strconv.Itoa(list)+")")
		tr, ok := e.build().(game.Trading)
		if !ok {
			continue
		}
		i := slices.IndexFunc(tr.Goods, func(g game.Item) bool {
			id := e.itemCard(rules.NoChar, g.Handle).Kind
			return tab.ItemTypes[tab.Items[id].Type].WeaponClass == xxvc.Grenade
		})
		if i < 0 {
			if _, err := e.doNow(game.Leave{}); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if _, err := e.ruleset().Resolve(nil, rules.SetCredits{Who: e.selected(), Total: 100000}, nil); err != nil {
			t.Fatal(err)
		}
		brMakeRoom(t, e, who)
		if _, err := e.doNow(game.Buy{Member: 0, Item: tr.Goods[i].Handle}); err != nil {
			t.Fatal("buy:", err)
		}
		if _, err := e.doNow(game.Leave{}); err != nil {
			t.Fatal(err)
		}
		items := e.listedItems(who)
		if _, err := e.ruleset().Resolve(nil, rules.Ready{Who: who, Item: items[len(items)-1]}, nil); err != nil {
			t.Fatal("ready:", err)
		}
		return
	}
	t.Skip("no shop list sells a grenade")
}

// TestBR1Throw ports rules TestThrow (cbtacts.pas CombatAimMenu with an area weapon, then the throw): a member with a
// grenade readied is offered a square (the aim cursor opens), and the square taken throws it there: the battle log
// says so, the flight is published as a game.Shot in COMSPR's art, and the grenade is spent from the stack.
func TestBR1Throw(t *testing.T) { brThrow(t, gameBR1) }

func TestBR2Throw(t *testing.T) { brThrow(t, gameBR2) }

func brThrow(t *testing.T, g brGame) {
	e := brBegun(t, g)
	brBuyGrenade(t, e)
	br1OwnBlock(t, e, 200, br1FarFight)
	for range br1FightCap {
		if myTurn(e.build()) {
			break
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	f, ok := e.build().(game.Fighting)
	if !ok || !myTurn(f) {
		t.Fatal("no party turn")
	}
	h := e.fact(e.unitID(f.Field.Turn), rules.WeaponFact{}).(rules.Handles)
	if len(h) == 0 {
		t.Skip("the readied grenade is no weapon of the turn (a launcher is needed)")
	}
	left := e.invItem(e.unitID(f.Field.Turn), h[0]).Count
	if err := e.TurnChoice(TurnAct{Verb: VerbAttack}, false); err != nil || e.cs.aim == nil {
		t.Fatalf("Aim opened no cursor: %v", err)
	}
	var me game.Unit
	for _, u := range f.Field.Units {
		if u.ID == f.Field.Turn {
			me = u
		}
	}
	at := rules.BoardPos{X: me.At.X + 3, Y: me.At.Y} // a square of the grenade's range (the areas it burst on are the ruleset's)
	e.events = nil
	if err := e.TurnChoice(TurnAct{Verb: VerbAim, X: at.X, Y: at.Y}, false); err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(e.ext.log.lines, func(l string) bool { return strings.Contains(l, " uses ") }) {
		t.Errorf("no throw in the log %q", e.ext.log.lines)
	}
	if !slices.ContainsFunc(e.events, func(ev game.Event) bool { _, ok := ev.(game.Shot); return ok }) {
		t.Error("the throw's flight is no event")
	}
	who := e.unitID(f.Field.Turn)
	if slices.Contains(e.listedItems(who), h[0]) { // the last of a stack is gone from the list
		if got := e.invItem(who, h[0]).Count; got >= left {
			t.Errorf("grenades %d, were %d", got, left)
		}
	}
}
