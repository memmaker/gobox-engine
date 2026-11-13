package rules1e

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestStartData: rules.json's creationStart reaches Tables.Start, and a
// rolled character follows it (Felix's R3 rulings 2, 3, 7, 9): platinum
// fixed, experience trained silently past level 1, hit points at least 1,
// a cleric's two first-level slots.
func TestStartData(t *testing.T) {
	start := StartData{
		Money:      StartMoney{Fixed: 1500},
		Experience: []int32{25000, 12500, 8333},
	}
	b, err := json.Marshal(map[string]any{"thac0Table": POR.Thac0, "creationStart": start})
	if err != nil {
		t.Fatal(err)
	}
	tab, err := POR.WithRules(strings.NewReader(string(b)), Layout{SlotStride: saveCount})
	if err != nil {
		t.Fatal(err)
	}
	if tab.Start.Money.Fixed != 1500 || len(tab.Start.Experience) != 3 {
		t.Fatalf("creationStart not read: %+v", tab.Start)
	}
	e := &RuleState{Tables: &tab}
	c := e.Fresh(Kind{Race: RaceHuman, Class: ClsFighter}, nil)
	e.RollCharacter(c)
	if c.Coins[CoinGold] != 1500 || c.Exp != 25000 || c.Levels[ClsFighter] <= 1 || c.MaxHP < 1 {
		t.Fatalf("fighter: gold %d, exp %d, level %d, hp %d", c.Coins[CoinGold], c.Exp, c.Levels[ClsFighter], c.MaxHP)
	}
	p := &RuleState{}
	cl := p.Fresh(Kind{Race: RaceHuman, Class: ClsCleric}, nil)
	cl.Stats[StatWis] = 9 // no wisdom bonus
	p.FirstLevel(cl)
	if cl.SpellSlots[SpellClassCleric][0] != startClericSlots || cl.Exp != 0 {
		t.Fatalf("POR cleric: slots %d, exp %d", cl.SpellSlots[SpellClassCleric][0], cl.Exp)
	}
}
