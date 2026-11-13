package asset

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// A sheet is one file of art: <name>.png, a paletted PNG holding every frame
// of every block stacked top to bottom, left-aligned, the pixels the
// original's colour indexes (0..15 EGA, from ColourBase on Amiga), and
// <name>.txt saying what each frame is:
//
//	# comment
//	block <id> image|anim|vga-image|vga-anim <colour base> <palette: hex RGB triplets, or -> <loop frame>
//	frame <width in 8-pixel cells> <height> <x> <y> <delay>
//
// A block's frames follow it; frames take the atlas rows in the order listed.
// An image block's frames share one size (Image.Count frames), an anim
// block's each have their own (Frame.Image).

// Kind is what a block holds.
type Kind string

const (
	Still Kind = "image" // one picture of Count frames (Block.Image)
	Anim  Kind = "anim"  // an animation (Block.Anim)
	// The VGA kinds' pixels index the block's Palette (256 RGB triplets)
	// directly: Buck Rogers', Pools of Darkness' pictures.
	VGAStill Kind = "vga-image"
	VGAAnim  Kind = "vga-anim"
)

// Pictured reports whether k is a still picture (Block.Image), not an animation.
func (k Kind) Pictured() bool { return k == Still || k == VGAStill }

// vga reports whether k's pixels index its block's own 256 colours.
func (k Kind) vga() bool { return k == VGAStill || k == VGAAnim }

// NoMask is a load without a transparent colour.
const NoMask = -1

// Block is one picture or animation of a sheet, by its id.
type Block struct {
	ID         int
	Kind       Kind
	ColourBase int    // added to the original colour indexes (Amiga palette slot); a mask colour c is ColourBase+c
	Palette    []byte // the RGB triplets the picture was drawn in (Amiga, VGA), nil: none of its own
	Loop       int    // a VGA animation's frame play goes back to after the last
	Frames     []Frame
}

// Image is the still block as a picture in its own colours (VGA, an Amiga
// block's own: Image.Palette); with mask >= 0 that colour is
// transparent (Clear), the original's masked load.
func (b *Block) Image(mask int) *Image {
	if !b.Kind.Pictured() || len(b.Frames) == 0 {
		return nil
	}
	f := b.Frames[0].Image
	im := NewImage(len(b.Frames), f.Width, f.Height)
	im.X, im.Y = f.X, f.Y
	for i, fr := range b.Frames {
		copy(im.Frame(i), fr.Image.Data)
	}
	im.Palette = b.colours()
	b.masked(im, mask)
	return im
}

// colours is the block's own colours as an Image.Palette, for its still
// and its animation alike: VGA's, else an Amiga block's (nil: EGA and the
// asset.Amiga slots).
func (b *Block) colours() []color.RGBA {
	if pal := b.vgaPalette(); pal != nil {
		return pal
	}
	return b.ownColours()
}

// ownColours is an Amiga block's own colours (cotab, sotsb, pod) as an
// Image.Palette: the block's from its ColourBase on, EGA below, Clear
// transparent; nil: none of its own (POR's Amiga art: the Amiga slots).
func (b *Block) ownColours() []color.RGBA {
	if len(b.Palette) == 0 || b.ColourBase < AmigaBase {
		return nil
	}
	pal := make([]color.RGBA, 256)
	for i := range 16 {
		pal[i] = EGARGBA(i)
	}
	for i := 0; i+2 < len(b.Palette) && b.ColourBase+i/3 < len(pal); i += 3 {
		pal[b.ColourBase+i/3] = color.RGBA{b.Palette[i], b.Palette[i+1], b.Palette[i+2], 255}
	}
	return pal
}

// Anim is the animation block's frames, each a copy; mask as Image.
func (b *Block) Anim(mask int) []Frame {
	if b.Kind != Anim && b.Kind != VGAAnim {
		return nil
	}
	out := make([]Frame, len(b.Frames))
	pal := b.colours()
	for i, f := range b.Frames {
		im := *f.Image
		im.Palette = pal
		im.Data = slices.Clone(f.Image.Data)
		b.masked(&im, mask)
		out[i] = Frame{Delay: f.Delay, Image: &im}
	}
	return out
}

// vgaPalette is a VGA block's Palette as Image.Palette (nil otherwise).
func (b *Block) vgaPalette() []color.RGBA {
	if !b.Kind.vga() {
		return nil
	}
	pal := make([]color.RGBA, len(b.Palette)/3)
	for i := range pal {
		pal[i] = color.RGBA{b.Palette[i*3], b.Palette[i*3+1], b.Palette[i*3+2], 255}
	}
	return pal
}

func (b *Block) masked(im *Image, mask int) {
	if mask == NoMask {
		return
	}
	im.Mask = Clear
	c := byte(b.ColourBase + mask)
	for i, p := range im.Data {
		if p == c {
			im.Data[i] = Clear
		}
	}
}

// WriteSheet writes blocks as dir/<name>.png and dir/<name>.txt; pal colours
// the PNG (the indexes are what counts, pal only makes it viewable).
func WriteSheet(dir, name string, blocks []Block, pal color.Palette) error {
	w := 0
	h := 0
	for _, b := range blocks {
		for _, f := range b.Frames {
			w, h = max(w, f.Image.Width*8), h+f.Image.Height
		}
	}
	atlas := image.NewPaletted(image.Rect(0, 0, max(w, 1), max(h, 1)), pal)
	var meta strings.Builder
	fmt.Fprintf(&meta, "# gobox/asset sheet %s.png: frames stacked top to bottom (asset/sheet.go)\n", name)
	y := 0
	for _, b := range blocks {
		p := "-"
		if b.Palette != nil {
			p = hex.EncodeToString(b.Palette)
		}
		fmt.Fprintf(&meta, "block %d %s %d %s %d\n", b.ID, b.Kind, b.ColourBase, p, b.Loop)
		for _, f := range b.Frames {
			im := f.Image
			fmt.Fprintf(&meta, "frame %d %d %d %d %d\n", im.Width, im.Height, im.X, im.Y, f.Delay)
			fw := im.Width * 8
			for r := range im.Height {
				copy(atlas.Pix[(y+r)*atlas.Stride:], im.Data[r*fw:(r+1)*fw])
			}
			y += im.Height
		}
	}
	var png1 bytes.Buffer
	if err := png.Encode(&png1, atlas); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, name+".png"), png1.Bytes(), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".txt"), []byte(meta.String()), 0o644)
}

// ReadSheet reads dir/<name>.png and .txt into its blocks by id.
func ReadSheet(dir, name string) (map[int]*Block, error) {
	f, err := os.Open(filepath.Join(dir, name+".png"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s.png: %w", name, err)
	}
	atlas, ok := img.(*image.Paletted)
	if !ok {
		return nil, fmt.Errorf("%s.png: not a paletted PNG", name)
	}
	meta, err := os.ReadFile(filepath.Join(dir, name+".txt"))
	if err != nil {
		return nil, err
	}
	blocks := map[int]*Block{}
	var cur *Block
	y := 0
	sc := bufio.NewScanner(bytes.NewReader(meta))
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		bad := func(err error) error { return fmt.Errorf("%s.txt:%d: %q: %w", name, n, line, err) }
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "block "):
			var b Block
			var p string
			if _, err := fmt.Sscanf(line, "block %d %s %d %s %d", &b.ID, &b.Kind, &b.ColourBase, &p, &b.Loop); err != nil {
				return nil, bad(err)
			}
			if b.Kind != Still && b.Kind != Anim && !b.Kind.vga() {
				return nil, bad(errors.New("unknown kind"))
			}
			if p != "-" {
				if b.Palette, err = hex.DecodeString(p); err != nil {
					return nil, bad(err)
				}
			}
			if blocks[b.ID] != nil {
				return nil, bad(errors.New("block listed twice"))
			}
			cur = &b
			blocks[b.ID] = cur
		case strings.HasPrefix(line, "frame "):
			var fr Frame
			im := &Image{Count: 1, Mask: NoMask}
			if _, err := fmt.Sscanf(line, "frame %d %d %d %d %d", &im.Width, &im.Height, &im.X, &im.Y, &fr.Delay); err != nil {
				return nil, bad(err)
			}
			if cur == nil {
				return nil, bad(errors.New("frame before any block"))
			}
			fw := im.Width * 8
			if y+im.Height > atlas.Rect.Dy() || fw > atlas.Rect.Dx() {
				return nil, bad(errors.New("frame outside the atlas"))
			}
			if cur.Kind.Pictured() && len(cur.Frames) > 0 && (fw != cur.Frames[0].Image.Width*8 || im.Height != cur.Frames[0].Image.Height) {
				return nil, bad(errors.New("an image block's frames differ in size"))
			}
			im.BPP = fw * im.Height
			im.Data = make([]byte, im.BPP)
			for r := range im.Height {
				copy(im.Data[r*fw:(r+1)*fw], atlas.Pix[(y+r)*atlas.Stride:])
			}
			y += im.Height
			fr.Image = im
			cur.Frames = append(cur.Frames, fr)
		default:
			return nil, bad(errors.New("not a block or frame line"))
		}
	}
	return blocks, sc.Err()
}

// Version is one release's art: a folder of sheets under the art folder.
type Version string

const (
	DOS   Version = "dos"
	Amiga Version = "amiga"
)

// Art reads the sheets of an art folder (<dir>/<version>/<file>), each sheet
// once, when first asked.
type Art struct {
	dir    string
	mu     sync.Mutex
	sheets map[string]map[int]*Block
}

// OpenArt is the art in dir.
func OpenArt(dir string) *Art { return &Art{dir: dir, sheets: map[string]map[int]*Block{}} }

// Has reports whether version v's art is there.
func (a *Art) Has(v Version) bool {
	_, err := os.Stat(filepath.Join(a.dir, string(v)))
	return err == nil
}

// Block is block id of file in version v; nil when v, the file or the block
// is not there. A sheet that is there but does not read is a broken import:
// it panics.
func (a *Art) Block(v Version, file string, id int) *Block {
	key := string(v) + "/" + strings.ToUpper(file)
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sheets[key]
	if !ok {
		var err error
		s, err = ReadSheet(filepath.Join(a.dir, string(v)), strings.ToUpper(file))
		if errors.Is(err, fs.ErrNotExist) {
			s, err = nil, nil
		}
		if err != nil {
			panic(fmt.Sprintf("art %s: %v (run the import again)", key, err))
		}
		a.sheets[key] = s
	}
	return s[id]
}

// FirstID is file's lowest block id in version v; false: no blocks.
func (a *Art) FirstID(v Version, file string) (int, bool) {
	a.Block(v, file, 0) // reads the sheet
	a.mu.Lock()
	defer a.mu.Unlock()
	ids := slices.Collect(maps.Keys(a.sheets[string(v)+"/"+strings.ToUpper(file)]))
	if len(ids) == 0 {
		return 0, false
	}
	return slices.Min(ids), true
}

// Dir is the imported art of a game whose original archives are in dosDir:
// assets/art beside it (as the Amiga's amiga/).
func Dir(dosDir string) string { return filepath.Join(filepath.Dir(dosDir), "assets", "art") }
