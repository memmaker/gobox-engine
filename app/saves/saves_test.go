package saves

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	mark := func(dir string) error { os.MkdirAll(dir, 0o755); return os.WriteFile(filepath.Join(dir, "M"), nil, 0o644) }
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
