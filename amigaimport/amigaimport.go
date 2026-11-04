// Package amigaimport finds a Gold Box game's Amiga data disk among .adf
// images (loose or zipped) and writes its graphics as DOS-layout DAX
// archives, plus the program disk's SOUND.BIN and the intro's TITLE.PNG.
// Which archives to convert, and how, is the game's data (Spec).
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
	"strings"

	"gobox/adf"
	"gobox/amiga"
	"gobox/dax"
)

// Gfx is how one archive converts.
type Gfx struct {
	Slot   byte // ega.Amiga palette
	Split  bool // frames pack two tiles side by side
	Sprite bool // animation block
	XOR    bool // sprite frames XORed against the first
}

// Spec is the game's data.
type Spec struct {
	Gfx     map[string]Gfx // lower-case archive name without .dax
	MinGfx  int            // archives a disk must hold to be the data disk
	Program string         // file holding the sound players ("program")
	Intro   string         // file holding the painted title ("intro")
}

// Import reads the .adf images under src and writes to out.
func Import(src, out string, s Spec) error {
	var disk []adf.File
	var snd []byte
	var title image.Image
	look := func(img []byte) bool {
		_, files, err := adf.Read(img)
		if err != nil {
			return false
		}
		n := 0
		for _, f := range files {
			if s.Program != "" && strings.EqualFold(f.Path, s.Program) && snd == nil {
				snd = amiga.Image(f.Data)
			}
			if s.Intro != "" && strings.EqualFold(f.Path, s.Intro) && title == nil {
				title = amiga.Title(f.Data)
			}
			if _, ok := s.Gfx[strings.TrimSuffix(strings.ToLower(f.Path), ".dax")]; ok {
				n++
			}
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
	if title != nil {
		var b bytes.Buffer
		png.Encode(&b, title)
		if err := os.WriteFile(filepath.Join(out, "TITLE.PNG"), b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	for _, f := range disk {
		name := strings.TrimSuffix(strings.ToLower(f.Path), ".dax")
		g, ok := s.Gfx[name]
		if !ok {
			continue
		}
		a := dax.OpenAmiga(f.Data)
		blocks := map[int][]byte{}
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
		if err := os.WriteFile(filepath.Join(out, strings.ToUpper(name)+".DAX"), dax.WriteDOS(blocks), 0o644); err != nil {
			return err
		}
	}
	return nil
}
