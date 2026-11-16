// frontgame.go: the presenter is the screens' RuleHost; its rule calls run on
// the game, a *classicGame.

package ui

import (
	"fmt"
	"gobox/gpu"
	"image"
)

var _ RuleHost = (*Presenter)(nil)

// Call queues f (call).
func (p *Presenter) Call(f func(*classicGame)) { p.call(func(e *classicGame) { f(e) }) }

// Edit queues f (edit).
func (p *Presenter) Edit(f func(*classicGame)) { p.edit(func(e *classicGame) { f(e) }) }

// Ask puts q up (nil: none).
func (p *Presenter) Ask(q *Question) { p.q = q }

// Asking is the question up (nil: none).
func (p *Presenter) Asking() *Question { return p.q }

// Shown is what the last frame showed, as the screens see it.
func (p *Presenter) Shown() Shown {
	var t *FrontLoot
	if in := p.inventory(); in != nil {
		t = in.treasure
	}
	return Shown{Treasure: t, Ask: p.askUp() != nil, Powers: p.powerPick() != nil, Sheet: p.sheetShown(), Inventory: p.inventoryUp(), Battle: p.board != nil,
		Camp: p.campUp(), MainMenu: p.mainMenuUp(), Door: p.door(), Resting: p.pacing().Resting, Party: len(p.party())}
}

// questionAct takes a while the presenter's question is up, at the gate
// or while camp rests (AskAct); then the powers and the rule calls of the
// screens up; false leaves a to the others.
func (p *Presenter) questionAct(a Action) bool {
	if taken, done := AskAct(p, a); done {
		return taken
	}
	return PowerAct(p, a) || p.viewAct(a) || p.turnAct(a) || StateAct(p, a)
}

// edit queues f, after which the picker's and the editor's pages are
// laid out again.
func (p *Presenter) edit(f func(e *classicGame)) {
	p.call(func(e *classicGame) { f(e); p.ran++ })
}

// openSlots puts the slot screen up (F10, the menu's Save/Load), or takes
// it down when it is up; not over another page of the whole window.
func (p *Presenter) openSlots(loadOnly bool) {
	switch {
	case p.slots != nil:
		p.slots = nil
	case p.eng != nil && !p.modal() && p.eng.SlotsView(loadOnly) != nil:
		p.slots = &SlotsUI{LoadOnly: loadOnly}
	}
}

// OpenSlots is F10 from the renderer's options (Save/Load).
func (p *Presenter) OpenSlots() { p.openSlots(false) }

// slotAct takes a decision a on the slot screen: false leaves it to the
// engine.
func (p *Presenter) slotAct(a Action) bool {
	sl := p.slots
	if a.Kind == ActCreate || a.Kind == ActModify { // only a new character is asked Roll or Point Buy
		p.scr.newChar = a.Kind == ActCreate
	}
	if sl == nil {
		if p.gridPick != nil && PickAct(p, &p.gridPick, a) { // a picker up takes its own actions first: none of the screens under it pages or picks for it
			return true
		}
		return p.questionAct(a) || PickAct(p, &p.gridPick, a) || IconAct(p, &p.icon, a) || CreateAct(p, &p.creating, a) || p.serviceAct(a)
	}
	return SlotsAct(p, &p.slots, a)
}

// frontScreens names the front's screens the presenter has up (Screens).
func (p *Presenter) frontScreens() string {
	return Screens(
		ScreenUp{Name: ScreenSlots, On: p.slots != nil},
		ScreenUp{Name: ScreenTitle, On: p.intro.Phase != IntroNone, State: fmt.Sprint(p.intro.Phase, len(p.intro.Flash))},
		ScreenUp{Name: ScreenPicker, On: p.gridPick != nil},
		ScreenUp{Name: ScreenIconEditor, On: p.icon != nil},
		ScreenUp{Name: ScreenCreation, On: p.creating},
		ScreenUp{Name: ScreenQuestion, On: p.q != nil})
}

// Send sends the action k, N n (send).
func (p *Presenter) Send(k ActKind, n int) { p.send(k, n) }

// DrawMenu draws a menu in the game's font (drawMenu).
func (p *Presenter) DrawMenu(dst *gpu.Image, lines []string, keys []byte, sel int, k float64, cx, top int) {
	p.drawMenu(dst, p.glyph, "", lines, keys, nil, sel, k, cx, top)
}

// MenuRects are the rows of the menu drawn last.
func (p *Presenter) MenuRects() []image.Rectangle { return p.menuRects }

// KeyWindow pops a question up (keyWindow).
func (p *Presenter) KeyWindow(dst *gpu.Image, title string, lines []string, keys []byte, sel int) {
	p.keyWindow(dst, title, lines, keys, nil, sel, false)
}

// WinRects are the rows of the pop-up drawn last.
func (p *Presenter) WinRects() []image.Rectangle { return p.winRects }

// DrawCards draws the party's cards where the 3D view has them, exp the
// way to the next level under each, and the tip of the one pointed at.
func (p *Presenter) DrawCards(dst *gpu.Image, exp bool) []image.Rectangle {
	p.expShown = exp
	p.drawParty(dst, p.shownParty(), p.cardMarked, cardAt(dst), nil)
	p.expShown = false
	p.layCards(dst)
	p.drawCardTip(dst)
	return p.cards
}

// PopUp is whether a pop-up question is up, and its highlight.
func (p *Presenter) PopUp() (*bool, *int) { return &p.ask, &p.sel }

// Point puts the pointer at x, y over the cards; moved: it was elsewhere.
func (p *Presenter) Point(x, y int) bool {
	p.cardHover = 1 + cardUnder(p.cards, x, y)
	if image.Pt(x, y) == p.pointer {
		return false
	}
	p.pointer = image.Pt(x, y)
	return true
}

// Calls hands over the rule calls the screens made since the last call
// (call); the game runs them on its rules before the actions.
func (p *Presenter) Calls() []func(*classicGame) {
	c := p.calls
	p.calls = nil
	return c
}

// SetEngine gives the presenter the game its queries (JournalPage) ask.
func (p *Presenter) SetEngine(e *classicGame) { p.eng, p.lay = e, layouts{} }
