package saves

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s := Store{Dir: t.TempDir(), Marker: "M"}
	mark := func(dir string) { os.MkdirAll(dir, 0o755); os.WriteFile(filepath.Join(dir, "M"), nil, 0o644) }
	if err := s.Write("a", mark); err != nil {
		t.Fatal(err)
	}
	os.Chtimes(filepath.Join(s.Path("a"), "M"), time.Now(), time.Now().Add(-time.Hour))
	s.Write("b", mark)
	if err := s.Write("c", func(string) {}); err != ErrIncomplete {
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
