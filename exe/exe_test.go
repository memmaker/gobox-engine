package exe

import (
	"encoding/binary"
	"testing"
)

// POR's START.EXE (optional; needs the original files).
func TestLoadPOR(t *testing.T) {
	im, err := Load("../../por-go/original/START.EXE", 0xC7C)
	if err != nil {
		t.Skip(err)
	}
	if len(im.Mem) != 68592 {
		t.Fatalf("len %d", len(im.Mem))
	}
	if s := string(im.Mem[im.DS+0x5F1:][:23]); s != "C D M T V A R L S B E J" {
		t.Fatalf("DS:5F0 = %q", s)
	}
}

// Synthetic EXEPACK file: header 2 paragraphs, packed = "AB" copy command,
// then the EXEPACK block with "RB" and dest_len = 1 paragraph.
func TestLoadSynthetic(t *testing.T) {
	b := make([]byte, 32+16+18)
	copy(b, "MZ")
	binary.LittleEndian.PutUint16(b[8:], 2)    // header paragraphs
	binary.LittleEndian.PutUint16(b[0x16:], 1) // cs: EXEPACK block at 32+16
	for i := 37; i < 48; i++ {
		b[i] = 0xFF // padding
	}
	// packed (processed backwards): data "AB", count=2, cmd=0xB3 (copy, last)
	copy(b[32:], []byte{'A', 'B', 2, 0, 0xB3})
	v := b[48:]
	binary.LittleEndian.PutUint16(v[12:], 1)
	copy(v[16:], "RB")
	im, err := unpack(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(im.Mem) != 16 || string(im.Mem[14:]) != "AB" || im.DS != 48 {
		t.Fatalf("%q ds=%d", im.Mem, im.DS)
	}
	if im.Bytes(-48+14, 2) == nil || im.Bytes(100, 1) != nil {
		t.Fatal("Bytes bounds")
	}
	if _, err := Load("nonexistent", 0); err == nil {
		t.Fatal("want error")
	}
}
