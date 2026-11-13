package orig

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"sync"

	"gobox/asset"
)

// The Gold Box data import shared by every game: each area set's GEO maps as
// blocks (asset.WriteBlocks) and its MON and ITEM records as JSON in the
// game's own record types (C character, A affect, I item: fixed-size structs
// in the DOS layout, read with encoding/binary).

// Monster is a monster's records: MON<set>CHA, SPC (affects) and ITM (items)
// blocks of one id. MON<n>.json holds it with the record as the rules' sheet
// (ImportData's sheet), so every game's monsters read the same.
type Monster[C, A, I any] struct {
	Record  C
	Affects []A `json:",omitempty"`
	Items   []I `json:",omitempty"`
}

// blockIDs are the ids a DAX archive can hold.
const blockIDs = 256

// sets are the area-set digits the archives are named by (GEO1, MON0CHA, ITEM3 ...).
const sets = 10

// ImportData writes the maps and records of l's archives into dst:
// GEO<n>.bin/.txt, WALLDEF<n>.bin/.txt, MON<n>.json (map id: Monster) and ITEM<n>.json (map id:
// items), for every set n whose archive exists. sheet is a game record as the rules' sheet and
// item a game item as the rules' item, the one form MON<n>.json and ITEM<n>.json hold them in.
func ImportData[C, A, I, S, T any](l *Loader, dst string, sheet func(C) S, item func(I) T) error {
	for s := range sets {
		n := strconv.Itoa(s)
		if l.Exists("GEO" + n + ".DAX") {
			if err := asset.WriteBlocks(dst, "GEO"+n, "GEO"+n+".bin: GEO blocks, one after another (gobox/dax ImportData)", Blocks(l, "GEO"+n)); err != nil {
				return err
			}
		}
		if l.Exists("WALLDEF" + n + ".DAX") {
			if err := asset.WriteBlocks(dst, "WALLDEF"+n, "WALLDEF"+n+".bin: WALLDEF blocks, one after another (gobox/dax ImportData)", Blocks(l, "WALLDEF"+n)); err != nil {
				return err
			}
		}
		if l.Exists("MON" + n + "CHA.DAX") {
			out := map[int]Monster[S, A, T]{}
			for id, m := range Monsters[C, A, I](l, "MON"+n) {
				out[id] = Monster[S, A, T]{Record: sheet(m.Record), Affects: m.Affects, Items: items(m.Items, item)}
			}
			if err := asset.WriteJSON(dst, "MON"+n, out); err != nil {
				return err
			}
		}
		if l.Exists("ITEM" + n + ".DAX") {
			out := map[int][]T{}
			for id, its := range ItemBlocks[I](l, "ITEM"+n) {
				out[id] = items(its, item)
			}
			if err := asset.WriteJSON(dst, "ITEM"+n, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// items is each of its as item makes it.
func items[I, T any](its []I, item func(I) T) []T {
	out := make([]T, len(its))
	for i, it := range its {
		out[i] = item(it)
	}
	return out
}

// Blocks is every block of archive name, by id.
func Blocks(l *Loader, name string) map[int][]byte {
	out := map[int][]byte{}
	for id := range blockIDs {
		if b := l.BlockShared(name, id); b != nil {
			out[id] = b
		}
	}
	return out
}

// Monsters decodes the monsters of family ("MON<n>"): one per non-empty CHA
// block. An empty block is no monster (the originals' "Unable to load
// monster"), so it has no entry.
func Monsters[C, A, I any](l *Loader, family string) map[int]Monster[C, A, I] {
	out := map[int]Monster[C, A, I]{}
	for id, b := range Blocks(l, family+"CHA") {
		if len(b) == 0 {
			continue
		}
		var m Monster[C, A, I]
		rec := make([]byte, binary.Size(&m.Record))
		copy(rec, b) // a short block reads as zero-padded
		binary.Read(bytes.NewReader(rec), binary.LittleEndian, &m.Record)
		m.Affects = Records[A](l.BlockShared(family+"SPC", id))
		m.Items = Records[I](l.BlockShared(family+"ITM", id))
		out[id] = m
	}
	return out
}

// ItemBlocks decodes archive name's item lists by id (an empty block: none).
func ItemBlocks[I any](l *Loader, name string) map[int][]I {
	out := map[int][]I{}
	for id, b := range Blocks(l, name) {
		out[id] = append([]I{}, Records[I](b)...)
	}
	return out
}

// Records decodes b as T records one after another (a partial last one is dropped).
func Records[T any](b []byte) []T {
	var out []T
	n := binary.Size(new(T))
	for ; n > 0 && len(b) >= n; b = b[n:] {
		var v T
		binary.Read(bytes.NewReader(b[:n]), binary.LittleEndian, &v)
		out = append(out, v)
	}
	return out
}

// TableHead is how a table file starts: POR, COTAB and SotSB count their
// records in 2 bytes, POD's file is the records alone (O079:0837).
type TableHead int

const (
	Counted TableHead = iota
	Bare
)

// Table decodes a table file (ITEMS) as n T records; records past the
// file's end are zero.
func Table[T any](path string, n int, head TableHead) ([]T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if head == Counted {
		b = b[min(2, len(b)):]
	}
	out := make([]T, n)
	b = append(b, make([]byte, n*binary.Size(new(T)))...)
	_, err = binary.Decode(b, binary.LittleEndian, out)
	return out, err
}

// ItemTableFile is the base item table in an import's data dir (JSON).
const ItemTableFile = "ITEMS"

// ImportItemTable writes the originals' ITEMS (n T records, head as the
// game's file starts) to dst as ItemTableFile; no ITEMS: nothing.
func ImportItemTable[T any](l *Loader, dst string, n int, head TableHead) error {
	if !l.Exists("ITEMS") {
		return nil
	}
	items, err := Table[T](l.Path("ITEMS"), n, head)
	if err != nil {
		return err
	}
	return asset.WriteJSON(dst, ItemTableFile, items)
}

// Data is an adventure's imported maps and records (ImportData's files in
// its data dir), each file read on first use. A file that is not there is
// none; one that does not read is a broken import: it panics.
type Data[C, A, I any] struct {
	dir      string
	mu       sync.Mutex
	blocks   map[string]map[int][]byte
	monsters map[string]map[int]Monster[C, A, I]
	items    map[string]map[int][]I
}

// OpenData is the imported data in dir.
func OpenData[C, A, I any](dir string) *Data[C, A, I] {
	return &Data[C, A, I]{dir: dir, blocks: map[string]map[int][]byte{}, monsters: map[string]map[int]Monster[C, A, I]{}, items: map[string]map[int][]I{}}
}

// Block is block id of a block file ("GEO2", "WALLDEF3"); nil: none. The
// slice is shared: callers must not modify it.
func (d *Data[C, A, I]) Block(file string, id int) []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	m, ok := d.blocks[file]
	if !ok {
		var err error
		if m, err = asset.ReadBlocks(d.dir, file); err != nil {
			panic(fmt.Sprintf("data %s: %v (run the import again)", file, err))
		}
		d.blocks[file] = m
	}
	return m[id]
}

// Monster is monster id of family ("MON2"); false: none.
func (d *Data[C, A, I]) Monster(family string, id int) (Monster[C, A, I], bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	m, ok := d.monsters[family]
	if !ok {
		m = readJSON[map[int]Monster[C, A, I]](d.dir, family)
		d.monsters[family] = m
	}
	r, ok := m[id]
	return r, ok
}

// Items is item list id of file ("ITEM2"); nil: none.
func (d *Data[C, A, I]) Items(file string, id int) []I {
	d.mu.Lock()
	defer d.mu.Unlock()
	m, ok := d.items[file]
	if !ok {
		m = readJSON[map[int][]I](d.dir, file)
		d.items[file] = m
	}
	return m[id]
}

// readJSON is dir/name.json (zero: none); a broken one panics.
func readJSON[T any](dir, name string) T {
	var v T
	if _, err := asset.ReadJSON(dir, name, &v); err != nil {
		panic(fmt.Sprintf("data %s: %v (run the import again)", name, err))
	}
	return v
}

// ItemTableBytes is dir's ItemTableFile as n T records in the DOS layout
// (the bytes a game's engine reads); nil: the import has none.
func ItemTableBytes[T any](dir string, n int) ([]byte, error) {
	var items []T
	if ok, err := asset.ReadJSON(dir, ItemTableFile, &items); !ok || err != nil {
		return nil, err
	}
	items = append(items, make([]T, max(0, n-len(items)))...)
	return binary.Append(nil, binary.LittleEndian, items[:n])
}
