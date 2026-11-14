package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/rules1e"
)

// TestPodHalberd: PoD's imported ITEMS has the halberd a two-handed weapon,
// not the original's flail in the gauntlets' slot.
func TestPodHalberd(t *testing.T) {
	dos := filepath.Join("..", "data", "pod", "sources", "dos")
	if _, err := os.Stat(filepath.Join(dos, "ITEMS")); err != nil {
		t.Skip("pod: no originals")
	}
	dst := t.TempDir()
	if err := PodImportDataTo(dos, dst); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dst, adventure.ItemTableFile+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var items []rules1e.ItemBase
	if err := json.Unmarshal(b, &items); err != nil {
		t.Fatal(err)
	}
	if got := items[podHalberd]; got != podHalberdBase {
		t.Errorf("halberd %+v, want %+v", got, podHalberdBase)
	}
}
