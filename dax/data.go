package dax

import (
	"bytes"
	"encoding/binary"
	"os"
	"strconv"

	"gobox/asset"
)

// The Gold Box data import shared by every game: each area set's GEO maps as
// blocks (asset.WriteBlocks) and its MON and ITEM records as JSON in the
// game's own record types (C character, A affect, I item: fixed-size structs
// in the DOS layout, read with encoding/binary).

// Monster is a monster's records: MON<set>CHA, SPC (affects) and ITM (items)
// blocks of one id.
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
// GEO<n>.bin/.txt, MON<n>.json (map id: Monster) and ITEM<n>.json (map id:
// items), for every set n whose archive exists.
func ImportData[C, A, I any](l *Loader, dst string) error {
	for s := range sets {
		n := strconv.Itoa(s)
		if l.Exists("GEO" + n + ".DAX") {
			if err := asset.WriteBlocks(dst, "GEO"+n, "GEO"+n+".bin: GEO blocks, one after another (gobox/dax ImportData)", Blocks(l, "GEO"+n)); err != nil {
				return err
			}
		}
		if l.Exists("MON" + n + "CHA.DAX") {
			if err := asset.WriteJSON(dst, "MON"+n, Monsters[C, A, I](l, "MON"+n)); err != nil {
				return err
			}
		}
		if l.Exists("ITEM" + n + ".DAX") {
			if err := asset.WriteJSON(dst, "ITEM"+n, ItemBlocks[I](l, "ITEM"+n)); err != nil {
				return err
			}
		}
	}
	return nil
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

// Monsters decodes the monsters of family ("MON<n>"): one per CHA block.
func Monsters[C, A, I any](l *Loader, family string) map[int]Monster[C, A, I] {
	out := map[int]Monster[C, A, I]{}
	for id, b := range Blocks(l, family+"CHA") {
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

// Table decodes a counted table file (ITEMS: a 2-byte count, then records)
// as n T records; records past the file's end are zero.
func Table[T any](path string, n int) ([]T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) < 2 {
		b = make([]byte, 2)
	}
	out := make([]T, n)
	b = append(b[2:], make([]byte, n*binary.Size(new(T)))...)
	_, err = binary.Decode(b, binary.LittleEndian, out)
	return out, err
}
