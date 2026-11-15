// layout.go: the page screens (the creation sheet, the grid pickers, the
// combat icon editor) laid out from the state on the presenter's page. Each
// layout is one function of the state's build (Version), the page's size and
// the page's palette, kept until one of them changes: nothing is laid out
// again while the game waits.

package ui

import (
	"fmt"
	"image"
	"image/color"

	"gobox/game"
)

// uiKey is what the screens' derivations read besides the game's State:
// its build (Version) and the front's own state over it.
type uiKey struct {
	version                   stateBuild
	items                     game.ItemsUp
	itemsOf, sheet, lookSheet int
	ask, refusal              *game.Ask
	speedUp                   bool
}

func (p *Presenter) uiKey() uiKey {
	r := p.eng.r
	return uiKey{p.eng.r.build(), r.items, r.itemsOf, r.sheet, r.lookSheet, r.ask, r.refusalAsk, r.speedUp}
}

// menuKey is what the party menu is made of.
type menuKey struct {
	ui                 uiKey
	training, creating bool
	slots, picker      bool
	selected           int
}

// layouts are the page screens the state puts up, laid out, and the
// screens' facts read from it.
type layouts struct {
	slots     memo[slotsFrom, *Slots]         // the slot screen as the game lists it
	combat    memo[fightSource, *game.Combat] // the live fight read from the State
	board     memo[boardKey, fightDrawn]      // the drawn fight board
	party     memo[stateBuild, []game.Member] // partyOf, by the State's build
	powers    memo[stateBuild, [][]Power]     // each card's context menu
	menuChars memo[stateBuild, int]           // the widest line of them
	inventory memo[uiKey, *invUp]
	sheet     memo[uiKey, *game.SheetPage]
	sheetUp   memo[uiKey, bool]
	menu      memo[menuKey, MainMenu]
	creation  memo[creationKey, *FrontCreation]
	picker    memo[pageKey, *FrontPicker]     // the game's own picker
	icon      memo[pageKey, *FrontIconEditor] // the game's own icon editor
	page      memo[laidKey, *laidPage]        // the text page laid out (shown.go)
	music     memo[uiKey, string]             // the own music's trigger playing, wanted on the state up (Front.Draw)
	ask       memo[uiKey, *FrontAsk]          // the state's question, or the door's menu
	asks      memo[uiKey, string]             // the game's question that wants an answer, named (asksID)
	training  memo[uiKey, *FrontTraining]
	flags     memo[flagsKey, *game.Flags]
}

// flagsKey is what the status flags are read from: the build and the title's step.
type flagsKey struct {
	version stateBuild
	title   titleStep
}

// creationKey is what the creation sheet's layout is made of.
type creationKey struct {
	version stateBuild
	grid    image.Point
	ask     *game.Ask // gobox's own question (Abandon's words)
	refusal *game.Ask
}

// pageKey is what a picker's or the icon editor's layout is made of.
type pageKey struct {
	version stateBuild
	ran     int // the picker's and editor's calls run
	grid    image.Point
	ink     [16]color.RGBA
	part    int // the icon editor's swatch selected
	shade   int
}

// creation is the creation sheet laid out on the page (nil: none up, or a
// picker over it).
func (p *Presenter) creation() *FrontCreation {
	if p.pickerOver() {
		return nil
	}
	return p.layoutCreation()
}

// layoutCreation is the creation sheet laid out on the page, a picker over it or not.
func (p *Presenter) layoutCreation() *FrontCreation {
	if p.eng == nil {
		return nil
	}
	r, s := p.eng.r, p.state()
	g := r.grid
	k := creationKey{p.eng.r.build(), g, r.ask, r.refusalAsk}
	return p.lay.creation.get(k, func() *FrontCreation {
		return Creation(r.modeOf(s), creationOf(s), r.pending(s), g.X, g.Y)
	})
}

// picker is the grid picker up laid out, in the page's palette ink (nil:
// none): the presenter's own, else the game's.
func (p *Presenter) picker(ink [16]color.RGBA) *FrontPicker {
	if p.eng == nil {
		return nil
	}
	k := pageKey{version: p.eng.r.build(), ran: p.ran, grid: p.eng.r.grid, ink: ink}
	if pk := p.gridPick; pk != nil {
		return pk.Layout(p.eng, k)
	}
	s := p.state()
	return p.lay.picker.get(k, func() *FrontPicker {
		return Picker(game.CommonOf(s).Picker, 0, p.eng.r.grid.X, p.eng.r.grid.Y)
	})
}

// iconEditor is the combat icon editor up laid out in the page's palette
// ink (nil: none, or a picker over it).
func (p *Presenter) iconEditor(ink [16]color.RGBA) *FrontIconEditor {
	if p.eng == nil || p.gridPick != nil {
		return nil
	}
	k := pageKey{version: p.eng.r.build(), ran: p.ran, grid: p.eng.r.grid, ink: ink}
	if ic := p.icon; ic != nil {
		return ic.Layout(p.eng, k)
	}
	s := p.state()
	return p.lay.icon.get(k, func() *FrontIconEditor {
		return IconEditor(game.CommonOf(s).IconEditor, nil, p.eng.Subject(), &ink, p.eng.r.grid.X, p.eng.r.grid.Y)
	})
}

// stateScreens names the page screens the state puts up (Screens): the
// inventory, the sheet, the creation sheet, the picker and the icon editor.
func (p *Presenter) stateScreens() string {
	if p.eng == nil {
		return ""
	}
	ink := p.ink()
	return fmt.Sprint(p.inventoryKind(), p.sheetShown(), p.sheetPage() != nil, p.creation() != nil, p.picker(ink) != nil, p.iconEditor(ink) != nil)
}

// party is the party as the screens show it (partyOf), kept per State.
func (p *Presenter) party() []game.Member {
	if p.eng == nil {
		return nil
	}
	return p.lay.party.get(p.eng.r.build(), func() []game.Member { return partyOf(p.state()) })
}

// cardPowers are the cards' context menus, by member.
func (p *Presenter) cardPowers() [][]Power {
	if p.eng == nil {
		return nil
	}
	return p.lay.powers.get(p.eng.r.build(), func() (out [][]Power) {
		for _, c := range cardsOf(p.state()) {
			out = append(out, powers(c.Menu))
		}
		return out
	})
}

// cardMenuChars is the widest line of the cards' context menus.
func (p *Presenter) cardMenuChars() int {
	if p.eng == nil {
		return 0
	}
	return p.lay.menuChars.get(p.eng.r.build(), func() (n int) {
		for _, ps := range p.cardPowers() {
			lines, _, _ := powerLines(ps)
			for _, l := range lines {
				n = max(n, len(l))
			}
		}
		return n
	})
}
