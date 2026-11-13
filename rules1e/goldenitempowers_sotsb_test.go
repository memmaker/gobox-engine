package rules1e

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"gobox/rules"
)

// TestGoldenItemPowersSotsb: readying and then unreadying an item with a
// power (giant strength 0..8, strength, grants an affect) gives a character
// the strength and affects sotsb's own item code gave (captured before its
// rules/ went); rules1e's handler for the power plays the same. Codes and
// affect numbers in the file are sotsb's own.

const (
	goldenSotsbPowerFlag     = 128 // an item power code is this plus the effect
	goldenSotsbGrantsAffect  = goldenSotsbPowerFlag + 0
	goldenSotsbStrengthSpell = 38 // sotsb's strength spell affect: rules1e keeps the natural strength in it
	goldenSotsbWizardryPower = goldenSotsbPowerFlag + 1
)

type goldenItemPowerCase struct {
	Code, Arg, Str, Str00 byte
	Ready                 bool
	WantStr, WantStr00    byte
	WantAffects           []int
}

func TestGoldenItemPowersSotsb(t *testing.T) {
	dir := filepath.Join("testdata", "sotsb")
	var cases []goldenItemPowerCase
	if err := readGolden(filepath.Join(dir, "itempowers.golden.json.gz"), &cases); err != nil {
		t.Fatal(err)
	}
	tb := loadGoldenTables(t, dir)
	var c *Character
	var it *Item
	for _, k := range cases {
		id, ok := SotsbItemPowers[k.Code]
		if !ok {
			t.Fatalf("code %d has no unified power", k.Code)
		}
		if k.Ready { // a new character and item for each ready, its unready following
			c = &Character{}
			c.Stats[StatStr], c.Str00 = Score(k.Str), k.Str00
			it = &Item{}
			it.Spells[1], it.Spells[2], it.Readied = ItemPower(k.Arg), id, 1
			if k.Code == goldenSotsbGrantsAffect {
				it.Spells[1] = ItemPower(sotsbAffect(AffectType(k.Arg)))
			}
			c.Items = []*Item{it}
		}
		r := &RuleState{Tables: &tb, Party: []*Character{c}}
		var flag byte
		if !k.Ready {
			it.Readied, flag = 0, 1
		}
		r.CallAffect(AffectType(id), c, &Affect{Item: it}, flag)
		if got, want := fmt.Sprint(c.Stats[StatStr], c.Str00), fmt.Sprint(k.WantStr, k.WantStr00); got != want {
			t.Errorf("code %d arg %d str %d/%d ready %v: strength %s, golden %s", k.Code, k.Arg, k.Str, k.Str00, k.Ready, got, want)
		}
		var got, want []AffectType
		for _, a := range c.Affects {
			if a.Type != sotsbAffect(goldenSotsbStrengthSpell) { // rules1e keeps the natural strength in a Strength affect; sotsb in Base
				got = append(got, a.Type)
			}
		}
		for _, n := range k.WantAffects {
			want = append(want, sotsbAffect(AffectType(n)))
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("code %d arg %d str %d/%d ready %v: affects %v, golden %v", k.Code, k.Arg, k.Str, k.Str00, k.Ready, got, want)
		}
	}
	t.Logf("%d cases", len(cases))
}

// TestWizardrySotsb: the ring of wizardry (power 129) doubles every
// magic-user spell level in rules1e (Felix 08:10, POR's ClassStats), where
// SotSB doubled the fifth only; unreadying it gives the slots back.
func TestWizardrySotsb(t *testing.T) {
	tb := loadGoldenTables(t, filepath.Join("testdata", "sotsb"))
	id := AffectType(SotsbItemPowers[goldenSotsbWizardryPower])
	r := &RuleState{Tables: &tb}
	c := &Character{}
	c.Levels[ClsMage], c.Stats[StatInt] = 5, 18
	it := &Item{}
	it.Spells[2], it.Readied = ItemPower(id), 1
	c.Items = []*Item{it}
	base := r.T().ClassStats(Who{Levels: map[rules.Class]int{ClassMage: 5}, Int: 18}).MageSlots
	r.CallAffect(id, c, &Affect{Item: it}, 0)
	for l, n := range base {
		if int(c.SpellSlots[SpellClassMage][l]) != 2*n {
			t.Errorf("ring readied: level %d slots %d, want %d", l+1, c.SpellSlots[SpellClassMage][l], 2*n)
		}
	}
	it.Readied = 0
	r.CallAffect(id, c, &Affect{Item: it}, 1)
	for l, n := range base {
		if int(c.SpellSlots[SpellClassMage][l]) != n {
			t.Errorf("ring unreadied: level %d slots %d, want %d", l+1, c.SpellSlots[SpellClassMage][l], n)
		}
	}
}
