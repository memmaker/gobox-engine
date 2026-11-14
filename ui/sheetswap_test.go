package ui

import (
	"testing"

	"gobox/game"
)

// TestSheetItemsSwap: Tab swaps the sheet and the inventory in place, they
// never stack: Esc on either, however often Tab went back and forth,
// leaves both.
func TestSheetItemsSwap(t *testing.T) {
	ex := game.Exploring{Common: game.Common{Party: []game.Member{{Name: "ARIN", Page: &game.SheetPage{Name: "ARIN"}}}}}
	for _, startItems := range []bool{false, true} {
		r := newRules(&fake{st: ex}, game.Data{})
		if startItems {
			r.send(game.OpenItems{Member: 0})
		} else {
			r.send(tOpenSheet{Member: 0})
		}
		for range 3 { // Tab, Tab, Tab
			if r.items == game.ItemsOwn {
				r.send(tCloseItems{All: false})
			} else {
				r.send(game.OpenItems{Member: -1})
			}
		}
		if r.items == game.ItemsOwn {
			r.send(tCloseItems{All: true}) // Esc on the inventory
		} else {
			r.send(tCloseSheet{}) // Esc on the sheet
		}
		if r.items != game.ItemsNone || r.sheet >= 0 {
			t.Errorf("started on items %v: after Esc items %v sheet %d", startItems, r.items, r.sheet)
		}
	}
}
