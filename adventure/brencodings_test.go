package adventure

import (
	"testing"

	"gobox/rules"
)

// TestBRStatusEncoding: BR's status byte both ways through the canonical SelStatusKind (setChar keeps v&127).
func TestBRStatusEncoding(t *testing.T) {
	for st, b := range BRStatus {
		k := Canonical("BrSelStatus", b, 0)
		if StatusKinds[k] != st {
			t.Errorf("byte %d: kind %s, want %s", b, StatusKinds[k], st)
		}
		if got := Canonical("BrSelStatus", b|0x80, 0); got != k { // the high bit is dropped: 129 is 1
			t.Errorf("byte %d|128: kind %d, want %d", b, got, k)
		}
	}
	if BRStatus[rules.StatusOkay] != 1 || BRStatus[rules.StatusDead] != 3 {
		t.Errorf("okay %d dead %d: the clinic tests 1 and 3", BRStatus[rules.StatusOkay], BRStatus[rules.StatusDead])
	}
	for _, g := range []int64{0, 7, 8, 9} { // StNoStatus, StPoisoned, StComatose, StNotHere: no rules status
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("byte %d decoded", g)
				}
			}()
			Canonical("BrSelStatus", g, 0)
		}()
	}
}

func TestBRCompanionEncoding(t *testing.T) {
	for g, want := range map[int64]int64{0: 0, 1: 1, 64: 1, 128: 1, 255: 1} {
		if got := Canonical("BrSelCompanion", g, 0); got != want {
			t.Errorf("write %d: %d, want %d", g, got, want)
		}
	}
}

// TestScriptSideWithCompanion: BR's side word 0, 0x80, 0x81 with the companion bit apart.
func TestScriptSideWithCompanion(t *testing.T) {
	for _, w := range ScriptSide {
		if w&ScriptCompanion != 0 {
			t.Errorf("side word %#x holds the companion bit", w)
		}
	}
}
