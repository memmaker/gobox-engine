package rulesxxvc

import (
	"reflect"
	"testing"

	"gobox/rules"
)

// savedState says for each field of Rules whether the save holds it (saved) or why a save need not (the reason).
// A new field fails TestSaveCoversRules until it is listed here.
var savedState = map[string]string{
	"chars": "saved", "order": "saved", "next": "saved", "purse": "saved", "ship": "saved", "foeSaved": "saved",
	"party": "saved", "memberCount": "saved", "acting": "saved", "clock": "saved", "seed": "saved", "empty": "saved",
	"walls":    "saved",
	"nextItem": "derived on Load: the largest handle the loaded characters carry (TestLoadKeepsItemHandlesApart)",
	"campaign": "the game's tables given at Begin, not the game's state", "options": "given at Begin", "tables": "given at Begin",
	"pool":   "a fight's: the credits the beaten left, banked when it ends; a game is saved outside a fight",
	"loot":   "a fight's: the handles of its loot",
	"morale": "a fight's (CheckMorale)", "hpPercent": "a fight's", "terms": "a fight's: BeginFight's", "moveBonus": "a fight's",
	"duelist": "a fight's", "tactics": "a fight's", "round": "a fight's", "initiative": "a fight's", "fight": "a fight's",
	"arena": "a fight's", "ground": "a fight's, as rules1e's is not saved", "space": "a ship fight's",
	"cloning":   "between a NewGroup and its Spawns, inside one script command",
	"held":      "the copies a Modify sheet holds until Keep or Exit: the party menu's, no game is saved under it",
	"phase":     "the engine's (EnterPhase): entered again when a loaded game's loop starts",
	"shopTerms": "a shop's (StockShop): set each time a script opens one; a loaded game opens it again",
}

// TestSaveCoversRules: every field of Rules is saved or has its reason in savedState, and the saved ones come back through
// Save and Load.
func TestSaveCoversRules(t *testing.T) {
	typ := reflect.TypeFor[Rules]()
	for i := range typ.NumField() {
		f := typ.Field(i)
		if _, ok := savedState[f.Name]; !ok {
			t.Errorf("Rules.%s is neither saved nor given a reason in savedState", f.Name)
		}
	}
	for name := range savedState {
		if _, ok := typ.FieldByName(name); !ok {
			t.Errorf("savedState names %s, which Rules has not", name)
		}
	}
	// a Character is saved as JSON: an unexported field would be lost
	ct := reflect.TypeFor[Character]()
	for i := range ct.NumField() {
		if f := ct.Field(i); !f.IsExported() {
			t.Errorf("Character.%s is unexported: the save would drop it", f.Name)
		}
	}
	r := New()
	a := r.Add(Character{Status: StOkay, HP: 3, MaxHP: 10, Order: 4, ScriptFlags: [ScriptFlagSlots]int{1, 2, 3}, Stashed: []Carried{{Handle: 9}}})
	r.party, r.memberCount, r.acting, r.clock[1], r.purse = []rules.CharID{a}, 1, 0, 7, 55
	r.walls = rules.RecordView{WallsA: [3]uint16{7, 127, 127}}
	b, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	l := New()
	if err := l.Load(b); err != nil {
		t.Fatal(err)
	}
	c := l.chars[a]
	if c == nil || c.HP != 3 || c.Order != 4 || c.ScriptFlags != [ScriptFlagSlots]int{1, 2, 3} || len(c.Stashed) != 1 {
		t.Errorf("character %+v", c)
	}
	if len(l.party) != 1 || l.memberCount != 1 || l.clock[1] != 7 || l.purse != 55 || l.walls.WallsA[0] != 7 {
		t.Errorf("session party %v count %d clock %v purse %d walls %v", l.party, l.memberCount, l.clock, l.purse, l.walls)
	}
}

// TestLoadKeepsItemHandlesApart: after a Load, handOut gives a new item a handle no loaded item has.
func TestLoadKeepsItemHandlesApart(t *testing.T) {
	r := New()
	a := r.Add(Character{Items: []Carried{{}, {}}})
	r.handOut(r.chars[a])
	b, _ := r.Save()
	l := New()
	if err := l.Load(b); err != nil {
		t.Fatal(err)
	}
	c := l.chars[a]
	c.Items = append(c.Items, Carried{})
	l.handOut(c)
	seen := map[rules.ItemHandle]bool{}
	for _, it := range c.Items {
		if it.Handle == rules.NoItem || seen[it.Handle] {
			t.Fatalf("handles %+v", c.Items)
		}
		seen[it.Handle] = true
	}
}
