package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
)

// The title (ovr02:0626 title_screen) is the presenter's: the adventure's
// Start sequence (SetStart), played until Continue ends it (titleDone).
// Until then Intro is set and the party menu waits.

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

// title is the adventure's Start sequence as the engine shows it (SetStart):
// which steps play; gobox composes them from the script.
type title struct {
	full, short game.Scene         // the title Scene with the Full intro option on, and off
	endAnim     bool               // the End starts with an animation (adventure.SeqEnding)
	end         *game.Scene        // the end sequence of pictures (SetEnd)
	after       []adventure.Effect // what the end leaves in the game (SetEnd)
	note        string             // the end's closing note (SetEnd)
}

// graphics is the run's graphics mode as the sequences' flashes take it.
func (e *Engine) graphics() adventure.Graphics {
	if e.Cfg.Graphics == 'E' { // DS:4A1E: the EGA driver
		return adventure.GraphicsEGA
	}
	return adventure.GraphicsOther
}

// titleScene is s's title Scene (full: the Full intro option), its
// pictures checked to be in the adventure's art.
func (e *Engine) titleScene(s adventure.Sequence, full bool) (game.Scene, error) {
	in, whole, err := adventure.TitleIntro(s, full, e.amigaUp(), e.graphics())
	if err != nil {
		return game.Scene{}, err
	}
	for _, n := range in.Pictures {
		if st := s.Step(n); st != nil {
			for _, l := range st.Layers {
				if err := e.hasArt(l.Art); err != nil {
					return game.Scene{}, fmt.Errorf("step %d: %w", n, err)
				}
			}
		}
	}
	return game.Scene{Kind: game.SceneTitle, Intro: in, Full: whole}, nil
}

// SetEnd takes the adventure's end sequence (its manifest's End; nil: none): what
// PROGRAM's ending and pod's save-and-quit play. A leading animation plays as a full-screen
// sequence (adventure.SeqEnding), picture steps as a scene of their own (title.end).
func (e *Engine) SetEnd(s *adventure.Sequence) error {
	e.title.endAnim, e.title.end, e.title.after, e.title.note = false, nil, nil, ""
	if s == nil {
		return nil
	}
	e.title.after, e.title.note = s.After, s.Note
	if s.LeadingAnim() != nil {
		e.title.endAnim = true
		return nil
	}
	sc, err := e.titleScene(*s, true)
	if err != nil {
		return fmt.Errorf("end: %w", err)
	}
	sc.Intro.End = true
	e.title.end = &sc
	return nil
}

// SetStart takes the adventure's title sequence (its manifest's Start; nil:
// none, the title waits for Continue with no steps). Its pictures are checked now.
func (e *Engine) SetStart(s *adventure.Sequence) error {
	if s == nil {
		s = &adventure.Sequence{}
	}
	t := title{endAnim: e.title.endAnim, end: e.title.end, after: e.title.after, note: e.title.note} // SetEnd's stay
	var err error
	if t.full, err = e.titleScene(*s, true); err != nil {
		return err
	}
	if t.short, err = e.titleScene(*s, false); err != nil {
		return err
	}
	for _, l := range s.Backdrop {
		if err := e.hasArt(l.Art); err != nil {
			return fmt.Errorf("backdrop: %w", err)
		}
	}
	e.title = t
	return nil
}

// errNoArt: a title picture the adventure folder does not have.
var errNoArt = errors.New("no such picture in the adventure's art")

// hasArt checks a sequence's picture is in the adventure's art: a DAX
// block (an animation's first frame stands for it), or a PNG in its art/.
func (e *Engine) hasArt(a adventure.Art) error {
	switch a.Kind {
	case adventure.ArtPNG:
		_, err := os.Stat(filepath.Join(e.adv, adventure.ArtDir, a.File))
		return err
	case adventure.ArtDAX:
		if e.LoadPic(a.File, a.Block) != nil {
			return nil
		}
		if b, _ := e.artBlocks(a.File, a.Block); b != nil && b.Kind == asset.Anim && len(b.Anim(asset.NoMask)) > 0 {
			return nil
		}
		return fmt.Errorf("%s %d: %w", a.File, a.Block, errNoArt)
	}
	return fmt.Errorf("art kind %d: %w", a.Kind, errNoArt)
}
