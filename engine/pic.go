package engine

import (
	"path/filepath"
	"strconv"
	"sync/atomic"

	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
)

// DOSArt (extra) is the DOS art shown (F12, the Graphics page). State is
// art-independent (Picture.Alt carries the DOS twin, the front picks);
// only the title sequence's steps and the icon designer's colours with no
// combat picture (amigaUp) read it.
var DOSArt atomic.Bool

// amigaUp is whether the Amiga art is shown: DOSArt off and the adventure
// has Amiga art (every game's import has it: por, cotab, sotsb, pod).
func (e *Engine) amigaUp() bool { return !DOSArt.Load() && e.artFolder().Has(asset.Amiga) }

// LoadSymbolSet = ovr35:003B — load 8X8D<n>.DAX block into set index,
// masked on colour 13.
func (e *Engine) LoadSymbolSet(l adventure.SymbolLoad) {
	if l.Set < 0 || l.Set > 4 {
		return
	}
	file := l.File
	if file == "" {
		file = e.set.File("8X8D")
	}
	e.sym[l.Set] = e.LoadPicMasked(file, l.Block, 0xD)
	if e.sym[l.Set] == nil {
		e.PromptRow24(game.InkWhite, "Unable to load "+file+"-"+strconv.Itoa(l.Block))
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
		e.arts = asset.OpenArt(filepath.Join(e.adv, adventure.ArtDir))
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

// named records where im came from (Pictures' File and Block).
func named(im *asset.Image, file string, block int) *asset.Image {
	if im != nil {
		im.File, im.Block = file, block
	}
	return im
}

// LoadPic = 0464:0111 (no masking).
func (e *Engine) LoadPic(name string, id int) *asset.Image { return e.loadPic(name, id, asset.NoMask) }

// framePicture is frame f of im as the game shows it (mask: the colour
// left out, -1 none), with its DOS twin's whole frame (F12); the
// presenter's to colour and place. nil im: none.
func framePicture(im *asset.Image, f, mask int) game.Picture {
	if im == nil || f >= im.Count {
		return game.Picture{}
	}
	g := framePix(im, f, mask)
	g.File, g.Block, g.Frame = im.File, im.Block, im.Base+f
	if a := im.Alt; a != nil && f < a.Count {
		alt := framePix(a, f, mask)
		if a.Width == im.Width && a.Height == im.Height { // fitted to im (twin): in its place
			alt.CellX, alt.CellY = g.CellX, g.CellY
		}
		g.Alt = &alt
	}
	return g
}

// framePix is frame f of im's own pixels and offset (framePicture).
func framePix(im *asset.Image, f, mask int) game.Picture {
	w := im.Width * 8
	d := im.Frame(f)
	g := game.Picture{W: w, H: len(d) / w, Pixels: d, Palette: im.Palette, Clear: -1, CellX: im.X, CellY: im.Y}
	if mask >= 0 && mask < 256 {
		g.Clear = mask
	}
	return g
}
