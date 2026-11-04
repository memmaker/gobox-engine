// Package saves is the shell's save slots on disk: one folder per slot
// under a root, listed newest first, written beside the old one and
// swapped in only once complete. The game writes its own files into the
// folder (Write); saves never reads them.
package saves

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// Store is the slots under Dir. A folder is a slot once it holds Marker
// (the file the game's save always writes); a folder starting with "." is
// a save still being written.
type Store struct {
	Dir    string
	Marker string
}

// Slot is one saved game.
type Slot struct {
	Name string
	At   time.Time // when Marker was written
	Meta Meta      // the slot's MetaFile (zero: none written)
}

// Meta is what a slot list shows of a save beyond its name and time: a
// game writes it into the slot with WriteMeta.
type Meta struct {
	Summary string // "Phlan, 4 members"
	Levels  []int  // the members' levels
}

// MetaFile is the file WriteMeta writes in a slot.
const MetaFile = "SLOT.META"

// WriteMeta writes m into the slot folder dir (inside Write's write).
func WriteMeta(dir string, m Meta) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, MetaFile), b, 0o644)
}

// WriteSide writes v, a fixed-size value (a uint64 purse, a struct of
// them), little-endian into file in the slot folder dir (inside Write's
// write).
func WriteSide[T any](dir, file string, v T) error {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, v); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), b.Bytes(), 0o644)
}

// ReadSide reads WriteSide's file of slot name into a T.
func ReadSide[T any](s Store, name, file string) (T, error) {
	var v T
	b, err := os.ReadFile(filepath.Join(s.Path(name), file))
	if err == nil {
		err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &v)
	}
	return v, err
}

// Path is the slot's folder.
func (s Store) Path(name string) string { return filepath.Join(s.Dir, name) }

// List is the slots, newest first.
//
// It is read again only when the folder changed (its time: a slot written,
// renamed in or removed): rules list the slots in every State, many a frame.
func (s Store) List() []Slot {
	st, err := os.Stat(s.Dir)
	key := s.Dir + "\x00" + s.Marker
	listed.Lock()
	defer listed.Unlock()
	if c, ok := listed.m[key]; ok && err == nil && st.ModTime().Equal(c.at) {
		return slices.Clone(c.slots)
	}
	out := s.list()
	if err == nil {
		if listed.m == nil {
			listed.m = map[string]listing{}
		}
		listed.m[key] = listing{st.ModTime(), out}
	}
	return slices.Clone(out)
}

// listed is List's last read of each folder.
var listed struct {
	sync.Mutex
	m map[string]listing
}

type listing struct {
	at    time.Time
	slots []Slot
}

// forget drops List's read of s, after s wrote or removed a slot (the
// folder's time may not have moved within its resolution).
func (s Store) forget() {
	listed.Lock()
	delete(listed.m, s.Dir+"\x00"+s.Marker)
	listed.Unlock()
}

func (s Store) list() []Slot {
	ents, _ := os.ReadDir(s.Dir)
	var out []Slot
	for _, en := range ents {
		st, err := os.Stat(filepath.Join(s.Dir, en.Name(), s.Marker))
		if en.IsDir() && err == nil && !strings.HasPrefix(en.Name(), ".") {
			sl := Slot{Name: en.Name(), At: st.ModTime()}
			if b, err := os.ReadFile(filepath.Join(s.Dir, en.Name(), MetaFile)); err == nil {
				json.Unmarshal(b, &sl.Meta)
			}
			out = append(out, sl)
		}
	}
	slices.SortFunc(out, func(a, b Slot) int { return b.At.Compare(a.At) })
	return out
}

// Index is the slot's place in List, or -1.
func (s Store) Index(name string) int {
	return slices.IndexFunc(s.List(), func(sl Slot) bool { return sl.Name == name })
}

// Delete deletes the slot, if there is one.
func (s Store) Delete(name string) {
	if s.Index(name) >= 0 {
		os.RemoveAll(s.Path(name))
		s.forget()
	}
}

// ErrIncomplete is Write's error when write left no Marker.
var ErrIncomplete = errors.New("saves: the save wrote no marker")

// Write saves slot name: write fills a fresh folder (dir) beside the old
// slot, which is replaced only once the new one holds Marker. write's
// error comes back, the old slot kept.
func (s Store) Write(name string, write func(dir string) error) error {
	tmp := s.Path("." + name + ".saving")
	os.RemoveAll(tmp)
	if err := write(tmp); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	if _, err := os.Stat(filepath.Join(tmp, s.Marker)); err != nil {
		return ErrIncomplete
	}
	os.RemoveAll(s.Path(name))
	defer s.forget()
	return os.Rename(tmp, s.Path(name))
}

// Name keeps what a slot's folder name may hold: letters, digits and " -_!'".
func Name(s string) string {
	s = strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || strings.ContainsRune(" -_!'", r) {
			return r
		}
		return -1
	}, s)
	return strings.TrimSpace(s)
}
