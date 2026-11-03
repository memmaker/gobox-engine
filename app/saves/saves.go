// Package saves is the shell's save slots on disk: one folder per slot
// under a root, listed newest first, written beside the old one and
// swapped in only once complete. The game writes its own files into the
// folder (Write); saves never reads them.
package saves

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
}

// Path is the slot's folder.
func (s Store) Path(name string) string { return filepath.Join(s.Dir, name) }

// List is the slots, newest first.
func (s Store) List() []Slot {
	ents, _ := os.ReadDir(s.Dir)
	var out []Slot
	for _, en := range ents {
		st, err := os.Stat(filepath.Join(s.Dir, en.Name(), s.Marker))
		if en.IsDir() && err == nil && !strings.HasPrefix(en.Name(), ".") {
			out = append(out, Slot{en.Name(), st.ModTime()})
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
	}
}

// ErrIncomplete is Write's error when write left no Marker.
var ErrIncomplete = errors.New("saves: the save wrote no marker")

// Write saves slot name: write fills a fresh folder (dir) beside the old
// slot, which is replaced only once the new one holds Marker.
func (s Store) Write(name string, write func(dir string)) error {
	tmp := s.Path("." + name + ".saving")
	os.RemoveAll(tmp)
	write(tmp)
	if _, err := os.Stat(filepath.Join(tmp, s.Marker)); err != nil {
		return ErrIncomplete
	}
	os.RemoveAll(s.Path(name))
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
