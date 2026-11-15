package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// TestSetCredits: the party's credits written lower are paid by the selected member first, then the others in
// party order; written higher the difference goes to the selected one (BR ecl.go setChar OffMoney, e.pay).
func TestSetCredits(t *testing.T) {
	for _, shared := range []bool{false, true} {
		r := New()
		r.Begin(rules.Campaign{}, rules.Options{SharedPurse: shared})
		var ids [3]rules.CharID
		for i, coins := range []rules.Copper{5, 10, 20} {
			ids[i] = r.Add(Character{Name: "m", Enabled: true, Coins: coins})
			if _, err := r.Resolve(nil, rules.JoinParty{Who: ids[i]}, nil); err != nil {
				t.Fatal(err)
			}
		}
		purse := func() rules.Copper {
			v, err := r.Fact(ids[0], rules.PurseFact{})
			if err != nil {
				t.Fatal(err)
			}
			return rules.Copper(v.(rules.Purse).Copper)
		}
		set := func(who rules.CharID, total rules.Copper) {
			if _, err := r.Resolve(nil, rules.SetCredits{Who: who, Total: total}, nil); err != nil {
				t.Fatal(err)
			}
		}
		coins := func() [3]rules.Copper {
			return [3]rules.Copper{r.chars[ids[0]].Coins, r.chars[ids[1]].Coins, r.chars[ids[2]].Coins}
		}
		if purse() != 35 {
			t.Fatalf("shared %v: purse %d", shared, purse())
		}
		set(ids[1], 20) // 15 owed: the second pays its 10, then the first its 5
		want := [3]rules.Copper{0, 0, 20}
		if shared {
			want = coins() // the one purse holds it all
		}
		if purse() != 20 || coins() != want {
			t.Fatalf("shared %v: lower: purse %d coins %v", shared, purse(), coins())
		}
		set(ids[2], 27) // 7 more: to the selected third
		if purse() != 27 || (!shared && coins() != [3]rules.Copper{0, 0, 27}) {
			t.Fatalf("shared %v: higher: purse %d coins %v", shared, purse(), coins())
		}
		if _, err := r.Resolve(nil, rules.SetCredits{Who: ids[0], Total: -1}, nil); err == nil {
			t.Fatal("negative total accepted")
		}
	}
}
