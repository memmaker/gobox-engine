package adventure

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// TestReadLook: no file is the zero Look; a written one reads back typed.
func TestReadLook(t *testing.T) {
	adv := t.TempDir()
	if l, err := ReadLook(adv); err != nil || l.Door != nil || l.ReadyMark != "" {
		t.Fatalf("absent file: %+v, %v", l, err)
	}
	if err := os.MkdirAll(filepath.Join(adv, DataDir), 0o755); err != nil {
		t.Fatal(err)
	}
	const js = `{"ReadyMark":"Yes ","Door":{"R":1,"G":2,"B":3,"A":255}}`
	if err := os.WriteFile(filepath.Join(adv, DataDir, LookFile), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := ReadLook(adv)
	if err != nil || l.ReadyMark != "Yes " || l.Door == nil || *l.Door != (color.RGBA{1, 2, 3, 255}) || l.AmigaGround != nil {
		t.Fatalf("read: %+v, %v", l, err)
	}
}
