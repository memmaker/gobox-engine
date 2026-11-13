package engine

import (
	"errors"
	"testing"

	"gobox/ecl"
)

// TestFindItemUnknownType: a script item number missing from ids.txt is a
// typed error, and FIND_ITEM answers "not found" as the original's type
// compare does, never a crash.
func TestFindItemUnknownType(t *testing.T) {
	const unknown = 0x30 // no ids.txt entry
	e := quickStarted(t)
	var u UnknownItemType
	if _, err := e.itemID(unknown); !errors.As(err, &u) || u.N != unknown {
		t.Fatalf("itemID(%#x): %v, want UnknownItemType", unknown, err)
	}
	e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Op: ecl.OpFindItem, Args: []ecl.Arg{{N: unknown}}}, {Op: ecl.OpExit}}}
	e.vm.PC, e.vm.Stop = 0, false
	e.eclRun()
	if e.vm.Flags[0] || !e.vm.Flags[1] {
		t.Fatalf("flags %v, want not found", e.vm.Flags)
	}
}
