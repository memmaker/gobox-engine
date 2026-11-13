package pool

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// TestWritesFailWithTheirError: every save write that fails (a read-only
// folder) is the error its function returns, not only a "Cannot write" prompt.
func TestWritesFailWithTheirError(t *testing.T) {
	e := &Engine{set: 1}
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	e.data, e.adv = "../data/por/sources", "../data/por/sources/assets"
	if err := os.Chmod(e.Cfg.SaveDir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(e.Cfg.SaveDir, 0o700)
	for name, err := range map[string]error{
		"WriteCharList": e.WriteCharList([]string{"A"}),
		"saveSlot":      e.saveSlot(),
		"writeLabels":   e.writeLabels(e.savePath(labelsFile)),
		"setRosterNext": e.setRosterNext(5),
	} {
		if !errors.Is(err, fs.ErrPermission) {
			t.Errorf("%s on a read-only folder: %v, want the permission error", name, err)
		}
	}
	if _, err := e.newRosterID(); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("newRosterID on a read-only folder: %v", err)
	}
}
