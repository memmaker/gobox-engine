// picker.go is the presenter's grid pickers and combat icon editor: the
// engine readies a picker (OpenPicker) or the icon designed (EditIcon),
// lays their pages out (PickerView, IconEditorView) and takes the picks
// and changes (Pick, IconSize, IconColour, IconRemove, EndIcon); opening,
// closing, the pages, the cursor and the swatch selected are here.

package present

import "gobox/front"

// The picker's and the icon editor's own decisions (sendA), made calls by
// PickAct and IconAct.
const (
	ActPicker     front.ActKind = -6  // open picker N (PickHead ..)
	ActPicked     front.ActKind = -7  // the picker takes picture N (the player agreed to Picker.Ask)
	ActIconSize   front.ActKind = -8  // icon editor: small or large
	ActIconColour front.ActKind = -9  // icon editor: part N's colour X (0, 1) becomes Y (Y < 0: only selects that swatch)
	ActIconRemove front.ActKind = -10 // icon editor: the dropped PNG goes, the game's own icon is back
)

// PickUI is a grid picker up: its page (-1: the one it opened on) and
// the page laid out, kept until a call or the window changes it.
type PickUI struct {
	Page int
	view *front.Picker
	key  any
}

// IconUI is the icon editor up: the swatch the palette sets and the page
// laid out (as PickUI's).
type IconUI struct {
	part, shade int
	view        *front.IconEditor
	key         any
}

// page is the picker's page laid out on g, again when key (the calls run
// and the window) changed.
func (pk *PickUI) Layout(g *classicGame, key any) *front.Picker {
	if pk.view == nil || pk.key != key {
		pk.view, pk.key = g.PickerView(pk.Page), key
		if pk.view != nil {
			pk.Page = pk.view.Page
		}
	}
	return pk.view
}

// Layout is the editor's page laid out on g, with the swatch selected
// (nil: none).
func (ic *IconUI) Layout(g *classicGame, key any) *front.IconEditor {
	if ic.view == nil || ic.key != key {
		ic.view, ic.key = g.IconEditorView(), key
	}
	if ic.view == nil {
		return nil
	}
	ed := *ic.view
	ed.Part, ed.Shade = ic.part, ic.shade
	return &ed
}

// OpenPicker opens picker kind when it has something to pick.
func OpenPicker(h RuleHost, pick **PickUI, kind int) {
	h.Edit(func(e *classicGame) {
		if e.OpenPicker(kind) {
			*pick = &PickUI{Page: -1}
		}
	})
}

// PickAct takes a decision a on the picker up (*pick, or one opening it):
// false leaves it to the front.
func PickAct(h RuleHost, pick **PickUI, a front.Action) bool {
	switch pk := *pick; {
	case a.Kind == ActPicker:
		OpenPicker(h, pick, a.N)
	case pk == nil && a.Kind == front.ActAdd:
		OpenPicker(h, pick, front.PickAdd)
	case pk == nil && a.Kind == front.ActForget:
		OpenPicker(h, pick, front.PickForget)
	case pk == nil:
		return false
	case a.Kind == front.ActPage:
		pk.Page, pk.view = a.N, nil
	case a.Kind == ActPicked:
		h.Edit(func(e *classicGame) {
			if e.Pick(a.N) {
				*pick = nil
			}
		})
	case a.Kind == front.ActClose:
		*pick = nil
		h.Edit(func(e *classicGame) { e.ClosePicker() })
	default:
		return false
	}
	return true
}

// IconAct takes a decision a on the icon editor up (*icon, or one
// opening it): false leaves it to the front.
func IconAct(h RuleHost, icon **IconUI, a front.Action) bool {
	ic := *icon
	if ic == nil {
		if a.Kind != front.ActIconEdit {
			return false
		}
		h.Edit(func(e *classicGame) {
			e.EditIcon(a.N)
			*icon = &IconUI{}
		})
		return true
	}
	switch a.Kind {
	case ActIconSize:
		h.Edit(func(e *classicGame) { e.IconSize() })
	case ActIconColour:
		if a.N < 0 || a.N > 5 || a.X < 0 || a.X > 1 {
			return true
		}
		ic.part, ic.shade = a.N, a.X
		if a.Y >= 0 {
			h.Edit(func(e *classicGame) { e.IconColour(a.N, a.X, a.Y) })
		}
	case ActIconRemove:
		h.Edit(func(e *classicGame) { e.IconRemove() })
	case front.ActIcon: // a PNG dropped on it
		h.Edit(func(e *classicGame) { e.IconPNG(a.Text) })
	case front.ActKeep, front.ActClose:
		*icon = nil
		h.Edit(func(e *classicGame) { e.EndIcon(a.Kind == front.ActKeep) })
	default:
		return false
	}
	return true
}
