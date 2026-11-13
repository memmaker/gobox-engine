package ui

import (
	"slices"

	"gobox/gpu"
)

// The one table of the game's bindings: what each key, mouse gesture and
// gamepad button does, by context. The help screen (HelpLines) is made
// from it and the host's gamepad (PadKeys) reads it, so neither can drift
// from the other, and the screens' keyboard dispatch looks the keys up in
// it (actOf, combatAct) and switches on the Act, so none of the three can
// drift. Every action has a keyboard, a mouse and a gamepad way; where one
// is missing, Gap says why. Letters and digits that pick a menu's or the
// bar's row are that row's data (its hotkey), and typed text is raw input.

// The window's keys as the game sends them, BIOS style: low byte ASCII,
// high byte scan code. The table below is the only place that names them.
const (
	KeyF1          = 0x3B00   // F1; F2..F8 follow at +0x100
	KeyF9          = 0x4300   // camp: the party's menu
	KeyF10         = 0x4400   // save / load
	KeyShiftF12    = ShiftF12 // the walls: hi-def, pixel art
	KeyFontPrev    = 0xFD00   // Ctrl-Shift-F (Ctrl-F is ASCII 6)
	KeyPadSelect   = 0xE000   // gamepad: choose the highlighted (as Enter)
	KeyPadMinimap  = 0xE100   // as Tab
	KeyPadOverview = 0xE200   // as O
	KeyPadPrev     = 0xE300   // gamepad: the previous party member (the left stick's button)
	KeyPadNext     = 0xE400   // ... the next (the right stick's)
	KeyPadStart    = 0xE500   // gamepad Start in combat: the options, even while aiming (where Esc cancels)

	keyUp, keyDown, keyLeft, keyRight = 0x4800, 0x5000, 0x4B00, 0x4D00
	keyPgUp, keyPgDn, keyHome, keyEnd = 0x4900, 0x5100, 0x4700, 0x4F00
	keyDel                            = 0x5300
	keyEnter, keyEsc, keyTab, keyBksp = 13, 27, 9, 8
	keyCtrlS, keyCtrlF, keyCtrlN      = 0x13, 6, 14
)

// memberCodes are F1-F8 and the member each picks.
func memberCodes() []Code {
	c := make([]Code, Members)
	for i := range c {
		c[i] = Code{K: KeyF1 + uint16(i)<<8, A: Member1 + Act(i)}
	}
	return c
}

// MemberDigits is a row a game opts into (Data.Bindings): the digits 1-6
// pick party member 1-6, as F1-F6 do (the original's keys in some games);
// the digits then no longer pick a menu's lines. Its mouse and pad ways
// are F1-F8's: a click on the card, the stick buttons.
var MemberDigits = Binding{Ctx: ctxMenus, Keys: "1-6", Desc: "member 1-6", Mouse: "click a card", Pad: "L3 R3", Codes: func() []Code {
	c := make([]Code, 6)
	for i := range c {
		c[i] = Code{K: '1' + uint16(i), A: Member1 + Act(i)}
	}
	return c
}()}

// helpKeyW and helpDescW are the help's columns: the keys (or the mouse's
// or pad's way), then what they do.
const helpKeyW, helpDescW = 12, 23

// The contexts, in the help's order.
const (
	ctxMenus   = "Menus"
	ctxAdv     = "Adventuring"
	ctxCombat  = "Combat"
	ctxAnytime = "Any time"
)

const (
	dUp, dDown, dLeft, dRight  = gpu.StandardGamepadButtonLeftTop, gpu.StandardGamepadButtonLeftBottom, gpu.StandardGamepadButtonLeftLeft, gpu.StandardGamepadButtonLeftRight
	padA, padB, padX, padY     = gpu.StandardGamepadButtonRightBottom, gpu.StandardGamepadButtonRightRight, gpu.StandardGamepadButtonRightLeft, gpu.StandardGamepadButtonRightTop
	padLB, padRB, padLT, padRT = gpu.StandardGamepadButtonFrontTopLeft, gpu.StandardGamepadButtonFrontTopRight, gpu.StandardGamepadButtonFrontBottomLeft, gpu.StandardGamepadButtonFrontBottomRight
	padBack, padStart          = gpu.StandardGamepadButtonCenterLeft, gpu.StandardGamepadButtonCenterRight
	padL3, padR3               = gpu.StandardGamepadButtonLeftStick, gpu.StandardGamepadButtonRightStick
)

var bindings = []Binding{
	{Ctx: ctxMenus, Keys: "Up Down", Desc: "move the highlight", Mouse: "point", Pad: "D-pad, stick",
		Codes:   []Code{{K: keyUp, A: Up}, {K: keyDown, A: Down}},
		PadKeys: []PadKey{{B: dUp, Key: keyUp, Repeat: true}, {B: dDown, Key: keyDown, Repeat: true}}},
	{Ctx: ctxMenus, Keys: "PgUp PgDn", Desc: "ten lines up, down", Mouse: "wheel", Pad: "LB RB",
		Codes:   []Code{{K: keyPgUp, A: PgUp}, {K: keyPgDn, A: PgDn}},
		PadKeys: []PadKey{{B: padLB, Key: keyPgUp, Repeat: true}, {B: padRB, Key: keyPgDn, Repeat: true}}},
	{Ctx: ctxMenus, Keys: "Home End", Desc: "first, last line", Mouse: "scroll bar", Pad: "LT RT",
		Codes:   []Code{{K: keyHome, A: Home}, {K: keyEnd, A: End}},
		PadKeys: []PadKey{{B: padLT, Key: keyHome, Repeat: true}, {B: padRT, Key: keyEnd, Repeat: true}}},
	{Ctx: ctxMenus, Keys: "Enter Space", Desc: "pick the highlighted", Mouse: "click", Pad: "A, Start",
		Codes:   []Code{{K: keyEnter, A: Enter}, {K: ' ', A: Space}, {K: KeyPadSelect, A: PadA}},
		PadKeys: []PadKey{{B: padA, Key: KeyPadSelect, Repeat: false}, {B: padStart, Key: keyEnter, Repeat: false}}},
	{Ctx: ctxMenus, Keys: "Letter 1-9", Desc: "pick that line", Mouse: "click", Pad: "A"},
	{Ctx: ctxMenus, Keys: "Y N", Desc: "answer a question", Mouse: "click", Pad: "A"},
	{Ctx: ctxMenus, Keys: "Esc", Desc: "back, No", Mouse: "right click", Pad: "B",
		Codes:   []Code{{K: keyEsc, A: Esc}},
		PadKeys: []PadKey{{B: padB, Key: keyEsc, Repeat: false}}},
	{Ctx: ctxMenus, Keys: "F1-F8", Desc: "member 1-8", Mouse: "click a card", Pad: "L3 R3",
		Codes: memberCodes()},
	{Ctx: ctxMenus, Keys: "F9", Desc: "camp: party menu", Mouse: "point", Pad: "Left Right",
		Codes: []Code{{K: KeyF9, A: InputPartyMenu}}},
	{Ctx: ctxMenus, Keys: "Tab", Desc: "sheet, items: switch", Mouse: "View, Items", Pad: "Y"},
	{Ctx: ctxMenus, Keys: "Left Right", Desc: "prev., next member", Mouse: "click a card", Pad: "L3 R3",
		Codes:   []Code{{K: keyLeft, A: InputLeft}, {K: keyRight, A: InputRight}, {K: KeyPadPrev, A: PrevMember}, {K: KeyPadNext, A: NextMember}},
		PadKeys: []PadKey{{B: dLeft, Key: keyLeft, Repeat: true}, {B: dRight, Key: keyRight, Repeat: true}, {B: padL3, Key: KeyPadPrev, Repeat: false}, {B: padR3, Key: KeyPadNext, Repeat: false}}},
	{Ctx: ctxMenus, Keys: "+ - Bksp", Desc: "spells: one more, less", Mouse: "click (more)", Pad: "A (more)", Gap: "one less is keyboard only",
		Codes: []Code{{K: '+', A: More}, {K: '=', A: More}, {K: '-', A: Less}}}, // Backspace: the Letters row's

	{Ctx: ctxAdv, Keys: "Arrows", Desc: "walk, turn", Mouse: "map: Walk", Pad: "D-pad, stick"}, // the Menus rows' arrows; the digits are the keypad (keypadActs)
	{Ctx: ctxAdv, Keys: "Enter", Desc: "command menu", Mouse: "click middle", Pad: "A, Start"}, // the Menus row's Enter
	{Ctx: ctxAdv, Keys: "C I V", Desc: "cast items view", Mouse: "card menu", Pad: "L3 R3, pick"},
	{Ctx: ctxAdv, Keys: "A", Desc: "camp", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "W", Desc: "sweep: search walking", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "S", Desc: "search this square", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "Space", Desc: "<act>: interact", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "M", Desc: "map: walk, labels", Mouse: "minimap", Pad: "A, then pick",
		Codes: []Code{{K: 'm', A: InputMap}}},
	{Ctx: ctxAdv, Keys: "Tab", Desc: "minimap", Mouse: "command menu", Pad: "Y",
		Codes:   []Code{{K: keyTab, A: Minimap}, {K: KeyPadMinimap, A: Minimap}},
		PadKeys: []PadKey{{B: padY, Key: KeyPadMinimap, Repeat: false}}},
	{Ctx: ctxAdv, Keys: "O", Desc: "party overview", Mouse: "command menu", Pad: "X",
		Codes:   []Code{{K: 'o', A: Party}, {K: KeyPadOverview, A: Overview}},
		PadKeys: []PadKey{{B: padX, Key: KeyPadOverview, Repeat: false}}},
	{Ctx: ctxAdv, Keys: "J", Desc: "quest journal", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "B", Desc: "clue book", Mouse: "command menu", Pad: "A, then pick"},
	{Ctx: ctxAdv, Keys: "Esc", Desc: "options", Mouse: "right click", Pad: "B"},                           // the Menus row's Esc
	{Ctx: ctxAdv, Keys: "W A S D", Desc: "No Grid: walk, look", Mouse: "right click", Pad: "right stick"}, // the free walk (free.go); its right click is mouselook on/off, the command menu has Options

	{Ctx: ctxCombat, Keys: "Arrows", Desc: "step, else the cursor", Mouse: "point", Pad: "D-pad, stick",
		Codes:   []Code{{K: keyUp, A: CursorUp}, {K: keyDown, A: CursorDown}, {K: keyLeft, A: CursorLeft}, {K: keyRight, A: CursorRight}},
		PadKeys: []PadKey{{B: dUp, Key: keyUp, Repeat: true}, {B: dDown, Key: keyDown, Repeat: true}, {B: dLeft, Key: keyLeft, Repeat: true}, {B: dRight, Key: keyRight, Repeat: true}}},
	{Ctx: ctxCombat, Keys: "Enter", Desc: "go, attack, press", Mouse: "click", Pad: "A",
		Codes:   []Code{{K: keyEnter, A: Enter}, {K: ' ', A: Space}, {K: KeyPadSelect, A: PadA}},
		PadKeys: []PadKey{{B: padA, Key: KeyPadSelect, Repeat: false}}},
	{Ctx: ctxCombat, Keys: "Tab", Desc: "target; else button", Mouse: "point, hover", Pad: "LB RB",
		Codes:   []Code{{K: keyTab, A: Tab}, {K: keyPgUp, A: PgUp}, {K: keyPgDn, A: PgDn}},
		PadKeys: []PadKey{{B: padLB, Key: keyPgUp, Repeat: true}, {B: padRB, Key: keyPgDn, Repeat: true}}},
	{Ctx: ctxCombat, Keys: "Esc", Desc: "cancel; else options", Mouse: "right click", Pad: "B, Start",
		Codes:   []Code{{K: keyEsc, A: Esc}, {K: KeyPadStart, A: InputOptions}},
		PadKeys: []PadKey{{B: padB, Key: keyEsc, Repeat: false}, {B: padStart, Key: KeyPadStart, Repeat: false}}},
	{Ctx: ctxCombat, Keys: "Letter 0-9", Desc: "press a button", Mouse: "click it", Pad: "LB RB, A"},
	{Ctx: ctxCombat, Keys: "Backspace", Desc: "end the turn", Mouse: "click End", Pad: "Y",
		Codes:   []Code{{K: keyBksp, A: EndTurn}},
		PadKeys: []PadKey{{B: padY, Key: keyBksp, Repeat: false}}},
	{Ctx: ctxCombat, Keys: "G", Desc: "guard", Mouse: "click Guard", Pad: "X",
		PadKeys: []PadKey{{B: padX, Key: 'g', Repeat: false}}},
	{Ctx: ctxCombat, Keys: "Q E", Desc: "turn the camera", Pad: "LT RT", Gap: "no mouse way to turn the camera",
		Codes:   []Code{{K: 'q', A: TurnLeft}, {K: 'e', A: TurnRight}},
		PadKeys: []PadKey{{B: padLT, Key: 'q', Repeat: false}, {B: padRT, Key: 'e', Repeat: false}}},
	{Ctx: ctxCombat, Keys: "W A S D", Desc: "pan the camera", Pad: "right stick", Gap: "no mouse way to pan the camera",
		Codes: []Code{{K: 'w', A: Up}, {K: 's', A: Down}, {K: 'a', A: InputLeft}, {K: 'd', A: InputRight}}},
	{Ctx: ctxCombat, Keys: "Home", Desc: "camera on the active", Pad: "R3", Gap: "no mouse way to recentre the camera",
		Codes:   []Code{{K: keyHome, A: Home}},
		PadKeys: []PadKey{{B: padR3, Key: keyHome, Repeat: false}}},
	{Ctx: ctxCombat, Keys: "Wheel", Desc: "zoom", Mouse: "wheel", Gap: "no keyboard or pad zoom"},
	{Ctx: ctxCombat, Keys: "L", Desc: "battle log", Pad: "Back", Gap: "no mouse way to open the log",
		Codes:   []Code{{K: 'l', A: Log}},
		PadKeys: []PadKey{{B: padBack, Key: 'l', Repeat: false}}},

	{Ctx: ctxAnytime, Keys: "?", Desc: "this help", Mouse: "command menu", Pad: "A, then pick",
		Codes: []Code{{K: '?', A: Help}}},
	{Ctx: ctxAnytime, Keys: "F10", Desc: "save, load", Mouse: "Esc menu", Pad: "B, then pick",
		Codes: []Code{{K: KeyF10, A: SaveLoad}}},
	{Ctx: ctxAnytime, Keys: "Shift-F12", Desc: "walls: hi-def, pixel", Gap: "the wall style has no mouse or pad way",
		Codes: []Code{{K: KeyShiftF12, A: WallStyle}}},
	{Ctx: ctxAnytime, Keys: "L D", Desc: "saves: load, delete", Mouse: "click footer", Pad: "X Y"},
	{Ctx: ctxAnytime, Keys: "Ctrl-S", Desc: "sound on, off", Mouse: "Esc > Audio", Pad: "B, then pick",
		Codes: []Code{{K: keyCtrlS, A: Sound}}},
	{Ctx: ctxAnytime, Keys: "Ctrl-N", Desc: "text box to notes", Mouse: "> Notes", Pad: "Back",
		Codes:   []Code{{K: keyCtrlN, A: InputNote}},
		PadKeys: []PadKey{{B: padBack, Key: keyCtrlN, Repeat: false}}},
	{Ctx: ctxAnytime, Keys: "Ctrl-F", Desc: "font (+Shift: back)", Gap: "the font has no mouse or pad way",
		Codes: []Code{{K: keyCtrlF, A: Font}, {K: KeyFontPrev, A: FontPrev}}},
	{Ctx: ctxAnytime, Keys: "Letters", Desc: "type a name, number", Gap: "typing has no mouse or pad way (no on-screen keyboard)",
		Codes: []Code{{K: keyBksp, A: Backspace}, {K: keyDel, A: Delete}}},
}

// table is bindings with the contexts in the help's order.
var table = Table{Rows: bindings, Order: []string{ctxMenus, ctxAdv, ctxCombat, ctxAnytime}, Combat: ctxCombat}

// setBindings puts the game's own rows (Data.Bindings) into the table
// before the last row (Any time's Letters).
func setBindings(own []Binding) {
	rows := slices.Clone(bindings[:len(bindings)-1])
	rows = append(append(rows, own...), bindings[len(bindings)-1])
	table.Rows = rows
}

// PadKeys are the gamepad's buttons and the keys they send (see ui.Table).
func PadKeys(combat bool) []PadKey { return table.PadKeys(combat) }

// HelpLines are the help screen's lines, made from bindings.
func HelpLines() [][2]string { return table.HelpLines() }

// actOf is what key k does outside combat (the rows of every context but
// Combat); letters match either case. None: the table does not bind k (a
// hotkey or a typed character, which the screen takes as raw input).
func actOf(k uint16) Act { return table.Act(k) }

// combatAct is what key k does on the combat map (the Combat rows).
func combatAct(k uint16) Act { return table.CombatAct(k) }

// keyOf is the first key code bound to a outside combat: the mouse's and
// wheel's way into the key handlers.
func keyOf(a Act) uint16 { return table.KeyOf(a) }

// keypad is digit k as the keypad's key (None: not a digit 1-9).
func keypad(k uint16) Act { return Keypad(k) }
