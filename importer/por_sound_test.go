package importer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/audio"
)

// TestSoundEqualsEXE: the imported sound tables are START.EXE's driver
// tables as decoded, and the driver plays them.
func TestSoundEqualsEXE(t *testing.T) {
	data := filepath.Join("..", "data", "por", "sources")
	exe, err := os.ReadFile(filepath.Join(data, adventure.DOSDir, "START.EXE"))
	if err != nil {
		t.Skip("no START.EXE")
	}
	want, err := audio.DecodeSoundData(exe, porSoundLayout)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(data, adventure.AssetsDir, adventure.DataDir, adventure.SoundFile))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("the imported sound tables differ from START.EXE's")
	}
	if _, err := audio.NewDriver(got, 'P'); err != nil {
		t.Fatal(err)
	}
}
