// Package amigaimport finds a Gold Box game's Amiga disks among .adf
// images (loose or zipped) and writes its graphics as DOS-layout DAX
// archives, plus the program disk's SOUND.BIN and the intro's TITLE.PNG.
// Pool of Radiance ships .dax archives; the later games GLIB libraries
// (.TLB, glib.go) spread over several disks. Which archives to convert,
// and how, is the game's data (Spec).
package amigaimport

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gobox/asset/orig"
)

// Logf reports everything the import leaves out (a file, a block, a frame):
// nothing is dropped unsaid. Tests replace it to collect the lines.
var Logf = log.Printf

// Gfx is how one archive converts.
type Gfx struct {
	Slot   byte // asset.Amiga palette (a 32-colour picture uses Slot and Slot+1)
	Split  bool // frames pack two tiles side by side (.dax)
	Sprite bool // animation block
	XOR    bool // .dax: sprite frames XORed against the first; .TLB: frames after the first are drawn over it (written whole)
	// .TLB only:
	Base   int  // an unindexed library's id of its entry 0 (SKY: 0xfa)
	Frames bool // an unindexed library is one block, id Base, of all its pictures (DUNGCOM: Base 1)
	Mask   byte // colour written for transparent pixels (the game's mask colour)
	// Pal is the file offset in Spec.Exe of the 16 12-bit RGB words that
	// colour this archive's pictures that carry no palette of their own.
	Pal int
}

// Spec is the game's data.
type Spec struct {
	Gfx     map[string]Gfx // lower-case archive name without .dax/.tlb
	MinGfx  int            // archives a disk must hold to be the data disk (without Disks)
	Program string         // file holding the sound players ("program")
	Intro   string         // file holding the painted title ("intro")
	Exe     string         // file holding the fixed palettes (Gfx.Pal)
	Font    string         // file holding the text font (written as FontFile; "": none)
	// Files are more files copied as they are, by base name, case-insensitive
	// (SotSB: "SECRET.MX", "sounds.glb"); each is written under the name given.
	Files []string
	// Disks are the volume names or file paths (case-insensitive) that mark
	// this game's disks. When set, every disk matching one is read and no
	// other: MinGfx is ignored, and of two releases the one whose marker
	// is listed wins.
	Disks []string
	// Skip are the disk files the import deliberately leaves out, by base
	// name (case-insensitive) with the reason; any other file that is not
	// converted, copied or a program/palette source is logged as unknown.
	Skip map[string]string
	// Alias merges a library into another archive, by base name: Pools of
	// Darkness' disk 1 8X8DA.TLB holds the 8X8D symbols of its own disk.
	Alias map[string]string
	// ProgramGfx are archives made of pictures the program holds
	// (Spec.Program's image, orig.AmigaProgramImage), by archive name.
	ProgramGfx map[string]ProgramGfx
}

// ProgramGfx is an archive of the program's own pictures: its blocks by
// id, coloured from palette slot Slot (asset.Amiga...).
type ProgramGfx struct {
	Slot   byte
	Blocks map[int]BitMap
}

// BitMap is a picture the program holds as an AmigaOS struct BitMap
// (u16 BytesPerRow, u16 Rows, u8 Flags, u8 Depth, u16 pad, the plane
// pointers) at address Addr of its image (orig.Base..orig.End); Width is
// the pixels from its left edge that the program blits of it (0: all).
type BitMap struct{ Addr, Width int }

// archive is the name a disk file's pictures are written under.
func (s Spec) archive(name string) string {
	if a, ok := s.Alias[name]; ok {
		return a
	}
	return name
}

// Import reads the .adf images under src and writes to out.
func Import(src, out string, s Spec) error {
	var disk []orig.AdfFile
	var snd []byte
	var title image.Image
	var exe, fnt []byte
	extra := map[string][]byte{}
	look := func(img []byte) bool {
		vol, files, err := orig.Read(img)
		if err != nil {
			return false
		}
		if len(s.Disks) > 0 {
			if !marked(vol, files, s.Disks) {
				return false
			}
			disk = append(disk, files...)
		}
		n := 0
		for _, f := range files {
			if s.Program != "" && strings.EqualFold(f.Path, s.Program) && snd == nil {
				snd = orig.AmigaProgramImage(f.Data)
			}
			if s.Font != "" && strings.EqualFold(f.Path, s.Font) && fnt == nil {
				fnt = f.Data
			}
			if s.Exe != "" && strings.EqualFold(f.Path, s.Exe) && exe == nil {
				exe = f.Data
			}
			for _, n := range s.Files {
				if strings.EqualFold(f.Path[strings.LastIndex(f.Path, "/")+1:], n) && extra[n] == nil {
					extra[n] = f.Data
				}
			}
			if s.Intro != "" && strings.EqualFold(f.Path, s.Intro) && title == nil {
				title = orig.Title(f.Data)
			}
			if _, ok := s.Gfx[strings.TrimSuffix(strings.ToLower(f.Path), ".dax")]; ok {
				n++
			}
		}
		if len(s.Disks) > 0 {
			return false // read them all
		}
		if n >= s.MinGfx {
			disk = files
		}
		return disk != nil && (snd != nil || s.Program == "")
	}
	filepath.WalkDir(src, func(p string, d fs.DirEntry, _ error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".adf":
			if b, err := os.ReadFile(p); err == nil && look(b) {
				return fs.SkipAll
			}
		case ".zip":
			z, err := zip.OpenReader(p)
			if err != nil {
				return nil
			}
			defer z.Close()
			for _, f := range z.File {
				if !strings.EqualFold(filepath.Ext(f.Name), ".adf") || strings.Contains(f.Name, "__MACOSX") {
					continue
				}
				r, err := f.Open()
				if err != nil {
					continue
				}
				b, _ := io.ReadAll(r)
				r.Close()
				if look(b) {
					return fs.SkipAll
				}
			}
		}
		return nil
	})
	if disk == nil && snd == nil {
		return errors.New("no Amiga data disk found")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if snd != nil {
		if err := os.WriteFile(filepath.Join(out, "SOUND.BIN"), snd, 0o644); err != nil {
			return err
		}
	}
	for n, b := range extra {
		if err := os.WriteFile(filepath.Join(out, n), b, 0o644); err != nil {
			return err
		}
	}
	if s.Font != "" {
		b, err := font(fnt)
		if err != nil {
			return fmt.Errorf("%s: %w", s.Font, err)
		}
		if err := os.WriteFile(filepath.Join(out, FontFile), b, 0o644); err != nil {
			return err
		}
	}
	if title != nil {
		var b bytes.Buffer
		png.Encode(&b, title)
		if err := os.WriteFile(filepath.Join(out, "TITLE.PNG"), b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	blocks, pals := map[string]map[int][]byte{}, map[string]map[int][]byte{}
	for _, f := range disk {
		name := strings.ToLower(f.Path)
		name = name[strings.LastIndex(name, "/")+1:]
		ext := filepath.Ext(name)
		if len(s.Disks) == 0 {
			name = strings.ToLower(f.Path) // por: the data disk's own names, as ever
			ext = ".dax"
		}
		name = s.archive(strings.TrimSuffix(name, ext))
		g, ok := s.Gfx[name]
		if !ok {
			continue
		}
		if blocks[name] == nil {
			blocks[name], pals[name] = map[int][]byte{}, map[int][]byte{}
		}
		switch ext {
		case ".dax":
			daxBlocks(name, f.Data, g, blocks[name])
		case ".tlb":
			tlbBlocks(name, f.Data, g, blocks[name], pals[name])
		}
	}
	logSkipped(disk, s)
	for name, pg := range s.ProgramGfx {
		if snd == nil {
			return fmt.Errorf("%s: no program image (Spec.Program %q) to read its pictures from", name, s.Program)
		}
		blocks[name] = map[int][]byte{}
		for id, bm := range pg.Blocks {
			b, err := bitMapBlock(snd, bm, pg.Slot)
			if err != nil {
				return fmt.Errorf("%s block %d: %w", name, id, err)
			}
			blocks[name][id] = b
		}
	}
	setArchives(blocks, pals)
	for name, b := range blocks {
		if p := s.Gfx[strings.TrimRight(name, "0123456789")].Pal; p > 0 && p+32 <= len(exe) {
			rgb := make([]byte, 48)
			for i := range rgb {
				w := int(exe[p+i/3*2])<<8 | int(exe[p+i/3*2+1])
				rgb[i] = byte(w>>(8-4*(i%3))&15) * 17
			}
			for id := range b {
				if pals[name][id] == nil {
					pals[name][id] = rgb
				}
			}
		}
		if err := os.WriteFile(filepath.Join(out, strings.ToUpper(name)+".DAX"), orig.WriteDOS(b), 0o644); err != nil {
			return err
		}
		if len(pals[name]) > 0 {
			if err := os.WriteFile(filepath.Join(out, strings.ToUpper(name)+"PAL.DAX"), orig.WriteDOS(pals[name]), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// skipReason is why a file is left out of every game's import (Spec.Skip
// adds the game's own).
var skipReason = map[string]string{
	"always.tlb":  "the Amiga's own UI pieces (compass, arrows, cursors): gobox draws the UI, and the DOS archives have no such pictures (a game whose window frame is here converts it: Spec.Gfx \"always\")",
	"ecl.glb":     "game data, not art or sound: gobox import reads it from the DOS originals",
	"geo.glb":     "game data, not art or sound: gobox import reads it from the DOS originals",
	"moncha.glb":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"monitm.glb":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"monspc.glb":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"item.glb":    "game data, not art or sound: gobox import reads it from the DOS originals",
	"walldef.glb": "game data, not art or sound: gobox import reads it from the DOS originals",
	"items.dat":   "item table (game data): gobox import reads it from the DOS originals",
	"item.dat":    "item table (game data): gobox import reads it from the DOS originals",
	"strings.glb": "text strings (game data): the DOS originals' text is the source",
	"curse.slb":   "SLBR sound library of the Amiga's player: not read (the songs, where the disks have them, are the .mx file)",
	"secret.slb":  "SLBR sound library of the Amiga's player: not read (the songs are SECRET.MX)",
	"ecl.dax":     "game data, not art or sound: gobox import reads it from the DOS originals",
	"geo.dax":     "game data, not art or sound: gobox import reads it from the DOS originals",
	"item.dax":    "game data, not art or sound: gobox import reads it from the DOS originals",
	"moncha.dax":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"monitm.dax":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"monspc.dax":  "game data, not art or sound: gobox import reads it from the DOS originals",
	"walldef.dax": "game data, not art or sound: gobox import reads it from the DOS originals",
}

// osDirs are the AmigaOS directories of a boot disk: no game data.
var osDirs = map[string]bool{"c": true, "devs": true, "l": true, "libs": true, "s": true, "fonts": true, "system": true}

// logSkipped reports the data disks' files that were neither converted nor
// copied, with the game's reason (Spec.Skip). AmigaOS files, icons, saved
// games (save) and files without an extension (programs) are not game data.
func logSkipped(disk []orig.AdfFile, s Spec) {
	for _, f := range disk {
		dir, _, _ := strings.Cut(strings.ToLower(f.Path), "/")
		base := f.Path[strings.LastIndex(f.Path, "/")+1:]
		lower, ext := strings.ToLower(base), strings.ToLower(filepath.Ext(base))
		name := strings.TrimSuffix(lower, ext)
		if _, gfx := s.Gfx[s.archive(name)]; gfx && (ext == ".tlb" || ext == ".dax") || ext == ".info" || ext == "" || osDirs[dir] || dir == "save" {
			continue
		}
		if strings.EqualFold(f.Path, s.Font) {
			continue // the text font (FontFile)
		}
		if slices.ContainsFunc(s.Disks, func(m string) bool { return strings.EqualFold(f.Path, m) }) {
			continue // the disk's marker file
		}
		if slices.ContainsFunc(s.Files, func(n string) bool { return strings.EqualFold(base, n[strings.LastIndex(n, "/")+1:]) }) {
			continue
		}
		if why, ok := firstReason(s.Skip[lower], skipReason[lower]); ok {
			Logf("amigaimport: skip %s: %s", f.Path, why)
		} else {
			Logf("amigaimport: skip %s: UNKNOWN (not in Gfx, Files or Skip)", f.Path)
		}
	}
}

// firstReason is the first non-empty reason.
func firstReason(reasons ...string) (string, bool) {
	for _, r := range reasons {
		if r != "" {
			return r, true
		}
	}
	return "", false
}

// marked: the disk's volume or one of its files is a marker.
func marked(vol string, files []orig.AdfFile, marks []string) bool {
	for _, m := range marks {
		if strings.EqualFold(vol, m) {
			return true
		}
		for _, f := range files {
			if strings.EqualFold(f.Path, m) {
				return true
			}
		}
	}
	return false
}

// daxBlocks converts a Pool of Radiance .dax archive.
func daxBlocks(name string, data []byte, g Gfx, blocks map[int][]byte) {
	a := orig.OpenAmiga(data)
	for _, e := range a.Entries {
		b := a.BlockShared(e.ID)
		if len(b) < 16 || e.ID > 255 {
			Logf("amigaimport: skip %s.dax block %d (%d bytes): too small or id past 255", name, e.ID, len(b))
			continue
		}
		if g.Sprite {
			blocks[e.ID] = orig.AmigaSprite(b, g.XOR, g.Slot)
		} else {
			blocks[e.ID] = orig.AmigaImage(b, g.Split, g.Slot)
		}
	}
}

// setArchives moves the ids past 255 of each archive (Curse HEAD 549: disk
// 5's own block 49) to the set's own one, name+set (HEAD5, block 49), which
// dax.Loader reads before the shared archive: a DOS archive keys a block
// by one byte.
func setArchives(blocks, pals map[string]map[int][]byte) {
	for name, b := range blocks {
		for id, blk := range b {
			if id < 256 {
				continue
			}
			set := name + strconv.Itoa(id/100)
			if blocks[set] == nil {
				blocks[set], pals[set] = map[int][]byte{}, map[int][]byte{}
			}
			blocks[set][id%100] = blk
			if p := pals[name][id]; p != nil {
				pals[set][id%100] = p
			}
			delete(b, id)
			delete(pals[name], id)
		}
	}
}

// bitMapDepth is the planes a program picture has: the four of the game's
// 16-colour screen.
const bitMapDepth = 4

// bitMapBlock is the picture of BitMap bm of the program image img
// (orig.Base..) as a DOS-layout block (orig.AmigaImage), slot its palette.
func bitMapBlock(img []byte, bm BitMap, slot byte) ([]byte, error) {
	at := func(a, n int) ([]byte, error) {
		o := a - orig.Base
		if o < 0 || o+n > len(img) {
			return nil, fmt.Errorf("address %#x+%d outside the program image %#x..%#x", a, n, orig.Base, orig.End)
		}
		return img[o : o+n], nil
	}
	hdr, err := at(bm.Addr, 8+4*bitMapDepth)
	if err != nil {
		return nil, err
	}
	bpr, rows, depth := be16(hdr, 0), be16(hdr, 2), int(hdr[5])
	if depth != bitMapDepth {
		return nil, fmt.Errorf("BitMap %#x has %d planes, want %d", bm.Addr, depth, bitMapDepth)
	}
	w := bpr // in bytes: 8-pixel cells
	if bm.Width > 0 {
		w = bm.Width / 8
	}
	stride := (w + 1) / 2 * 2 // orig.AmigaImage's rows: whole 16-pixel words
	if stride > bpr {
		return nil, fmt.Errorf("BitMap %#x: width %d past its %d bytes a row", bm.Addr, bm.Width, bpr)
	}
	raw := make([]byte, 12, 12+bitMapDepth*rows*stride)
	for i, v := range []int{rows, w, 0, 0, 1, rows * stride} { // h, w, x, y, frames, plane size
		raw[2*i], raw[2*i+1] = byte(v>>8), byte(v)
	}
	for p := range bitMapDepth {
		plane, err := at(be32(hdr, 8+4*p), bpr*rows)
		if err != nil {
			return nil, fmt.Errorf("BitMap %#x plane %d: %w", bm.Addr, p, err)
		}
		for y := range rows {
			raw = append(raw, plane[y*bpr:y*bpr+stride]...)
		}
	}
	return orig.AmigaImage(raw, false, slot), nil
}
