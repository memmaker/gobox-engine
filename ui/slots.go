// slots.go is the presenter's save slots (F10, the party menu's Load):
// the game's SlotsView lists them while they are up (slotsView), and Save,
// DeleteSlot and Load act.

package ui

// The slot screen's own decisions (sendA with the slot's name as Text),
// made calls by slotAct.
const ActSaveSlot, ActDeleteSlot ActKind = -3, -4

// ActSaveGame puts the slot screen up to save (the party menu's Save Current *classicGame).
const ActSaveGame ActKind = -11

// slotsView is the slot screen up as the game lists it (nil: none up, or
// none offered now).
func (p *Presenter) slotsView() *Slots {
	if p.slots == nil || p.eng == nil {
		return nil
	}
	k := slotsFrom{p.eng.G.Version(), p.eng.r.ran, p.slots.LoadOnly}
	return p.lay.slots.get(k, func() *Slots { return p.eng.SlotsView(k.loadOnly) })
}

// slotsFrom is what the slot screen is listed from: the game's State, the
// steps done (they may write the saves) and whether it only loads.
type slotsFrom struct {
	version, ran uint64
	loadOnly     bool
}

// SlotsUI is the slot screen up.
type SlotsUI struct{ LoadOnly bool }

// SlotsAct takes a decision a on the slot screen up (*slots): false
// leaves it to the game.
func SlotsAct(h RuleHost, slots **SlotsUI, a FrontAction) bool {
	sl := *slots
	switch name := a.Text; a.Kind {
	case ActClose:
		*slots = nil
	case ActSaveSlot:
		h.Call(func(e *classicGame) { refused(e, e.Save(name)) })
	case ActDeleteSlot:
		h.Call(func(e *classicGame) { refused(e, e.DeleteSlot(name)) })
	case ActLoad:
		h.Call(func(e *classicGame) {
			if e.Load(name, sl.LoadOnly) {
				*slots = nil
			}
		})
	default:
		return false
	}
	return true
}
