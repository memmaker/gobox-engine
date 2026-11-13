package ui

import (
	"image"
	"path/filepath"
	"slices"
	"strings"

	"gobox/gpu"
	"gobox/ui/input"
)

// memberKey is the member key k picks where a member applies: F1-F8 that
// member, the pad's stick buttons the previous and next of n after cur
// (none selected: the first), and with lr Left and Right too (where the
// screen has no other use for them). ok: k is a member key.
func memberKey(k uint16, cur, n int, lr bool) (int, bool) {
	switch a := actOf(k); {
	case a >= input.Member1 && a < input.Member1+input.Members:
		i := int(a - input.Member1)
		return i, i < n
	case n == 0:
	case a == input.PrevMember || lr && a == input.InputLeft:
		if cur <= 0 {
			return n - 1, true
		}
		return cur - 1, true
	case a == input.NextMember || lr && a == input.InputRight:
		return (cur + 1) % n, true
	}
	return -1, false
}

// globalKey: k works on every screen (keyAct): help, the minimap, save and
// load, the sound, the font. A screen passes it on rather than swallowing it.
func globalKey(k uint16) bool {
	switch actOf(k) {
	case input.Help, input.Minimap, input.SaveLoad, input.Sound, input.Font, input.FontPrev, input.WallStyle:
		return true
	}
	return false
}

// frontOnly: k opens one of gobox's own screens (the slots, the map, the
// overview), which the game never asked for.
func frontOnly(k uint16) bool {
	switch actOf(k) {
	case input.SaveLoad, input.InputMap, input.Overview, input.Party:
		return true
	}
	return false
}

// arrow turns a cursor key a into its decision on v: the member pick's candidate (ActLead), else a pager's page
// (ActPage -1 or +1).
func (p *Presenter) arrow(a input.Act) {
	back := a == input.Up || a == input.InputLeft || a == input.PgUp || a == input.Home
	on := a == input.Down || a == input.InputRight || a == input.PgDn || a == input.End
	if n := len(p.party()); n > 0 && p.choosing() && (back || on) {
		i := p.selectedIndex()
		switch {
		case on:
			i = (i + 1) % n // none: the first
		case i <= 0:
			i = n - 1
		default:
			i--
		}
		p.send(ActLead, i)
		return
	}
	switch {
	case p.picker(p.ink()) != nil: // a picker pages to an absolute page (pickerKey), never by -1 and +1
	case back:
		p.send(ActPage, -1)
	case on:
		p.send(ActPage, 1)
	}
}

func (p *Presenter) send(kind ActKind, n int) {
	if kind == ActChoose || kind == ActOK { // a question answered: the next menu pops up, the same one again too
		p.asked, p.pageAsked = "", ""
	}
	if kind == ActLoad && p.slots == nil { // the party menu's Load
		p.openSlots(true)
		return
	}
	if kind == ActSaveGame && p.slots == nil { // the party menu's Save
		p.openSlots(false)
		return
	}
	if p.slotAct(Action{Kind: kind, N: n}) {
		return
	}
	p.inv.acts = append(p.inv.acts, Action{Kind: kind, N: n})
}

// keyAct turns key k into the action it stands for on the screen v shows.
func (p *Presenter) keyAct(k uint16) {
	if t := p.typingAsk(); t != nil { // the line being typed is edited here
		if !p.typing {
			p.typed, p.typedCur, p.typing = t.Text, len(t.Text), true
		}
		switch done, cancel, _ := editKey(k, &p.typed, &p.typedCur, t.Max, t.Kind == AskNumber); {
		case done: // the whole line, once: the game never sees a line half typed
			p.inv.acts = append(p.inv.acts, Action{Kind: ActText, Text: p.typed})
			p.typing = false
		case cancel:
			p.send(ActClose, 0)
			p.typing = false
		}
		return
	}
	p.typing = false
	if p.askAnswered(k) {
		return
	}
	if p.askList() != nil && p.askMenu() != nil && !p.choosing() && p.listKey(k) {
		return
	}
	if p.screenUp() && p.campKey(k) {
		return
	}
	if p.exploring() && p.exploreKey(k) {
		return
	}
	if n, ok := memberKey(k, p.selectedIndex(), len(p.party()), false); ok { // F1-F8, the stick buttons: select the member (exploring: and open their context menu, as a card's click)
		if p.command {
			p.cardClick(n)
			return
		}
		p.send(ActLead, n)
		return
	}
	c := byte(k)
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	switch a := actOf(k); {
	case a == input.Enter || a == input.PadA:
		p.send(ActOK, 0)
	case a == input.Esc:
		p.send(ActClose, 0)
	case a == input.Minimap:
		Host.MapOff = !Host.MapOff // pure UI: the presenter's own option, never the game's
		if OnSettings != nil {
			OnSettings(Settings)
		}
	case a == input.Overview && p.exploring():
		p.overview = true
	case a == input.Overview:
		p.send(ActOverview, 0)
	case a == input.Font:
		if !p.cycleFont(1) {
			p.send(ActFont, 1)
		}
	case a == input.FontPrev:
		if !p.cycleFont(-1) {
			p.send(ActFont, -1)
		}
	case a == input.Sound:
		p.send(ActSound, 0)
	case a == input.SaveLoad:
		p.openSlots(false)
	case a == input.WallStyle:
		Host.PixelWalls = !Host.PixelWalls // newView puts it in use
		if OnSettings != nil {             // saved with the host's settings
			OnSettings(Settings)
		}
	case input.Cursor(a):
		p.arrow(a)
	case a == input.Help:
		p.openHelp()
	case k >= 128 || c < ' ': // another key the table binds (or none): nothing here
	case menuIndex(p.askMenu(), c) >= 0:
		p.send(ActChoose, menuIndex(p.askMenu(), c))
	case a == input.InputMap:
		p.openMap()
	case a == input.Party && p.exploring():
		p.overview = true
	case a == input.Party:
		p.send(ActOverview, 0)
	case keypad(k) != input.None:
		p.arrow(keypad(k))
	case c == ' ' || p.askMenu() == nil && !p.mainMenuUp(): // "press any key" (the party menu ignores keys it has no row for)
		p.send(ActOK, 0)
	}
}

// Drop takes a file dropped on the window: a PNG while the sheet is up
// becomes the member's portrait, on the icon editor (or dropped on the
// creation sheet's icon, cell cx, cy) the combat icon.
func (p *Presenter) Drop(path string, cx, cy int, in bool) {
	if !strings.EqualFold(filepath.Ext(path), ".png") {
		return
	}
	r := p.iconBox()
	switch {
	case p.icon != nil:
		IconAct(p, &p.icon, Action{Kind: ActIcon, Text: path})
	case p.iconEditor(p.ink()) != nil || in && r != (CellRect{}) && image.Pt(cx, cy).In(image.Rect(r.X1, r.Y1, r.X2+1, r.Y2+1)):
		p.inv.acts = append(p.inv.acts, Action{Kind: ActIcon, Text: path})
	case p.portraitBox() != (CellRect{}) || p.inventoryKind() == invOwn: // the sheet or the member's inventory
		p.inv.acts = append(p.inv.acts, Action{Kind: ActPortrait, N: -1, Text: path})
	}
}

// menuIndex is the item of m that hotkey c picks, or -1.
func menuIndex(m *Menu, c byte) int {
	if m != nil {
		for i, it := range m.Commands {
			if it.Key == c {
				return i
			}
		}
	}
	return -1
}

// Mouse takes the pointer on text cell (cx, cy) of the page (in: on the
// page at all), a left or right click and wheel steps (up positive): the
// clicks the screens take, Esc on the right button, the highlight on the
// wheel; a click elsewhere answers "press a key".
func (p *Presenter) Mouse(cx, cy int, in, left, right bool, wheel int) {
	p.sync()
	if p.introUp() != nil { // its clicks are Click's
		return
	}
	if p.journal.open { // its rows on the window's pixels (Hover, Click)
		p.journalMouse(p.camp.pointer, false, right, wheel)
		return
	}
	if p.overview { // its clicks are Click's
		return
	}
	if p.fullMap != nil {
		p.mapMouse(cx, cy, in, left, right)
		return
	}
	if p.clue != nil { // its clicks are Click's
		p.clueMouse(p.camp.pointer, false, right, wheel)
		return
	}
	if pg, _ := p.pages(); pg != nil { // the pager: the wheel and the ◀ ▶ marks page, the right button closes, another click pages on
		w, h := p.pageSize()
		switch {
		case wheel != 0:
			p.pagerTurn(-unit(wheel))
		case right:
			p.closePager()
		case left && in:
			d := 1
			if x, ok := pagerMark(pg, w/8, -1); ok && cx == x && cy == h/8-2 {
				d = -1
			}
			p.pagerTurn(d)
		}
		return
	}
	for ; wheel != 0; wheel -= unit(wheel) {
		if p.fight && p.fightLog { // wheel up: up, a line a step, as every list
			p.logTop = max(0, p.logTop-unit(wheel))
			continue
		}
		if p.fight {
			p.battle.Zoom(wheel > 0, Host.Zoom)
			continue
		}
		if p.screenUp() || p.askList() != nil || p.pageMenu { // over a list: a line at most a pause, the rest of a fling dropped
			if c := &p.camp; now().Sub(c.wheeled) < 2*wheelPause {
				break
			}
			p.camp.wheeled = now() // a list or menu (camp-style, the list popup, the page's: save slots, drop-downs, the sheet): the wheel moves the highlight, a line a step, up for up
			k := keyOf(input.Down)
			if wheel > 0 {
				k = keyOf(input.Up)
			}
			p.key(k)
		}
	}
	if p.pageScreen() && (p.pageMenu || p.picker(p.ink()) != nil) { // the page screens' own pointer (screens.go): the right button is Esc
		if right {
			p.screenKey(keyOf(input.Esc))
			right = false
		}
		if p.picker(p.ink()) != nil || p.creation() != nil || p.iconEditor(p.ink()) != nil {
			if in {
				p.screenMouse(cx, cy, left)
			}
			return
		}
	}
	switch { // the right button: on a card that member's sheet, else it closes the context menu, else Esc
	case !right:
	case p.confirm() != "": // a yes/no question: No
		p.confirmPick(1)
	case p.cardHover > 0 && (p.active || p.main):
		p.ctx = 0
		p.send(ActMember, p.cardHover-1)
	case p.fight && p.pointer.In(p.fightCard): // the active member's sheet
		p.send(ActMember, 0)
	case p.fight: // as Esc: cancel, else the options
		p.fightKey(keyOf(input.Esc))
	case p.ctx > 0:
		p.ctx = 0
	case p.askMenu() != nil && len(p.askMenu().Commands) == 1 && p.typingAsk() == nil: // "PRESS <ENTER> TO CONTINUE" takes no Esc: the right button continues too
		p.menu = false
		p.send(ActOK, 0)
	case p.menu && p.exploring(): // the command bar's popup: it alone closes (key)
		p.menu = false
	case p.screenUp(): // back one level, as Esc
		p.campBack()
	case p.exploring() && !p.menu && p.active && p.view != nil: // as Esc: the options
		p.view.OpenOptions()
	default:
		p.send(ActClose, 0)
	}
	if r := p.portraitBox(); left && in && gpu.IsKeyPressed(gpu.KeyShift) && image.Pt(cx, cy).In(image.Rect(r.X1, r.Y1, r.X2+1, r.Y2+1)) { // Shift-click on the sheet's portrait: the next head
		p.send(ActHead, -1)
		return
	}
	switch {
	case !left || !in:
	case p.typingAsk() == nil && (p.askMenu() == nil || len(p.askMenu().Commands) == 1): // no bar, or one with nothing to choose (the scripts' "PRESS <ENTER> TO CONTINUE")
		p.send(ActOK, 0)
	}
}

// unit is -1 for a negative n, else 1.
func unit(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

// askMenu is the question's prompt and options when v asks for an option,
// a line of a list or a member (nil otherwise).
func (p *Presenter) askMenu() *Menu {
	if a := p.askUp(); a != nil && a.Kind == AskModify { // the options, then the presenter's Raise and Lower (modifyItems) on the line highlighted
		return &Menu{Prompt: a.Prompt, Commands: append(slices.Clone(a.Commands), modifyItems...)}
	}
	if a := p.askUp(); a != nil && (a.Kind == AskMenu || a.Kind == AskList || a.Kind == AskMember) {
		return &a.Menu
	}
	return nil
}

// modifyItems are the options AskModify adds to the game's: they raise
// and lower the line highlighted (ActRaise X +1, -1), as Right and Left,
// + and - and a click on the line (+1) do.
var modifyItems = []Command{{Text: "Raise", Key: '+'}, {Text: "Lower", Key: '-'}}

// modifying: v asks for a stat raised or lowered (AskModify).
func (p *Presenter) modifying() bool { return p.askUp() != nil && p.askUp().Kind == AskModify }

// askDefault is the option v's question highlights first (Ask.Default).
func (p *Presenter) askDefault() int {
	if p.askUp() != nil {
		return p.askUp().Default
	}
	return 0
}

// askList is the list v asks a line of (nil: none).
func (p *Presenter) askList() []FrontEntry {
	if a := p.askUp(); a != nil && (a.Kind == AskList || a.Kind == AskModify) {
		return a.Entries
	}
	return nil
}

// editKey is the one line editor over *text with its cursor *cur: a
// printable character (a digit only, for a number) is raw input and goes
// in at the cursor, up to max; the other keys are the table's: Enter is
// done, Esc cancels, Left/Right/Home/End move the cursor, Backspace takes
// the character before it and Delete the one under it. ok: k was taken.
func editKey(k uint16, text *string, cur *int, max int, digits bool) (done, cancel, ok bool) {
	t, c := *text, min(len(*text), *cur)
	if k >= 0x20 && k < 0x7F { // a character
		if digits && (k < '0' || k > '9') || len(t) >= max {
			return false, false, false
		}
		*text, *cur = t[:c]+string(rune(k))+t[c:], c+1
		return false, false, true
	}
	switch a := actOf(k); {
	case a == input.Enter:
		return true, false, true
	case a == input.Esc:
		return false, true, true
	case a == input.InputLeft:
		c = c - min(c, 1)
	case a == input.InputRight:
		c = min(len(t), c+1)
	case a == input.Home:
		c = 0
	case a == input.End:
		c = len(t)
	case a == input.Backspace && c > 0:
		t, c = t[:c-1]+t[c:], c-1
	case a == input.Delete && c < len(t):
		t = t[:c] + t[c+1:]
	default:
		return false, false, false
	}
	*text, *cur = t, c
	return false, false, true
}

// caretLine is text with its cursor at cur, in a field w+1 cells wide: a
// '_' over the character at cur while on (the blink), scrolled to keep
// the cursor in the field, padded with blanks.
func caretLine(text string, cur, w int, on bool) string {
	cur = min(cur, len(text))
	if on {
		if cur < len(text) {
			text = text[:cur] + "_" + text[cur+1:]
		} else {
			text += "_"
		}
	}
	off := max(0, cur-w)
	text = text[min(off, len(text)):min(len(text), off+w+1)]
	return text + strings.Repeat(" ", w+1-len(text))
}

// typingAsk is the line the question asks to be typed (nil: none).
func (p *Presenter) typingAsk() *FrontAsk {
	if a := p.askUp(); a != nil && (a.Kind == AskText || a.Kind == AskNumber) {
		return a
	}
	return nil
}

// choosing: v asks for a party member.
func (p *Presenter) choosing() bool { return p.askUp() != nil && p.askUp().Kind == AskMember }

// spellPick is the spells v asks to pick from (nil: none).
func (p *Presenter) spellPick() []Spell {
	if a := p.askUp(); a != nil && a.Kind == AskSpell {
		out := make([]Spell, len(a.Entries)) // never nil: an empty list still takes Esc
		for i, e := range a.Entries {
			out[i] = Spell{ID: e.Spell, Name: e.Text}
		}
		return out
	}
	return nil
}

// confirm is v's yes/no question (AskConfirm), "" for none.
func (p *Presenter) confirm() string {
	if a := p.askUp(); a != nil && a.Kind == AskConfirm && p.sfOf() == nil { // the ship fight asks on its bar
		return a.Prompt
	}
	return ""
}

// confirmStart is the row (0 Yes, 1 No) a's yes/no starts on: No, unless
// the game starts on Yes (Yes, or Default naming its Yes option, whatever
// the options' order).
func confirmStart(a *FrontAsk) int {
	if a != nil && a.NoPick {
		return -1 // no row lit
	}
	if a != nil && (a.Yes || a.Default >= 0 && a.Default < len(a.Commands) && strings.EqualFold(a.Commands[a.Default].Text, "Yes")) {
		return 0
	}
	return 1
}

// askHeld: k does nothing on v's question, a choice as a yes/no: Esc
// under NoEsc, Enter and the arrows under NoPick (only its hotkeys answer).
func (p *Presenter) askHeld(k uint16) bool {
	a := p.askUp()
	if a == nil {
		return false
	}
	switch actOf(k) {
	case input.Esc:
		return a.NoEsc
	case input.Enter, input.Up, input.Down, input.InputLeft, input.InputRight, input.Space, input.PadA:
		return a.NoPick
	}
	return false
}

// yesNoKeys are every yes/no question's hotkeys.
var yesNoKeys = []byte{'Y', 'N'}

// confirmKey takes k on the engine's yes/no question as on every Yes/No:
// the arrows, Enter, Space, pad A, Y and N; Esc (pad B) is No. Only the
// global keys go on.
func (p *Presenter) confirmKey(k uint16) bool {
	if a := p.askUp(); a != nil && a.NoPick { // no highlight: only Y and N answer (Esc as the game says)
		switch {
		case k == 'y' || k == 'Y':
			p.confirmPick(0)
		case k == 'n' || k == 'N':
			p.confirmPick(1)
		case actOf(k) == input.Esc && !a.NoEsc:
			p.confirmPick(1)
		case !slices.Contains([]input.Act{input.Enter, input.Up, input.Down, input.Esc, input.InputLeft, input.InputRight}, actOf(k)):
			return !globalKey(k)
		}
		return true
	}
	switch i, back, ok := MenuNav(actOf(k), k, &p.confirmSel, yesNoKeys); {
	case back && p.askUp() != nil && p.askUp().NoEsc: // the game keeps it up on Esc
	case back:
		p.confirmPick(1)
	case i >= 0:
		p.confirmPick(i)
	case !ok:
		return !globalKey(k)
	}
	return true
}

// confirmPick answers the yes/no question: row 0 is Yes, anything else No.
func (p *Presenter) confirmPick(i int) {
	if i == 0 {
		p.send(ActYes, 0)
		return
	}
	p.send(ActNo, 0)
}

// keypadDirs are the direction keys as the keypad's digits (ActDir N).
var keypadDirs = map[input.Act]int{input.Up: 8, input.Down: 2, input.InputLeft: 4, input.InputRight: 6, input.Home: 7, input.PgUp: 9, input.End: 1, input.PgDn: 3}

// askAnswered sends what key k answers on v's question by its data: a
// direction (AskMove: ActDir), another member (Ask.Reselect,
// F1..F8: ActLead); false: k is none of these.
func (p *Presenter) askAnswered(k uint16) bool {
	a := p.askUp()
	if a == nil {
		return false
	}
	if d, ok := keypadDirs[actOf(k)]; ok && a.Kind == AskMove {
		p.send(ActDir, d)
		return true
	}
	if n, ok := memberKey(k, p.selectedIndex(), len(p.party()), false); a.Reselect && ok { // F1..F8, the pad's stick buttons
		p.send(ActLead, n)
		return true
	}
	if a.View && (k == 'v' || k == 'V') {
		p.send(ActView, -1)
		return true
	}
	return false
}
