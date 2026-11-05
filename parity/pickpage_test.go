package parity

import (
	"testing"

	"gobox/front"
	"gobox/front/present"
)

// TestPickerPages: the pool head picker turns its page.
func TestPickerPages(t *testing.T) {
	e := partyMenu(t)
	p := &present.Presenter{}
	g := present.WithRules(gameOf(e), p)
	p.SetEngine(g)
	if err := e.CreateCharacter(); err != nil {
		t.Fatal(err)
	}
	settled(e, front.Action{})
	c := gameOf(e)
	step := func(as ...front.Action) *front.Picker {
		for _, a := range as {
			p.Decide(c.View(), a)
			for _, f := range p.Calls() {
				f(c)
			}
			for _, a := range p.Actions() {
				c.Answer(a)
			}
			c.Settle()
		}
		return p.Own(c.View()).Picker
	}
	v := step(front.Action{Kind: front.ActPicker, N: front.PickHead})
	if v == nil {
		t.Fatal("no picker")
	}
	t.Logf("heads %d pages %d", v.N, v.Pages)
	p.Key(0x5100, c.View())
	if v = step(); v.Page != 1 {
		t.Fatalf("PgDn: page %d", v.Page)
	}
}
