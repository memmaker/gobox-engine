package dax

import (
	"os"
	"path/filepath"
	"testing"
)

// SetDOS switches BlockShared between the Amiga stand-in and the DOS block (F12).
func TestSetDOS(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "amiga"), 0o755)
	os.WriteFile(filepath.Join(dir, "PIC.DAX"), WriteDOS(map[int][]byte{1: []byte("dos")}), 0o644)
	os.WriteFile(filepath.Join(dir, "amiga", "PIC.DAX"), WriteDOS(map[int][]byte{1: []byte("amiga")}), 0o644)
	l, _ := NewLoader(dir)
	l.UseAmiga(filepath.Join(dir, "amiga"))
	for _, c := range []struct {
		dos  bool
		want string
	}{{false, "amiga"}, {true, "dos"}, {false, "amiga"}} {
		l.SetDOS(c.dos)
		if got := string(l.BlockShared("PIC", 1)); got != c.want {
			t.Errorf("SetDOS(%v): %q, want %q", c.dos, got, c.want)
		}
	}
}
