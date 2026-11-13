// frontgame.go: the presenter is the screens' RuleHost; its rule calls run on
// the game, a *classicGame.

package present

import (
	"fmt"
	"gobox/ega"
	"gobox/front"
	"gobox/gpu"
	"image"
	"slices"
)

var _ RuleHost = (*Presenter)(nil)

// Call queues f (call).
func (p *Presenter) Call(f func(*classicGame)) { p.call(func(e *classicGame) { f(e) }) }

// Edit queues f (edit).
func (p *Presenter) Edit(f func(*classicGame)) { p.edit(func(e *classicGame) { f(e) }) }

// ownCreation puts the creation sheet up in v, over the party menu.
func (p *Presenter) ownCreation(v front.View) front.View {
	if p.creating && p.eng != nil {
		v.Creation = p.eng.CreationView()
		if p.creating = v.Creation != nil; !p.creating { // the game ended it itself (a step after Keep, a save): the party menu is back
			return v
		}
		v = p.eng.SheetIcon(v) // its icon boxes take the draft's combat icon
		v.MainMenu, v.Logo = nil, nil
	}
	return v
}

// Ask puts q up (nil: none).
func (p *Presenter) Ask(q *front.Question) { p.q = q }

// Asking is the question up (nil: none).
func (p *Presenter) Asking() *front.Question { return p.q }

// Shown is what the last frame showed, as the screens see it.
func (p *Presenter) Shown() Shown {
	v := p.shownView
	var t *front.Loot
	if in := p.inventory(); in != nil {
		t = in.treasure
	}
	return Shown{Treasure: t, Ask: v.Ask != nil, Powers: v.Powers != nil, Sheet: v.Sheet != nil, Inventory: p.inventoryUp(), Battle: p.board != nil,
		Camp: v.Camp != nil, MainMenu: p.mainMenuUp(), Loot: v.Loot, Door: v.Door, Resting: p.pacing().Resting, Party: len(v.Party)}
}

// ownQuestion is v with the presenter's question up.
func (p *Presenter) ownQuestion(v front.View) front.View {
	q := p.q
	if q == nil && v.Ask == nil {
		l := v.Loot
		if p.inventoryUp() { // the loot screen's item list is the treasure: its Take, All, Exit, not the gate menu over it
			l = nil
		}
		q = GateQuestion(p, l, v.Door)
	}
	if q != nil {
		a := q.Ask
		v.Ask, v.Exploring, v.CanCast, v.Interact = &a, false, false, false // the question takes the keys
		if a.Kind == front.AskMember {
			v.Party = slices.Clone(v.Party)
			for i := range v.Party {
				v.Party[i].Selected = i == q.Cand
			}
		}
	}
	return v
}

// questionAct takes a while the presenter's question is up, at the gate
// or while camp rests (AskAct); then the powers and the rule calls of the
// screens up; false leaves a to the others.
func (p *Presenter) questionAct(a front.Action) bool {
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

// ownPicker puts the picker and the icon editor up in v, over the screens
// under them (the creation sheet, camp, the party menu).
func (p *Presenter) ownPicker(v front.View) front.View {
	if p.eng == nil || p.gridPick == nil && p.icon == nil {
		return v
	}
	key := [2]any{p.ran, p.ui().grid}
	v.Creation, v.SheetPage, v.Camp, v.MainMenu, v.Sheet, v.Inventory = nil, nil, nil, nil, nil, nil
	v.Portrait, v.IconBox, v.Logo = front.CellRect{}, front.CellRect{}, nil
	if p.icon != nil {
		if ed := p.icon.Layout(p.eng, key); ed != nil {
			v.IconEditor, v.IconDrop = ed, true
		}
	}
	if p.gridPick != nil {
		v.Picker, v.IconEditor, v.IconDrop = p.gridPick.Layout(p.eng, key), nil, false
		if v.Picker != nil { // pictures named by DAX ref, as View.Pictures
			pk := *v.Picker
			pk.Pics = daxPictures(indexedPictures(pk.Pics, v.Ink))
			v.Picker = &pk
		}
	}
	return v
}

// openSlots puts the slot screen up (F10, the menu's Save/Load), or takes
// it down when it is up; not over another page of the whole window.
func (p *Presenter) openSlots(v front.View, loadOnly bool) {
	switch {
	case p.slots != nil:
		p.slots = nil
	case p.eng != nil && !p.modal() && p.eng.SlotsView(loadOnly) != nil:
		p.slots = &SlotsUI{LoadOnly: loadOnly}
	}
}

// OpenSlots is F10 from the renderer's options (Save/Load).
func (p *Presenter) OpenSlots() { p.openSlots(p.shownView, false) }

// own is v with the presenter's own screens the engine's View carries:
// the slots.
func (p *Presenter) own(v front.View) front.View {
	if p.eng != nil && p.eng.Resets() != p.resets { // a load or Back to Main Menu closed what they were over
		p.resets = p.eng.Resets()
		p.gridPick, p.icon, p.creating, p.training, p.q, p.clue = nil, nil, false, false, nil, nil
	}
	v = p.blinked(p.paged(pixView(v)))
	if v.Intro != nil && len(p.intro.Flash) > 0 { // the title's lightning (ovr02:0626 title_screen): palette entry 1 in the flash step's EGA colour
		v.Ink[1] = ega.RGBA(p.intro.Flash[0])
	}
	v.Pictures = daxPictures(indexedPictures(v.Pictures, v.Ink))
	v = p.ownQuestion(p.ownPicker(p.ownCreation(p.ownTraining(v))))
	if p.slots != nil && p.eng != nil {
		v.MainMenu, v.Logo = nil, nil // the page covers the party menu: its keys are the page's
	}
	return v
}

// slotAct takes a decision a on the slot screen: false leaves it to the
// engine.
func (p *Presenter) slotAct(a front.Action) bool {
	sl := p.slots
	if a.Kind == front.ActCreate || a.Kind == front.ActModify { // only a new character is asked Roll or Point Buy
		p.scr.newChar = a.Kind == front.ActCreate
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
	return front.Screens(
		front.ScreenUp{Name: front.ScreenSlots, On: p.slots != nil},
		front.ScreenUp{Name: front.ScreenTitle, On: p.intro.Phase != IntroNone, State: fmt.Sprint(p.intro.Phase, len(p.intro.Flash))},
		front.ScreenUp{Name: front.ScreenPicker, On: p.gridPick != nil},
		front.ScreenUp{Name: front.ScreenIconEditor, On: p.icon != nil},
		front.ScreenUp{Name: front.ScreenCreation, On: p.creating},
		front.ScreenUp{Name: front.ScreenQuestion, On: p.q != nil})
}

// Send sends the action k, N n (send).
func (p *Presenter) Send(k front.ActKind, n int) { p.send(k, n) }

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
	v := p.shownView
	p.expShown = exp
	p.drawParty(dst, partyOf(p.state()), p.cardSelected(), cardAt(dst))
	p.expShown = false
	p.cards = p.cards[:0]
	for i := range v.Party {
		p.cards = append(p.cards, cardAt(dst)(i, len(v.Party)))
	}
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
func (p *Presenter) Calls() []func(*classicGame) { return p.calls.Take() }

// SetEngine gives the presenter the game its queries (JournalPage) ask.
func (p *Presenter) SetEngine(e *classicGame) { p.eng = e }
