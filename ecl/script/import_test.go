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
