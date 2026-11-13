package games

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/importer"
)

// TestImportContents imports the originals into a scratch folder and checks
// every entry of importer.Contents and of hostFiles is there.
func TestImportContents(t *testing.T) {
	data := filepath.Join("..", "..", "data", "por", "sources")
	dos := filepath.Join(data, "dos")
	if _, err := os.Stat(filepath.Join(dos, "ITEMS")); err != nil {
		t.Skip("no originals:", err)
	}
	for _, h := range hostFiles { // untracked: absent in a clean export (the gate's)
		if _, err := os.Stat(filepath.Join(data, h.From)); err != nil {
			t.Skip("no host files:", err)
		}
	}
	root := t.TempDir()
	if err := importer.Run("por", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "por")
	if err := importer.CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	for _, h := range hostFiles {
		if _, err := os.Stat(filepath.Join(dst, h.To)); err != nil {
			t.Error(err)
		}
	}
}
