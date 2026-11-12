package pool

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"

	"gobox/asset"
	"gobox/game"
	"gobox/game/play"
	"gobox/importer"
)

// fullPicture draws a whole-screen picture (320x200) in the middle of the
// page: on a page of another shape (a full screen) it gets black bars,
// never a crop or a stretch.
func (e *Engine) fullPicture(pic *asset.Image) {
	e.ClearScreen()
	e.DrawPicture((textCols-minCols)/2, (textRows-minRows)/2, pic, 0)
	e.snap.full = true
}

// The title (ovr02:0626 title_screen) is the presenter's: the adventure's
// Start sequence (SetStart), played until Continue ends it (titleDone).
// Until then View.Intro is set and the party menu waits.

// titleDone ends the title: the party menu comes up.
func (e *Engine) titleDone() {
	if !e.ext.intro {
		return
	}
	e.ext.intro = false
	e.SetPaletteColor(1, 1)
	e.ClearScreen()
	if !e.titleMusic() { // extra: the title plays on into the main menu
		e.Music(0)
	}
}

// title is the adventure's Start sequence as the engine shows it (SetStart).
type title struct {
	full, short play.Scene // the title Scene with the Full intro option on, and off
	credits     []game.Credit
	boxes       []image.Rectangle
	backdrop    []game.Layer
	remaster    [2]string
	seqs        []game.Sequence // the full-screen animations State.Sequence names (SetEnd)
}

// seqEnding is the name of the end sequence (game.Data.Sequences).
const seqEnding = "ending"

// SetEnd takes the adventure's end sequence (its manifest's End; nil: none): the
// ending ProgramSaveQuit plays before the run ends.
func (e *Engine) SetEnd(s *importer.Sequence) {
	e.title.seqs = nil
	if s == nil {
		return
	}
	if q := play.LeadingAnim(*s, seqEnding); q != nil {
		e.title.seqs = append(e.title.seqs, *q)
	}
}

// hasSequence is whether the sequence name was given.
func (e *Engine) hasSequence(name string) bool {
	return slices.ContainsFunc(e.title.seqs, func(q game.Sequence) bool { return q.Name == name })
}

// SetStart takes the adventure's title sequence (its manifest's Start; nil:
// none, the title waits for Continue with no steps). Its pictures load now.
func (e *Engine) SetStart(s *importer.Sequence) error {
	if s == nil {
		s = &importer.Sequence{}
	}
	var t title
	var err error
	g := importer.GraphicsOther
	if e.Cfg.Graphics == 'E' { // DS:4A1E: the EGA driver
		g = importer.GraphicsEGA
	}
	if t.full, err = play.TitleScene(*s, true, g, play.Common{}, e.titleArt); err != nil {
		return err
	}
	if t.short, err = play.TitleScene(*s, false, g, play.Common{}, e.titleArt); err != nil {
		return err
	}
	if t.credits, t.boxes, t.backdrop, err = play.TitleData(*s, e.titleArt); err != nil {
		return err
	}
	t.remaster = s.Remaster
	e.title = t
	return nil
}

// errNoArt: a title picture the adventure folder does not have.
var errNoArt = errors.New("no such picture in the adventure's art")

// titleArt loads a title sequence's picture: a DAX block, or a PNG in the
// adventure's art/.
func (e *Engine) titleArt(a importer.Art) (game.Picture, error) {
	switch a.Kind {
	case importer.ArtPNG:
		p := filepath.Join(e.adv, ArtDir, a.File)
		if _, err := os.Stat(p); err != nil {
			return game.Picture{}, err
		}
		return game.Picture{Source: game.SourcePNG, File: p}, nil
	case importer.ArtDAX:
		im := e.LoadPic(a.File, a.Block)
		if im == nil {
			return game.Picture{}, fmt.Errorf("%s %d: %w", a.File, a.Block, errNoArt)
		}
		return game.Picture{File: a.File, Block: a.Block, W: im.Width * 8, H: im.Height, Pixels: im.Frame(0), Clear: -1}, nil
	}
	return game.Picture{}, fmt.Errorf("art kind %d: %w", a.Kind, errNoArt)
}

// Title logo (extra): the remastered main menu shows the title picture's
// lettering on black, cut by the presenter (presentdata's LogoCut) from
// TITLE picture 2's pixels.

// titleLogo is TITLE picture 2's pixels (nil if missing), read once.
func (e *Engine) titleLogo() *Picture {
	if e.snap.logo != nil {
		return e.snap.logo
	}
	im := e.LoadPic("TITLE", 2)
	if im == nil {
		return nil
	}
	e.snap.logo = &Picture{Pix: im.Frame(0), PixW: im.Width * 8}
	return e.snap.logo
}
