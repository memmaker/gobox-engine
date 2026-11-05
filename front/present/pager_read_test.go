package present

import (
	"testing"

	"gobox/front"
)

// The pager stays on its page while the game's Read is up (the same pointer
// every View), and closes once it is gone.
func TestSyncReadKeepsPage(t *testing.T) {
	r := &front.Read{Title: "J", Lines: []string{"a", "b"}}
	p := &Presenter{}
	p.syncRead(front.View{Read: r})
	p.pager.P = 1
	p.syncRead(front.View{Read: r})
	if p.pager == nil || p.pager.P != 1 {
		t.Fatalf("pager reset: %+v", p.pager)
	}
	p.syncRead(front.View{})
	if p.pager != nil {
		t.Errorf("pager kept after the read ended")
	}
}
