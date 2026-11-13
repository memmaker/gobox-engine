package ui

import "gobox/gpu"

// BIOSKey is the code INT 16h gave for window key k (the original BIOS table, which
// the hosts share): low byte ASCII for Enter, Esc,
// Backspace and Tab, else the scan code in the high byte (arrows, Home,
// End, PgUp, PgDn, Delete, F1-F10, and the keypad's 1-9 as the same
// keys; KP5 is 0x4C00). ok is false for any other key.
// ShiftF12 is the BIOS code for Shift-F12, which the window sends itself
// (BIOSKey maps unshifted keys only).
const ShiftF12 = 0x8800

// ArtF12 is the BIOS code of F12, which the window itself takes: the other art.
const ArtF12 = 0x8600

// ScriptTimed is a script's \t<ms> wait (a synthetic key, not a BIOS one): the low
// 12 bits are frames (60 Hz), taken at once and counted whether or not the
// game is idle, so a blow can be shot mid-air.
const ScriptTimed Key = 0xE000

// F12Key is the code of window key k when it is F12 (shift: Shift-F12, the wall style).
func F12Key(k gpu.Key, shift bool) (code uint16, ok bool) {
	switch {
	case k != gpu.KeyF12:
		return 0, false
	case shift:
		return ShiftF12, true
	}
	return ArtF12, true
}

func BIOSKey(k gpu.Key) (code uint16, ok bool) {
	switch k {
	case gpu.KeyEnter, gpu.KeyKPEnter:
		return 13, true
	case gpu.KeyEscape:
		return 27, true
	case gpu.KeyBackspace:
		return 8, true
	case gpu.KeyTab:
		return 9, true
	case gpu.KeyF1, gpu.KeyF2, gpu.KeyF3, gpu.KeyF4, gpu.KeyF5, gpu.KeyF6, gpu.KeyF7, gpu.KeyF8:
		return 0x3B00 + uint16(k-gpu.KeyF1)<<8, true
	case gpu.KeyF9:
		return 0x4300, true
	case gpu.KeyF10:
		return 0x4400, true
	case gpu.KeyArrowUp, gpu.KeyKP8:
		return 0x4800, true
	case gpu.KeyArrowDown, gpu.KeyKP2:
		return 0x5000, true
	case gpu.KeyArrowLeft, gpu.KeyKP4:
		return 0x4B00, true
	case gpu.KeyArrowRight, gpu.KeyKP6:
		return 0x4D00, true
	case gpu.KeyHome, gpu.KeyKP7:
		return 0x4700, true
	case gpu.KeyEnd, gpu.KeyKP1:
		return 0x4F00, true
	case gpu.KeyPageUp, gpu.KeyKP9:
		return 0x4900, true
	case gpu.KeyPageDown, gpu.KeyKP3:
		return 0x5100, true
	case gpu.KeyKP5:
		return 0x4C00, true
	case gpu.KeyDelete:
		return 0x5300, true
	}
	return 0, false
}
