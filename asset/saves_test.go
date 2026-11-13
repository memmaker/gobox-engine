package asset

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	mark := func(dir string) error {
		os.MkdirAll(dir, 0o755)
		return os.WriteFile(filepath.Join(dir, "M"), nil, 0o644)
	}
	if err := s.Write("a", mark); err != nil {
		t.Fatal(err)
	}
	os.Chtimes(filepath.Join(s.Path("a"), "M"), time.Now(), time.Now().Add(-time.Hour))
	s.Write("b", mark)
	if err := s.Write("c", func(string) error { return nil }); err != ErrIncomplete {
		t.Fatal(err)
	}
	if l := s.List(); len(l) != 2 || l[0].Name != "b" || l[1].Name != "a" {
		t.Fatal(l)
	}
	s.Delete("b")
	if s.Index("b") != -1 || s.Index("a") != 0 {
		t.Fatal(s.List())
	}
	if n := Name(" Ab/c:1! "); n != "Abc1!" {
		t.Fatal(n)
	}
}

// A slot keeps its metadata and typed side data; a writer's error comes
// back and keeps the old slot.
func TestMetaSide(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	err := s.Write("a", func(dir string) error {
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, "M"), nil, 0o644)
		WriteSide(dir, "PURSE", uint64(1<<40))
		return WriteMeta(dir, Meta{Summary: "Phlan", Levels: []int{3, 4}})
	})
	if err != nil {
		t.Fatal(err)
	}
	if l := s.List(); len(l) != 1 || l[0].Meta.Summary != "Phlan" || len(l[0].Meta.Levels) != 2 {
		t.Fatal(l)
	}
	if v, err := ReadSide[uint64](s, "a", "PURSE"); err != nil || v != 1<<40 {
		t.Fatal(v, err)
	}
	bad := errors.New("full")
	if err := s.Write("a", func(string) error { return bad }); err != bad || s.Index("a") != 0 {
		t.Fatal(err)
	}
}

// List reads the folder again only when it changed: the rules list the slots in every State.
func TestListCachedUntilChange(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	if err := s.Write("a", func(d string) error { os.MkdirAll(d, 0o755); return os.WriteFile(filepath.Join(d, "M"), nil, 0o644) }); err != nil {
		t.Fatal(err)
	}
	if n := len(s.List()); n != 1 {
		t.Fatalf("%d slots, want 1", n)
	}
	s.Write("b", func(d string) error { os.MkdirAll(d, 0o755); return os.WriteFile(filepath.Join(d, "M"), nil, 0o644) })
	if n := len(s.List()); n != 2 {
		t.Fatalf("after a write: %d slots, want 2", n)
	}
	os.RemoveAll(s.Path("a")) // a game removing a slot itself
	if n := len(s.List()); n != 1 {
		t.Fatalf("after a removal: %d slots, want 1", n)
	}
}

func BenchmarkList(b *testing.B) {
	s := Store{Dir: b.TempDir(), Marker: "M"}
	for i := range 30 {
		s.Write(string(rune('a'+i)), func(d string) error { os.MkdirAll(d, 0o755); return os.WriteFile(filepath.Join(d, "M"), nil, 0o644) })
	}
	for range b.N {
		s.List()
	}
}

// A slot that cannot be removed or replaced (a read-only folder) is the
// error of Delete and Write, with the slot still there.
func TestStoreErrorsComeBack(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	mark := func(dir string) error {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "M"), nil, 0o644)
	}
	if err := s.Write("a", mark); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.Dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(s.Dir, 0o700)
	if err := s.Delete("a"); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Delete in a read-only folder: %v, want the permission error", err)
	}
	if err := s.Write("a", mark); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Write in a read-only folder: %v, want the permission error", err)
	}
	s2 := Store{Dir: t.TempDir(), Marker: "M"}
	if err := s2.Write("a", mark); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s2.Dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(s2.Dir, 0o700)
	var err error
	s2.DeleteJoin("a", &err)
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("DeleteJoin: %v, want the permission error", err)
	}
}
