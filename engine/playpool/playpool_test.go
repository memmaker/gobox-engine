package playpool

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHomeSaves: the first run makes the saves' home, seeded from the folder's
// saves; a second run keeps it (ported from pod's playpool).
func TestHomeSaves(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	adv := t.TempDir()
	os.MkdirAll(filepath.Join(adv, "saves"), 0o755)
	os.WriteFile(filepath.Join(adv, "saves", "A.PC"), []byte("x"), 0o644)
	d, err := HomeSaves("pod", adv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(d, "A.PC")); err != nil {
		t.Fatal(err)
	}
	if again, err := HomeSaves("pod", adv); err != nil || again != d {
		t.Fatal("second call moved", again, err)
	}
}
