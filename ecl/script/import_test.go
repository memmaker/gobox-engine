package script

import (
	"errors"
	"testing"
)

func TestApplyPatch(t *testing.T) {
	got, err := ApplyPatch("a\nb\nc\nd\n", "--- a\n+++ b\n@@ -2,2 +2,2 @@\n b\n-c\n+C\n")
	if err != nil || got != "a\nb\nC\nd\n" {
		t.Errorf("%q %v", got, err)
	}
	if _, err := ApplyPatch("a\nx\n", "@@ -1,1 +1,1 @@\n-b\n+c\n"); err == nil {
		t.Error("a patch that does not match must fail")
	}
}

func TestHeader(t *testing.T) {
	if _, err := V1.checkHeader(V1.Header() + "x"); err != nil {
		t.Error(err)
	}
	for _, text := range []string{"x\n", "version 1 0\n", "version 99 1\n"} {
		if _, err := V1.checkHeader(text); !errors.Is(err, ErrStale) {
			t.Errorf("%q: %v", text, err)
		}
	}
}

// A vendor that falls through into the No branch keeps it: only the gate goes.
func TestFoldKeepsNoBranch(t *testing.T) {
	label := func(n int) Arg { return Arg{Type: ArgVar, N: n, Label: true} }
	row := func(s string) Arg { return Arg{Type: ArgText, S: s} }
	s := &Script{Labels: map[string]int{"no": 5}, At: map[int]string{5: "no"}, Code: []Instr{
		{Op: OpHMenu, Args: []Arg{{Type: ArgVar, N: 1}, {N: 2}, row("YES"), row("NO")}},
		{Op: OpOnGoto, Args: []Arg{{Type: ArgVar, N: 1}, {N: 2}, label(2), label(5)}},
		{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 7}}},
		{Op: OpClearMonsters},
		{Op: OpCombat},
		{Op: OpPrint, Args: []Arg{row("BYE")}},
		{Op: OpExit},
	}}
	names := &Names{Version: V1, Vars: map[int]string{7: varTemple}}
	if f := FoldVendorGates(s, 0, names); len(f) != 1 {
		t.Fatalf("folds %+v", f)
	}
	found := false
	for _, in := range s.Code {
		found = found || in.Op == OpPrint
	}
	if !found {
		t.Error("the No branch was cut")
	}
}
