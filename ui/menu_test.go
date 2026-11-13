package ui

import (
	"testing"
)

func TestTitleLines(t *testing.T) {
	got := TitleLines("a b\n\nc")
	if len(got) != 3 || got[0] != "a b" || got[1] != "" {
		t.Fatalf("%q", got)
	}
}
