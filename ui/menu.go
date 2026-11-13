package ui

import (
	"bytes"
	"gobox/ui/input"
	"strings"
)

// MenuRow is one line of every menu: its text, its hotkey (0 none) and what
// picking it does (nil: nothing to pick).
type MenuRow struct {
	Text string
	Key  byte
	Do   func()
}

// RowKeys are the rows' hotkeys: a row that can be picked and has no
// letter gets a digit (DigitKeys).
func RowKeys(rows []MenuRow) []byte {
	k := make([]byte, len(rows))
	for i, r := range rows {
		k[i] = r.Key
	}
	return DigitKeys(k, func(i int) bool { return rows[i].Do != nil })
}

// RowsOff are the rows with nothing to pick (Do nil).
func RowsOff(rows []MenuRow) []bool {
	off := make([]bool, len(rows))
	for i, r := range rows {
		off[i] = r.Do == nil
	}
	return off
}

// DigitKeys gives the rows i of keys with no hotkey and on(i) the digits
// 1-9 in order (those another row has already left out), then none.
func DigitKeys(keys []byte, on func(int) bool) []byte {
	d := byte('1')
	for i := range keys {
		if keys[i] != 0 || !on(i) {
			continue
		}
		for d <= '9' && bytes.IndexByte(keys, d) >= 0 {
			d++
		}
		if d > '9' {
			break
		}
		keys[i], d = d, d+1
	}
	return keys
}

// MenuNav is the one key handler of every menu over rows with hotkeys keys.
// k is the key code and a what the game's bindings make of it (input.None
// if unbound): Up/Down (wrapping), PgUp/PgDn, Home/End move *sel; Enter,
// Space and pad A pick it, a hotkey (either case) picks its row; Esc is
// back and never picks. It returns the row picked (-1 none), back, and
// whether k was taken.
func MenuNav(a input.Act, k uint16, sel *int, keys []byte) (pick int, back, ok bool) {
	n := len(keys)
	if n == 0 {
		return -1, a == input.Esc, a == input.Esc
	}
	s := min(max(*sel, 0), n-1)
	switch {
	case a == input.Up:
		s = (s + n - 1) % n
	case a == input.Down:
		s = (s + 1) % n
	case a == input.PgUp:
		s = max(0, s-10)
	case a == input.PgDn:
		s = min(n-1, s+10)
	case a == input.Home:
		s = 0
	case a == input.End:
		s = n - 1
	case a == input.Enter || a == input.Space || a == input.PadA:
		*sel = s
		return s, false, true
	case a == input.Esc:
		return -1, true, true
	case k > ' ' && k < 128:
		h := byte(k)
		if h >= 'a' && h <= 'z' {
			h -= 'a' - 'A'
		}
		if i := bytes.IndexByte(keys, h); i >= 0 {
			*sel = i
			return i, false, true
		}
		return -1, false, false
	default:
		return -1, false, false
	}
	*sel = s
	return -1, false, true
}

// MenuNavH is MenuNav for a horizontal bar: Left is Up and Right is Down.
func MenuNavH(a input.Act, k uint16, sel *int, keys []byte) (pick int, back, ok bool) {
	switch a {
	case input.InputLeft:
		a = input.Up
	case input.InputRight:
		a = input.Down
	}
	return MenuNav(a, k, sel, keys)
}

// HotText is row s as drawn with hotkey key: the column of its white letter
// (-1 none); a key not in the text, and every digit (a row's digit in its
// text is a price, not the key), is shown in front, as "[E] Camp".
func HotText(s string, key byte) (string, int) {
	if key == 0 {
		return s, -1
	}
	if i := strings.IndexByte(strings.ToUpper(s), key); i >= 0 && (key < '0' || key > '9') {
		return s, i
	}
	return "[" + string(key) + "] " + s, 1
}

// Colour marks in a menu title or row: the text from one to the next is
// gold (InkGold) or red (InkRed).
const (
	InkGold = '\x01'
	InkRed  = '\x02'
	InkFoot = '\x03' // the rest of the title goes on the bottom border, in gold
	InkHead = '\x04' // cyan: a submenu's heading
)

// VisLen is s's length in cells, its colour marks left out.
func VisLen(s string) int {
	return len(s) - strings.Count(s, string(rune(InkGold))) - strings.Count(s, string(rune(InkRed))) - strings.Count(s, string(rune(InkHead)))
}

// WrapWords wraps s at w columns on spaces; there is always a line.
func WrapWords(s string, w int) []string {
	var out []string
	line := ""
	for _, word := range strings.Fields(s) {
		for len(word) > w {
			out = append(out, word[:w])
			word = word[w:]
		}
		switch {
		case line == "":
			line = word
		case len(line)+1+len(word) <= w:
			line += " " + word
		default:
			out = append(out, line)
			line = word
		}
	}
	return append(out, line)
}

// TitleLines are a menu title's lines: its own (\n) and blank ones kept,
// each wrapped at 24 cells, so the box stays narrow.
func TitleLines(title string) []string {
	var out []string
	for _, l := range strings.Split(title, "\n") {
		if w := WrapWords(l, 24); len(w) > 0 {
			out = append(out, w...)
		} else {
			out = append(out, "")
		}
	}
	return out
}
