// slots.go is the presenter's save slots (F10, the party menu's Load):
// the game's SlotsView lists them while they are up (slotsView), and Save,
// DeleteSlot and Load act.

package present

import "gobox/front"

// The slot screen's own decisions (sendA with the slot's name as Text),
// made calls by slotAct.
const ActSaveSlot, ActDeleteSlot front.ActKind = -3, -4

// ActSaveGame puts the slot screen up to save (the party menu's Save Current *classicGame).
const ActSaveGame front.ActKind = -11

// slotsView is the slot screen up as the game lists it (nil: none up, or
// none offered now).
func (p *Presenter) slotsView() *front.Slots {
	if p.slots == nil || p.eng == nil {
		return nil
	}
	k := slotsFrom{p.eng.G.Version(), p.eng.r.ran, p.slots.LoadOnly}
	if p.slotsMade.key != k || !p.slotsMade.ok {
		p.slotsMade.v, p.slotsMade.key, p.slotsMade.ok = p.eng.SlotsView(k.loadOnly), k, true
	}
	return p.slotsMade.v
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
func SlotsAct(h RuleHost, slots **SlotsUI, a front.Action) bool {
	sl := *slots
	switch name := a.Text; a.Kind {
	case front.ActClose:
		*slots = nil
	case ActSaveSlot:
		h.Call(func(e *classicGame) { refused(e, e.Save(name)) })
	case ActDeleteSlot:
		h.Call(func(e *classicGame) { refused(e, e.DeleteSlot(name)) })
	case front.ActLoad:
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
