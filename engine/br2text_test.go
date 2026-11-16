package engine

import (
	"gobox/game"
	"slices"
	"strings"
	"testing"
)

// TestBR2TextNumber: BR2's scripts use a text address also as a number (the original's memory has one byte per
// address, a string its characters from the address on: eclvm.pas / buckrogers2-go ecl.go getVar, setVar read and write the byte).
// The engine's text field is that run of cells: a number through a byte operand is cell 0 (area20: Word7B04 = 0, += 1,
// 5 == Word7B04 as a drink counter; area19: Word7B04 = 0), and a number read of a text field is its first character
// (area36: Text(SelName), which prints the number: buckrogers2-go eclops.go opPrint, Typ < 128).
func TestBR2TextNumber(t *testing.T) {
	e := brBegun(t, gameBR2)
	br1OwnBlock(t, e, 201, strings.Join([]string{
		"    $Word7B04 = \"HI\"",
		"    Word7B05 = Word7B04", // 'H'
		"    Word7B04 = 0",
		"    Word7B06 = Word7B04",
		"    Word7B04 += 1",
		"    Word7B07 = Word7B04",
		"    Text(SelName)",
		"    PressKey()",
	}, "\n"))
	got := []int64{br1Cell(t, e, "Word7B05"), br1Cell(t, e, "Word7B06"), br1Cell(t, e, "Word7B07")}
	if !slices.Equal(got, []int64{'H', 0, 1}) {
		t.Errorf("number reads %v, want 'H', 0, 1", got)
	}
	if tx := paneText(commonOf(e.build())); tx != "100" { // the first roster character's name starts with 'd'
		t.Errorf("Text(SelName) printed %q, want the first character's code 100", tx)
	}
}

// TestBR2ByteBufferText: BR2's scripts keep text in the byte buffers: area82/112 InputString into NewGameBytes and
// compare it ("A76T2"), area17 puts one letter into EntryBytes_21 and prints it.
func TestBR2ByteBufferText(t *testing.T) {
	e := brBegun(t, gameBR2)
	br1OwnBlock(t, e, 201, strings.Join([]string{
		"    $EntryBytes_21 = \"B\"",
		"    Print($EntryBytes_21)",
		"    PressKey()",
		"    $NewGameBytes = InputString(6)",
		"    if $NewGameBytes == \"A76T2\" { Word7B05 = 5 }",
		"    Print($NewGameBytes)",
		"    PressKey()",
	}, "\n"))
	if got := paneText(commonOf(e.build())); got != "B" {
		t.Fatalf("EntryBytes_21 printed %q", got)
	}
	br1Answer(t, e)
	if _, err := e.doNow(game.Reply{Text: "a76t2"}); err != nil {
		c := commonOf(e.build())
		t.Fatalf("%v; %T ask %+v allowed %+v", err, e.build(), c.Ask, c.Allowed)
	}
	if got := br1Cell(t, e, "Word7B05"); got != 5 {
		t.Errorf("the typed line did not compare equal: %d", got)
	}
	if got := paneText(commonOf(e.build())); got != "BA76T2" { // Print goes on in the same line
		t.Errorf("printed %q", got)
	}
}

// TestBR2Logbook: BR2's Logbook(words, n) (area113: Logbook("HIS SPEECH", 62)) reads entry n of its Log Book
// (importer/br2/journal, buckrogers2-go data/journal/LOGBOOK.TXT).
func TestBR2Logbook(t *testing.T) {
	e := brBegun(t, gameBR2)
	br1OwnBlock(t, e, 201, "    Logbook(\"HIS SPEECH\", 62)\n    PressKey()")
	var read bool
	for range 6 {
		if sc, ok := e.build().(game.Scene); ok && sc.Kind == game.SceneHandout {
			read = len(sc.Texts) == 1 && sc.Texts[0].Title == "Log Entry 62" && sc.Handout == "Log Book"
			break
		}
		br1Answer(t, e)
	}
	if !read {
		t.Errorf("entry 62 not read: %T %q", e.build(), paneText(commonOf(e.build())))
	}
}
