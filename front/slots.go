// slots.go is the presenter's save slots (F10, the party menu's Load):
// the game's SlotsView fills View.Slots while they are up, and Save,
// DeleteSlot and Load act.

package front

// The slot screen's own decisions (sendA with the slot's name as Text),
// made calls by slotAct.
const ActSaveSlot, ActDeleteSlot ActKind = -3, -4

// ActSaveGame puts the slot screen up to save (the party menu's Save Current Game).
const ActSaveGame ActKind = -11

// SlotsUI is the slot screen up.
type SlotsUI struct{ LoadOnly bool }

// SlotsAct takes a decision a on the slot screen up (*slots): false
// leaves it to the game.
func SlotsAct(h Host, slots **SlotsUI, a Action) bool {
	sl := *slots
	switch name := a.Text; a.Kind {
	case ActClose:
		*slots = nil
	case ActSaveSlot:
		h.Call(func(e Game) { refused(e, e.Save(name)) })
	case ActDeleteSlot:
		h.Call(func(e Game) { refused(e, e.DeleteSlot(name)) })
	case ActLoad:
		h.Call(func(e Game) {
			if e.Load(name, sl.LoadOnly) {
				*slots = nil
			}
		})
	default:
		return false
	}
	return true
}
