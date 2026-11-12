package pool

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"rules1e"
	"testing"

	"gobox/audio"
	"gobox/audio/sounddata"
	"gobox/char"
	"gobox/core/goldbox"
	"gobox/dax"
)

// TestDataEqualsDAX: every GEO, MON and ITEM block and the ITEMS table read
// back from the imported data (gdata) as the archives decode them.
func TestDataEqualsDAX(t *testing.T) {
	data := filepath.Join("..", "..", "por-go", "data")
	l, err := dax.NewLoader(filepath.Join(data, DOSDir))
	if err != nil {
		t.Skip("no game files")
	}
	e := &Engine{adv: filepath.Join(data, AssetsDir)}
	d := e.gdata()
	n := 0
	for s := goldbox.Set(1); s <= sets; s++ {
		for id := range 256 {
			if b, got := l.BlockShared(s.File("GEO"), id), d.geoBlock(s.File("GEO"), id); !bytes.Equal(b, got) || (b == nil) != (got == nil) {
				t.Fatalf("%s %d: differs", s.File("GEO"), id)
			}
			b := l.BlockShared(s.File("MON")+"CHA", id)
			m, ok := d.monster(s.File("MON"), id)
			if (b != nil) != ok {
				t.Fatalf("%sCHA %d: in the archive %v, in the data %v", s.File("MON"), id, b != nil, ok)
			}
			if ok {
				var c rules1e.Character
				c.Sheet = char.SheetOf(b)
				var aff []rules1e.AffectRecord
				for _, a := range char.ReadRecords[rules1e.AffectFile](l.BlockShared(s.File("MON")+"SPC", id), rules1e.AffectRecSize) {
					aff = append(aff, a.Record())
				}
				var its []rules1e.Item
				for _, it := range char.ReadItems(l.BlockShared(s.File("MON")+"ITM", id)) {
					its = append(its, it)
				}
				if c.Sheet != char.SheetOf(char.Encode(m.Record)) || !reflect.DeepEqual(aff, m.Affects) || !reflect.DeepEqual(its, m.Items) {
					t.Fatalf("%s %d: differs", s.File("MON"), id)
				}
				n++
			}
			b = l.BlockShared(s.File("ITEM"), id)
			got := d.itemBlock(s.File("ITEM"), id)
			want := []rules1e.Item{}
			for _, it := range char.ReadItems(b) {
				want = append(want, it)
			}
			if b == nil {
				want = nil
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("%s %d: differs", s.File("ITEM"), id)
			}
		}
	}
	raw, err := os.ReadFile(l.Path("ITEMS"))
	if err != nil {
		t.Fatal(err)
	}
	want := make([]rules1e.ItemBase, ItemTypes)
	binary.Decode(append(raw[2:], make([]byte, ItemTypes*16)...), binary.LittleEndian, want)
	if !reflect.DeepEqual(want, d.itemTable()) {
		t.Fatal("ITEMS differs")
	}
	t.Logf("%d monsters", n)
}

// TestSoundEqualsEXE: the imported sound tables are START.EXE's driver
// tables as decoded, and the driver plays them.
func TestSoundEqualsEXE(t *testing.T) {
	data := filepath.Join("..", "..", "por-go", "data")
	exe, err := os.ReadFile(filepath.Join(data, DOSDir, "START.EXE"))
	if err != nil {
		t.Skip("no START.EXE")
	}
	want, err := sounddata.Decode(exe, SoundLayout)
	if err != nil {
		t.Fatal(err)
	}
	e := &Engine{adv: filepath.Join(data, AssetsDir)}
	got, err := e.SoundData()
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
