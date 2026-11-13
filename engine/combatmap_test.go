package engine

import (
	"gobox/rules"
	"testing"
)

func TestLine(t *testing.T) {
	l := rules.NewLine(0, 0, 6, -3)
	var dirs []byte
	for l.Step() {
		dirs = append(dirs, l.Dir)
	}
	if l.X != 6 || l.Y != -3 || len(dirs) != 6 || l.Dir != 8 || dirs[0] != 1 || dirs[1] != 2 {
		t.Fatalf("line ended at %d,%d dirs %v last %d", l.X, l.Y, dirs, l.Dir)
	}
}
