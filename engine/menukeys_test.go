package engine

import (
	"gobox/ecl"
	"strings"
	"testing"
)

// TestMenuKeys: pod's command 49 says "what do you do?", then asks the HMenu.
func TestMenuKeys(t *testing.T) {
	e := quickStarted(t)
	out := testField(t, e, "Scratch1")
	args := []ecl.Arg{{Type: 1, Ref: out}, {N: 2}, {Type: 0x80, S: "FIGHT"}, {Type: 0x80, S: "FLEE"}}
	e.vm.Script = &ecl.Script{Code: []ecl.Instr{{Args: args}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	e.opMenuKeys()
	if !strings.Contains(e.paneNote(), menuKeysPrompt) {
		t.Fatalf("pane %q, want the prompt", e.paneNote())
	}
	if e.eclWait == nil {
		t.Fatal("no menu asked")
	}
}
