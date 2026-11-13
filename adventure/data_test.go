package adventure

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestBlocksRoundTrip: blocks and JSON read back as written; a missing set
// reads as none, a length past the end fails.
func TestBlocksRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := map[int][]byte{3: {1, 2, 3}, 0: {9}}
	if err := WriteBlocks(dir, "GEO1", "test", in); err != nil {
		t.Fatal(err)
	}
	out, err := ReadBlocks(dir, "GEO1")
	if err != nil || len(out) != 2 || !bytes.Equal(out[3], in[3]) || !bytes.Equal(out[0], in[0]) {
		t.Fatalf("read %v, %v", out, err)
	}
	if out, err := ReadBlocks(dir, "GEO2"); out != nil || err != nil {
		t.Fatalf("missing: %v, %v", out, err)
	}
	os.WriteFile(filepath.Join(dir, "GEO1.txt"), []byte("block 1 0 9\n"), 0o644)
	if _, err := ReadBlocks(dir, "GEO1"); err == nil {
		t.Fatal("past the end read")
	}
	if err := WriteJSON(dir, "x", map[int]string{1: "a"}); err != nil {
		t.Fatal(err)
	}
	var m map[int]string
	if ok, err := ReadJSON(dir, "x", &m); !ok || err != nil || m[1] != "a" {
		t.Fatalf("json %v %v %v", m, ok, err)
	}
	if ok, err := ReadJSON(dir, "y", &m); ok || err != nil {
		t.Fatal("missing json read")
	}
}
