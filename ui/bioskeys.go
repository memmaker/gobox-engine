package ui

import "gobox/gpu"

// BIOSKey is the code INT 16h gave for window key k (the original BIOS table, which
// the hosts share): low byte ASCII for Enter, Esc,
// Backspace and Tab, else the scan code in the high byte (arrows, Home,
// End, PgUp, PgDn, Delete, F1-F11, and the keypad's 1-9 as the same
// keys; KP5 is 0x4C00). ok is false for any other key.
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
	case gpu.KeyF11:
		return 0x8500, true
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
