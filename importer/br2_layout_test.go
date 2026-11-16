package importer

import (
	"testing"

	"gobox/ecl"
)

// Buck Rogers 2's full script layout (br2/scripts/layout/global.txt), proven against the byte code of every ECL block of
// the originals, with BR1's helpers (br1_layout_test.go).

// br2TextBytes are the aux words the scripts use as text and also clear through a byte operand (`V_7A00 = ""`, `V_7B04 = 0`
// in areas 19 and 20: the same word reused as a counter); 7C00 is the name, which area 36's Text(V_7C00) names through a byte operand.
var br2TextBytes = map[int]bool{0x7A00: true, 0x7B04: true, 0x7C00: true}

func TestBR2LayoutShape(t *testing.T) {
	lay := readBRLayout(t, "br2")
	for name, addr := range map[string]int{"DivideRemainder": 0x7F3F, "LoopCounter": 0x4CF6, "LoopLimit": 0x4CF7, "EntryBytes": 0x4C00, "NewGameBytes": 0x7F79} {
		r, err := lay.Named(name)
		if err != nil || lay.Address(r) != addr {
			t.Errorf("%s: %v at %04X, want %04X", name, err, lay.Address(r), addr)
		}
	}
	for _, f := range lay.Fields {
		if f.Width == ecl.WidthBytes && (f.Owner != ecl.OwnerLocal || (f.Name != "EntryBytes" && f.Name != "NewGameBytes")) {
			t.Errorf("byte buffer %s: want exactly the local EntryBytes and NewGameBytes", f.Name)
		}
	}
}

func TestBR2LayoutResolvesEveryOperand(t *testing.T) {
	checkEveryOperand(t, "br2", operandLeniency{
		wideFields: map[string]bool{"DivideRemainder": true, "SelHpMax": true, "SelHpCur": true},
		textBytes:  br2TextBytes,
	})
}

func TestBR2LayoutImportRoundTrip(t *testing.T) { checkImportRoundTrip(t, "br2") }
