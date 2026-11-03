package ui

import (
	"gobox/gpu"
	"gobox/input"
)

// A Table is a game's one list of bindings: what each key, mouse gesture
// and gamepad button does, by context. The help screen (HelpLines) is made
// from it and the host's gamepad (PadKeys) reads it, so neither can drift
// from the other, and the screens' key dispatch looks keys up in it (Act,
// CombatAct) and switches on the Act. The game owns the rows and key codes.

// PadKey is a gamepad button and the key it sends.
type PadKey struct {
	B      gpu.StandardGamepadButton
	Key    uint16
	Repeat bool // held, it repeats
}

// Code is a key code the row binds and the Act it stands for.
type Code struct {
	K uint16
	A input.Act
}

// Binding is one action in context Ctx: its keys, mouse and pad as the
// help shows them, the key codes it binds and their Acts, the buttons the
// pad sends for it, and a missing way (Gap, "" none).
type Binding struct {
	Ctx, Desc        string
	Keys, Mouse, Pad string
	Codes            []Code
	PadKeys          []PadKey
	Gap              string
}

// Table is the rows, the contexts in the help's order, and which of them
// is the combat one (its codes and pad keys apply apart from the others').
type Table struct {
	Rows   []Binding
	Order  []string
	Combat string
}

// PadKeys are the gamepad's buttons and the keys they send: in combat
// (with no menu up) the combat rows' first, then the other rows'.
func (t *Table) PadKeys(combat bool) []PadKey {
	var out []PadKey
	seen := map[gpu.StandardGamepadButton]bool{}
	add := func(in bool) {
		for _, b := range t.Rows {
			if (b.Ctx == t.Combat) != in {
				continue
			}
			for _, k := range b.PadKeys {
				if !seen[k.B] {
					seen[k.B] = true
					out = append(out, k)
				}
			}
		}
	}
	if combat {
		add(true)
	}
	add(false)
	return out
}

// HelpLines are the help screen's lines: a heading per context, then each
// action's keys and what it does, then its mouse and pad ways.
func (t *Table) HelpLines() [][2]string {
	var out [][2]string
	for _, c := range t.Order {
		out = append(out, [2]string{c, ""})
		for _, b := range t.Rows {
			if b.Ctx == c {
				out = append(out, [2]string{b.Keys, b.Desc})
			}
		}
		for _, dev := range []struct {
			name string
			way  func(Binding) string
		}{{"Mouse", func(b Binding) string { return b.Mouse }}, {"Gamepad", func(b Binding) string { return b.Pad }}} {
			out = append(out, [2]string{c + ", " + dev.name, ""})
			for _, b := range t.Rows {
				if w := dev.way(b); b.Ctx == c && w != "" {
					out = append(out, [2]string{w, b.Desc})
				}
			}
		}
	}
	return out
}

// Act is what key k does outside combat; letters match either case. None:
// the table does not bind k (a hotkey or a typed character).
func (t *Table) Act(k uint16) input.Act { return t.lookup(k, false) }

// CombatAct is what key k does on the combat map (the Combat rows).
func (t *Table) CombatAct(k uint16) input.Act { return t.lookup(k, true) }

func (t *Table) lookup(k uint16, combat bool) input.Act {
	if k >= 'A' && k <= 'Z' {
		k += 'a' - 'A'
	}
	for _, b := range t.Rows {
		if (b.Ctx == t.Combat) != combat {
			continue
		}
		for _, c := range b.Codes {
			if c.K == k {
				return c.A
			}
		}
	}
	return input.None
}

// KeyOf is the first key code bound to a outside combat: the mouse's and
// wheel's way into the key handlers.
func (t *Table) KeyOf(a input.Act) uint16 {
	for _, b := range t.Rows {
		for _, c := range b.Codes {
			if c.A == a && b.Ctx != t.Combat {
				return c.K
			}
		}
	}
	return 0
}

var keypadActs = [...]input.Act{input.End, input.Down, input.PgDn, input.Left, input.None, input.Right, input.Home, input.Up, input.PgUp}

// Keypad is digit k as the keypad's key (1 End .. 9 PgUp; None: not a
// digit 1-9).
func Keypad(k uint16) input.Act {
	if k >= '1' && k <= '9' {
		return keypadActs[k-'1']
	}
	return input.None
}
