// layout.go: the page screens (the creation sheet, the grid pickers, the
// combat icon editor) laid out from the state on the presenter's page. Each
// layout is one function of the state's build (Version), the page's size and
// the page's palette, kept until one of them changes: nothing is laid out
// again while the game waits.

package present

import (
	"fmt"
	"image"
	"image/color"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// memo is a value derived from the state, kept while its key stays.
type memo[K comparable, V any] struct {
	key K
	val V
	ok  bool
}

// get is the value for key: built by build when key is not the last one's.
func (m *memo[K, V]) get(key K, build func() V) V {
	if !m.ok || m.key != key {
		m.key, m.val, m.ok = key, build(), true
	}
	return m.val
}

// uiKey is what the screens' derivations read besides the game's State:
// its build (Version) and the front's own state over it.
type uiKey struct {
	version                   uint64
	items                     game.ItemsUp
	itemsOf, sheet, lookSheet int
	ask, refusal              *play.Ask
	speedUp                   bool
}

func (p *Presenter) uiKey() uiKey {
	r := p.eng.r
	return uiKey{p.eng.G.Version(), r.items, r.itemsOf, r.sheet, r.lookSheet, r.ask, r.refusalAsk, r.speedUp}
}

// menuKey is what the party menu is made of.
type menuKey struct {
	ui                 uiKey
	ran                int // the art changed (F12, a new series): the logo
	training, creating bool
	slots, picker      bool
	selected           int
}

// layouts are the page screens the state puts up, laid out, and the
// screens' facts read from it.
type layouts struct {
	inventory memo[uiKey, *invUp]
	sheet     memo[uiKey, *game.SheetPage]
	sheetUp   memo[uiKey, bool]
	menu      memo[menuKey, front.MainMenu]
	creation  memo[creationKey, *front.Creation]
	picker    memo[pageKey, *front.Picker]     // the game's own picker
	icon      memo[pageKey, *front.IconEditor] // the game's own icon editor
}

// creationKey is what the creation sheet's layout is made of.
type creationKey struct {
	version uint64
	grid    image.Point
	ask     *play.Ask // gobox's own question (Abandon's words)
	refusal *play.Ask
}

// pageKey is what a picker's or the icon editor's layout is made of.
type pageKey struct {
	version uint64
	ran     int // the picker's and editor's calls run
	grid    image.Point
	ink     [16]color.RGBA
	part    int // the icon editor's swatch selected
	shade   int
}

// creation is the creation sheet laid out on the page (nil: none up, or a
// picker over it).
func (p *Presenter) creation() *front.Creation {
	if p.pickerOver() {
		return nil
	}
	return p.layoutCreation()
}

// layoutCreation is the creation sheet laid out on the page, a picker over it or not.
func (p *Presenter) layoutCreation() *front.Creation {
	if p.eng == nil {
		return nil
	}
	r, s := p.eng.r, p.state()
	k := creationKey{p.eng.G.Version(), r.grid, r.ask, r.refusalAsk}
	return p.lay.creation.get(k, func() *front.Creation {
		return Creation(r.modeOf(s), creationOf(s), r.pending(s), r.grid.X, r.grid.Y)
	})
}

// picker is the grid picker up laid out, in the page's palette ink (nil:
// none): the presenter's own, else the game's.
func (p *Presenter) picker(ink [16]color.RGBA) *front.Picker {
	if p.eng == nil {
		return nil
	}
	k := pageKey{version: p.eng.G.Version(), ran: p.ran, grid: p.eng.r.grid, ink: ink}
	if pk := p.gridPick; pk != nil {
		return pk.Layout(p.eng, k)
	}
	s := p.state()
	return p.lay.picker.get(k, func() *front.Picker {
		return Picker(play.CommonOf(s).Picker, image.Point{}, 0, p.eng.r.grid.X, p.eng.r.grid.Y)
	})
}

// iconEditor is the combat icon editor up laid out in the page's palette
// ink (nil: none, or a picker over it).
func (p *Presenter) iconEditor(ink [16]color.RGBA) *front.IconEditor {
	if p.eng == nil || p.gridPick != nil {
		return nil
	}
	k := pageKey{version: p.eng.G.Version(), ran: p.ran, grid: p.eng.r.grid, ink: ink}
	if ic := p.icon; ic != nil {
		return ic.Layout(p.eng, k)
	}
	s := p.state()
	return p.lay.icon.get(k, func() *front.IconEditor {
		return IconEditor(play.CommonOf(s).IconEditor, nil, p.eng.Subject(), &ink, p.eng.r.grid.X, p.eng.r.grid.Y)
	})
}

// stateScreens names the page screens the state puts up (Screens): the
// inventory, the sheet, the creation sheet, the picker and the icon editor.
func (p *Presenter) stateScreens() string {
	if p.eng == nil {
		return ""
	}
	ink := p.shownView.Ink
	return fmt.Sprint(p.inventoryKind(), p.sheetShown(), p.sheetPage() != nil, p.creation() != nil, p.picker(ink) != nil, p.iconEditor(ink) != nil)
}
