package ui

import (
	"testing"
)

func UITestMenuNav(t *testing.T) {
	sel := 0
	keys := []byte{'C', 'E'}
	if i, back, _ := MenuNav(Esc, 27, &sel, keys); i != -1 || !back {
		t.Fatal("Esc picked a row")
	}
	if i, _, _ := MenuNav(None, 'e', &sel, keys); i != 1 || sel != 1 {
		t.Fatal("hotkey e did not pick row 1")
	}
	if MenuNav(Down, 0x5000, &sel, keys); sel != 0 {
		t.Fatal("Down did not wrap")
	}
	if s, at := HotText("Camp", 'E'); s != "[E] Camp" || at != 1 {
		t.Fatal(s)
	}
	if s, _ := HotText("Cure Blindness  1000g", '1'); s != "[1] Cure Blindness  1000g" {
		t.Fatal(s)
	}
}

func UITestDigitKeys(t *testing.T) {
	keys := DigitKeys([]byte{'B', 0, '2', 0, 0}, func(i int) bool { return i != 3 })
	if string(keys) != "B12\x003" {
		t.Fatalf("%q", keys)
	}
	many := DigitKeys(make([]byte, 12), func(int) bool { return true })
	if many[8] != '9' || many[9] != 0 {
		t.Fatalf("%q", many)
	}
}

func TestTitleLines(t *testing.T) {
	got := TitleLines("a b\n\nc")
	if len(got) != 3 || got[0] != "a b" || got[1] != "" {
		t.Fatalf("%q", got)
	}
}
