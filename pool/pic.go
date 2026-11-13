package pool

import (
	"gobox/ega"
	"image/color"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"gobox/asset"
)

// DOSArt (extra) shows the DOS pictures, portraits, animations and combat
// icons in place of the Amiga ones (F12, the Graphics page; Image.Alt).
var DOSArt atomic.Bool

// amigaUp is whether the Amiga art is shown: DOSArt off and the adventure
// has Amiga art (every game's import has it: por, cotab, sotsb, pod).
func (e *Engine) amigaUp() bool { return !DOSArt.Load() && e.artFolder().Has(asset.Amiga) }

// ownColours is an Amiga block's own colours (cotab, sotsb, pod) as an
// Image.Palette: the block's from its ColourBase on, EGA below, Clear
// transparent; nil: none of its own (POR's Amiga art: the ega.Amiga slots).
// Without it the presenter colours every game's Amiga art in POR's slots.
func ownColours(b *asset.Block) []color.RGBA {
	if len(b.Palette) == 0 || b.ColourBase < ega.AmigaBase {
		return nil
	}
	pal := make([]color.RGBA, 256)
	for i := range 16 {
		pal[i] = ega.RGBA(i)
	}
	for i := 0; i+2 < len(b.Palette) && b.ColourBase+i/3 < len(pal); i += 3 {
		pal[b.ColourBase+i/3] = color.RGBA{b.Palette[i], b.Palette[i+1], b.Palette[i+2], 255}
	}
	return pal
}

// symFix = DS:2722 / coab symbol_set_fix: first symbol id of each set.
var symFix = [5]int{0x001, 0x02E, 0x074, 0x0BA, 0x100}

// LoadSymbolSet = ovr35:003B — load 8X8D<n>.DAX block into set index,
// masked on colour 13.
func (e *Engine) LoadSymbolSet(l SymbolLoad) {
	if l.Set < 0 || l.Set > 4 {
		return
	}
	file := l.File
	if file == "" {
		file = e.set.File("8X8D")
	}
	e.sym[l.Set] = e.LoadPicMasked(file, l.Block, 0xD)
	if e.sym[l.Set] == nil {
		e.PromptRow24(ega.White, "Unable to load "+file+"-"+strconv.Itoa(l.Block))
		e.Exit()
	}
}

// LoadPicMasked = 0464:0111 with the mask flag set.
func (e *Engine) LoadPicMasked(name string, id int, mask byte) *asset.Image {
	return e.loadPic(name, id, int(mask))
}

// loadPic is picture id of file name from the imported art (mask: its
// transparent colour, or asset.NoMask): the Amiga one where there is one,
// with its DOS twin (Image.Alt), else the DOS one.
func (e *Engine) loadPic(name string, id, mask int) *asset.Image {
	b, dos := e.artBlocks(name, id)
	if b == nil || !b.Kind.Pictured() {
		return nil
	}
	im := named(b.Image(mask), name, id)
	if im.Palette == nil { // a VGA block's is set
		im.Palette = ownColours(b)
	}
	if dos != nil && dos.Kind.Pictured() {
		twin(im, named(dos.Image(mask), name, id))
	}
	return im
}

// artBlocks is block id of file name as shown, the Amiga one where there is
// one (with the DOS one it stands in for), else the DOS one (dos nil).
func (e *Engine) artBlocks(name string, id int) (b, dos *asset.Block) {
	a := e.artFolder()
	if b := a.Block(asset.Amiga, name, id); b != nil {
		return b, a.Block(asset.DOS, name, id)
	}
	return a.Block(asset.DOS, name, id), nil
}

// artFolder is the imported art of the engine's data (rules.ImportArt),
// opened on first use.
func (e *Engine) artFolder() *asset.Art {
	if e.arts == nil {
		e.arts = asset.OpenArt(filepath.Join(e.adv, ArtDir))
	}
	return e.arts
}

// twin makes alt im's Image.Alt (F12's DOS art): a smaller one fitted to
// im's size, a bigger one kept at its own (the page centres it on im's place).
func twin(im, alt *asset.Image) {
	if im == nil || alt == nil {
		return
	}
	if alt.Width <= im.Width && alt.Height <= im.Height && (alt.Width != im.Width || alt.Height != im.Height) {
		alt = alt.Fit(im.Width, im.Height)
	}
	im.Alt = alt
}

// art is the picture shown for frame f of im: its DOS twin while DOSArt is
// on (F12 and the Graphics page switch the Amiga art for the DOS art).
func (e *Engine) art(im *asset.Image, f int) *asset.Image {
	if im != nil && im.Alt != nil && f < im.Alt.Count && DOSArt.Load() {
		return im.Alt
	}
	return im
}

// named records where im came from (View.Pictures' File and Block).
func named(im *asset.Image, file string, block int) *asset.Image {
	if im != nil {
		im.File, im.Block = file, block
	}
	return im
}

// LoadPic = 0464:0111 (no masking).
func (e *Engine) LoadPic(name string, id int) *asset.Image { return e.loadPic(name, id, asset.NoMask) }

// DrawPicture = 0464:1951 draw_picture(x, y, pic): blit frame at cell (x,y).
func (e *Engine) DrawPicture(x, y int, im *asset.Image, frame int) {
	if im == nil || frame >= im.Count {
		return
	}
	d := im.Frame(frame)
	w := im.Width * 8
	e.picture(x*8, y*8, w, im.Height, d, -1, im, frame)
}

// DrawPictureMasked = 0464:08D4 (x, y, frame, mode): like DrawPicture;
// with mode bit 0 and a masked picture the mask pixels are not drawn.
func (e *Engine) DrawPictureMasked(x, y int, im *asset.Image, frame, mode int) {
	masked := mode&1 != 0
	if im == nil || frame >= im.Count {
		return
	}
	d := im.Frame(frame)
	w := im.Width * 8
	m := -1
	if masked {
		m = im.Mask
	}
	e.picture(x*8, y*8, w, im.Height, d, m, im, frame)
}
