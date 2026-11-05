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
	"image"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gobox/adf"
	"gobox/amiga"
	"gobox/dax"
)

// Gfx is how one archive converts.
type Gfx struct {
	Slot   byte // ega.Amiga palette (a 32-colour picture uses Slot and Slot+1)
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
	// Files are more files copied as they are, by base name, case-insensitive
	// (SotSB: "SECRET.MX", "sounds.glb"); each is written under the name given.
	Files []string
	// Disks are the volume names or file paths (case-insensitive) that mark
	// this game's disks. When set, every disk matching one is read and no
	// other: MinGfx is ignored, and of two releases the one whose marker
	// is listed wins.
	Disks []string
}

// Import reads the .adf images under src and writes to out.
func Import(src, out string, s Spec) error {
	var disk []adf.File
	var snd []byte
	var title image.Image
	var exe []byte
	extra := map[string][]byte{}
	look := func(img []byte) bool {
		vol, files, err := adf.Read(img)
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
				snd = amiga.Image(f.Data)
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
				title = amiga.Title(f.Data)
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
		name = strings.TrimSuffix(name, ext)
		g, ok := s.Gfx[name]
		if !ok {
			continue
		}
		if blocks[name] == nil {
			blocks[name], pals[name] = map[int][]byte{}, map[int][]byte{}
		}
		switch ext {
		case ".dax":
			daxBlocks(f.Data, g, blocks[name])
		case ".tlb":
			tlbBlocks(f.Data, g, blocks[name], pals[name])
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
		if err := os.WriteFile(filepath.Join(out, strings.ToUpper(name)+".DAX"), dax.WriteDOS(b), 0o644); err != nil {
			return err
		}
		if len(pals[name]) > 0 {
			if err := os.WriteFile(filepath.Join(out, strings.ToUpper(name)+"PAL.DAX"), dax.WriteDOS(pals[name]), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// marked: the disk's volume or one of its files is a marker.
func marked(vol string, files []adf.File, marks []string) bool {
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
func daxBlocks(data []byte, g Gfx, blocks map[int][]byte) {
	a := dax.OpenAmiga(data)
	for _, e := range a.Entries {
		b := a.BlockShared(e.ID)
		if len(b) < 16 || e.ID > 255 {
			continue
		}
		if g.Sprite {
			blocks[e.ID] = dax.AmigaSprite(b, g.XOR, g.Slot)
		} else {
			blocks[e.ID] = dax.AmigaImage(b, g.Split, g.Slot)
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
