package games

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/engine"
	"gobox/importer"
	"gobox/rules1e"
)

// TestOriginals imports the originals in $COTAB_DATA (skipped when unset)
// and checks the folder loads (pool.ReadGameTables, rules1e.ReadLayout,
// rules.json) and every character file under the originals, and the game
// data's items, go through rules1e's Import and Export byte-equal.
func TestCotabOriginals(t *testing.T) {
	dos := os.Getenv("COTAB_DATA")
	if dos == "" {
		t.Skip("COTAB_DATA (the DOS originals) unset")
	}
	root := t.TempDir()
	if err := importer.Run("cotab", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "cotab")
	if _, err := engine.ReadGameTables(dst); err != nil {
		t.Fatal(err)
	}
	l, err := rules1e.ReadLayout(filepath.Join(dst, "data", rules1e.LayoutFile))
	if err != nil {
		t.Fatal(err)
	}
	tb, err := rules1e.POR.LoadRules(filepath.Join(dst, "data", "rules.json"), l)
	if err != nil {
		t.Fatal(err)
	}
	var items []rules1e.ItemBase // the base item table as pool reads it
	if ok, err := asset.ReadJSON(filepath.Join(dst, "data"), orig.ItemTableFile, &items); !ok || err != nil || len(items) != engine.ItemTypes {
		t.Fatalf("%s.json: %d items, %v %v", orig.ItemTableFile, len(items), ok, err)
	}
	ld, err := orig.NewLoader(dos)
	if err != nil {
		t.Fatal(err)
	}
	var data []rules1e.ItemFile
	for set := range 10 {
		for id := range 256 {
			data = append(data, orig.Records[rules1e.ItemFile](ld.BlockShared(fmt.Sprintf("MON%dITM", set), id))...)
		}
		for _, its := range orig.ItemBlocks[rules1e.ItemFile](ld, fmt.Sprintf("ITEM%d", set)) {
			data = append(data, its...)
		}
	}
	b, _ := binary.Append(nil, binary.LittleEndian, data)
	checked, errs := rules1e.CheckLaterDir[rules1e.Record422](tb, items, dos, rules1e.ItemSize, rules1e.ReadItems(b),
		rules1e.LaterSave{Record: ".GUY", Items: ".SWG", Affects: ".FX"}, rules1e.LaterSave{Record: ".SAV", Items: ".SWG", Affects: ".FX"})
	t.Logf("%d character files, %d data items", len(checked), len(data))
	if len(checked) == 0 {
		t.Error("no character files under", dos)
	}
	for _, err := range errs {
		t.Error(err)
	}
}
