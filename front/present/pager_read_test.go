package present

import (
	"testing"

	"gobox/front"
)

// A View's Read is built anew each frame: the pager stays on its page.
func TestSyncReadKeepsPage(t *testing.T) {
	lines := []string{"a", "b"}
	p := &Presenter{}
	p.syncRead(front.View{Read: &front.Read{Title: "J", Lines: lines}})
	p.pager.P = 1
	p.syncRead(front.View{Read: &front.Read{Title: "J", Lines: lines}})
	if p.pager == nil || p.pager.P != 1 {
		t.Fatalf("pager reset: %+v", p.pager)
	}
	p.syncRead(front.View{})
	if p.pager != nil {
		t.Errorf("pager kept after the read ended")
	}
}
