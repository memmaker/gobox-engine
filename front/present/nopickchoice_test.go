package present

import (
	"testing"

	"gobox/front"
	"gobox/input"
)

// TestNoPickChoice: NoPick and NoEsc hold on a choice as on a yes/no:
// Enter and Esc do nothing, the hotkey answers.
func TestNoPickChoice(t *testing.T) {
	v := front.View{Ask: &front.Ask{Kind: front.AskMenu, NoPick: true, NoEsc: true,
		Menu: front.Menu{Commands: []front.Command{{Text: "Yes", Key: 'Y'}, {Text: "No", Key: 'N'}}}}}
	p := &Presenter{}
	for _, k := range []uint16{keyOf(input.Enter), keyOf(input.Esc), keyOf(input.Down)} {
		if !p.pageMenuKey(k, v) || len(p.inv.acts) != 0 {
			t.Fatalf("key %d answered: %v", k, p.inv.acts)
		}
	}
	p.pageMenuKey('n', v)
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActChoose || p.inv.acts[0].N != 1 {
		t.Fatalf("N: %v", p.inv.acts)
	}
}
