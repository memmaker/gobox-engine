package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestConfirmsDropMember: the game's DropMember confirm asks its texts one
// after another, on its default, and drops only after the last Yes.
func TestConfirmsDropMember(t *testing.T) {
	confirms = map[string]game.Confirm{"DropMember": {Asks: []string{"Drop %s?", "Really drop %s?"}}}
	defer func() { confirms = nil }()
	p := &Presenter{}
	v := front.View{Camp: &front.Camp{}, Party: []front.Member{{Name: "Ann", Selected: true}}}
	p.memberAct(front.ActDropMember, 0)
	p.inv.acts = nil
	title, rows := p.campItems(v)
	if title != "Drop Ann?" || p.camp.sel != 1 {
		t.Fatalf("first ask %q on row %d", title, p.camp.sel)
	}
	rows[0].Do()
	title, rows = p.campItems(v)
	if title != "Really drop Ann?" || p.camp.sel != 1 || len(p.inv.acts) != 0 {
		t.Fatalf("second ask %q on row %d, sent %v", title, p.camp.sel, p.inv.acts)
	}
	rows[0].Do()
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActDropMember || p.camp.screen != campMain {
		t.Fatalf("after both: %v screen %d", p.inv.acts, p.camp.screen)
	}
}

// TestConfirmText: %s is the name, %p the price.
func TestConfirmText(t *testing.T) {
	if s := confirmText("Sell %s for %p?", "a sword", "5 gp"); s != "Sell a sword for 5 gp?" {
		t.Fatal(s)
	}
}
