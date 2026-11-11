package dax

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gobox/asset"
	"gobox/ega"
)

// ArtFormat changes when the sheets an import writes change: an older art
// folder is stale (ImportArtIfStale).
const ArtFormat = "3"

// ImportArt is the art import: every block of dosDir's DAX archives that
// holds a picture or an animation (artBlock), as the loader gives it,
// becomes a gobox/asset sheet, the DOS art in dst/dos and the Amiga blocks
// of amigaDir standing in for it (UseAmiga) in dst/amiga, under the DOS
// name; dst/key.txt names what was read. Nothing at run time reads the
// archives' art.
func ImportArt(dosDir, amigaDir, dst string) error {
	l, err := NewLoader(dosDir)
	if err != nil {
		return err
	}
	l.UseAmiga(amigaDir)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	files, err := daxNames(dosDir)
	if err != nil {
		return err
	}
	amigaFiles := files
	if l.HasAmiga() { // the names the Amiga archives stand in for (AmigaShared): a set's own, the shared one's sets
		am, err := daxNames(amigaDir)
		if err != nil {
			return err
		}
		for _, n := range am {
			amigaFiles = append(amigaFiles, n)
			if strings.TrimRight(n, "0123456789") == n {
				for set := 1; set <= 9; set++ {
					amigaFiles = append(amigaFiles, n+strconv.Itoa(set))
				}
			}
		}
		slices.Sort(amigaFiles)
		amigaFiles = slices.Compact(amigaFiles)
	}
	pal := sheetPalette()
	for _, v := range []asset.Version{asset.DOS, asset.Amiga} {
		if v == asset.Amiga && !l.HasAmiga() {
			continue
		}
		dir := filepath.Join(dst, string(v))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		names := files
		if v == asset.Amiga {
			names = amigaFiles
		}
		for _, name := range names {
			dos := l.File(name)
			var blocks []asset.Block
			for id := range 256 {
				var b, p []byte
				if dos != nil {
					b = dos.BlockShared(id)
				}
				if v == asset.Amiga {
					if b = l.AmigaShared(name, id); b != nil {
						p = l.AmigaPalette(name, id)
					}
				}
				if b == nil {
					continue
				}
				blk, ok := artBlock(name, id, b)
				if !ok {
					continue
				}
				if p != nil {
					blk.Palette = p
				}
				blocks = append(blocks, blk)
			}
			if len(blocks) == 0 {
				continue
			}
			if err := asset.WriteSheet(dir, name, blocks, pal); err != nil {
				return err
			}
		}
	}
	key, err := SourceKey(ArtFormat, dosDir, amigaDir)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, "key.txt"), []byte(key), 0o644)
}

// ImportArtIfStale runs ImportArt when dst is missing or was imported from
// other archives or by another format; without dosDir there is nothing to
// import.
func ImportArtIfStale(dosDir, amigaDir, dst string) error {
	if _, err := os.Stat(dosDir); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	key, err := SourceKey(ArtFormat, dosDir, amigaDir)
	if err != nil {
		return err
	}
	have, err := os.ReadFile(filepath.Join(dst, "key.txt"))
	if err == nil && string(have) == key {
		return nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return ImportArt(dosDir, amigaDir, dst)
}

// daxNames are dir's DAX archives' names, upper case, without .DAX.
func daxNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if n := e.Name(); strings.EqualFold(filepath.Ext(n), ".dax") {
			out = append(out, strings.ToUpper(strings.TrimSuffix(n, filepath.Ext(n))))
		}
	}
	return out, nil
}

// SourceKey names what an import of format reads from dirs: the format and
// every DAX archive's size and time (an import is stale when it changes).
func SourceKey(format string, dirs ...string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "format %s\n", format)
	for _, d := range dirs {
		ents, err := os.ReadDir(d)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		for _, e := range ents {
			if !strings.EqualFold(filepath.Ext(e.Name()), ".dax") {
				continue
			}
			i, err := e.Info()
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "%s/%s %d %d\n", filepath.Base(d), e.Name(), i.Size(), i.ModTime().UnixNano())
		}
	}
	return b.String(), nil
}

// XORed reports whether file's animations store their frames after the
// first XORed against it (PIC*, FINAL*).
func XORed(file string) bool {
	up := strings.ToUpper(file)
	return strings.HasPrefix(up, "PIC") || strings.HasPrefix(up, "FINAL")
}

// artBlock decodes block b of file unmasked: an EGA picture or animation
// (the layout it fills exactly, blockKind; a picture where both fit), else
// a VGA one (ParseVGA, ParseVGAAnim); ok is false for a block that is
// neither (a script, a map, a font). An EGA block's colour base is its
// Amiga tag's slot.
func artBlock(file string, id int, b []byte) (blk asset.Block, ok bool) {
	blk = asset.Block{ID: id}
	switch blockKind(b) {
	case asset.Still:
		im, err := ParseImage(b, false, 0)
		if err != nil { // blockKind's Still is the exact size ParseImage reads
			return blk, false
		}
		blk.Kind, blk.ColourBase = asset.Still, colourBase(b[9:])
		blk.Frames = stillFrames(im)
	case asset.Anim:
		blk.Kind, blk.Frames = asset.Anim, ParseSprite(b, false, 0, XORed(file))
		if len(b) > 22 { // the first frame's header holds the tag
			blk.ColourBase = colourBase(b[13:])
		}
	default:
		if im := ParseVGA(b, false, 0); im != nil {
			blk.Kind, blk.Palette, blk.Frames = asset.VGAStill, rgb(im.Palette), stillFrames(im)
		} else if fs, loop := ParseVGAAnim(b); fs != nil {
			blk.Kind, blk.Palette, blk.Loop, blk.Frames = asset.VGAAnim, rgb(fs[0].Image.Palette), loop, fs
			for _, f := range fs {
				f.Image.Palette = nil // the block's
			}
		} else {
			return blk, false
		}
	}
	return blk, true
}

// stillFrames are im's frames, each an image of its own.
func stillFrames(im *Image) []asset.Frame {
	var out []asset.Frame
	for i := range im.Count {
		fr := NewImage(1, im.Width, im.Height)
		fr.X, fr.Y = im.X, im.Y
		copy(fr.Data, im.Frame(i))
		out = append(out, asset.Frame{Image: fr})
	}
	return out
}

// rgb is pal as RGB triplets (Block.Palette).
func rgb(pal []color.RGBA) []byte {
	out := make([]byte, 0, len(pal)*3)
	for _, c := range pal {
		out = append(out, c.R, c.G, c.B)
	}
	return out
}

// blockKind is the EGA layout b fills exactly: a picture's (ParseImage,
// also where an animation's fits too) or an animation's (ParseSprite); ""
// for neither.
func blockKind(b []byte) asset.Kind {
	still := len(b) >= 17 && 17+int(b[8])*le16(b, 0)*le16(b, 2)*8/packed(b[9:]) == len(b)
	anim := len(b) >= 1
	p := 1
	for f := 0; anim && f < int(b[0]); f++ {
		if p+21 > len(b) {
			anim = false
			break
		}
		p += 21 + le16(b, p+4)*le16(b, p+6)*8/packed(b[p+12:])
	}
	switch {
	case still:
		return asset.Still
	case anim && p == len(b):
		return asset.Anim
	}
	return ""
}

func le16(b []byte, i int) int { return int(b[i]) | int(b[i+1])<<8 }

// packed is how many pixels a byte of a picture holds.
func packed(hdr []byte) int {
	if wide(hdr) {
		return 1
	}
	return 2
}

// colourBase is where amigaTag moves a tagged header's colours (0: untagged).
func colourBase(hdr []byte) int {
	if len(hdr) < 3 || hdr[0] != 'A' || hdr[1] != 'M' {
		return 0
	}
	return ega.AmigaBase + 16*int(hdr[2])
}

// sheetPalette colours a sheet's PNG: the EGA colours, the rest grey (an
// Amiga picture's own colours are its block's Palette).
func sheetPalette() color.Palette {
	p := make(color.Palette, 256)
	for i := range p {
		p[i] = color.Gray{Y: byte(i)}
	}
	for i := range 16 {
		p[i] = ega.RGBA(i)
	}
	return p
}
