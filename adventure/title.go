package adventure

import (
	"errors"
	"fmt"
	"image"
	"slices"

	"gobox/game"
)

// Sequence is a title or end sequence as data: its steps in order, then the
// backdrop that stays under the main menu. gobox plays it; a key or click
// ends a step's hold and goes no further (front/latch.go).
type Sequence struct {
	Steps    []SequenceStep
	Credits  []Credit `json:",omitempty"` // the credits pages' lines, placed
	Boxes    []Cells  `json:",omitempty"` // the credits' boxes (nil: one frame on the page's edges)
	Backdrop []Layer  `json:",omitempty"` // under the main menu after the steps (COTAB: painting and logo), the menu zooming and fading in over it
	Short    []int    `json:",omitempty"` // the steps (indexes) a short intro plays (COTAB: SSI logo only unless the config asks); nil: all
	Full     bool     `json:",omitempty"` // the full intro by default (the original's config: POR's POOL.CFG); the host's "Full intro" option overrides it
	After    []Effect `json:",omitempty"` // what an end sequence leaves in the game once it is over (the ruleset does it)
	Note     string   `json:",omitempty"` // an end sequence's closing note, shown once it is over
}

// Effect is what an end sequence does to the game when it is over; the ruleset
// carries it out (core.HealParty, core.SetWord).
type Effect int

const (
	EffectHealParty Effect = iota + 1 // every party member back to full hit points and well
	EffectOpenHall                    // the training hall teaches (the party menu offers it)
)

// StepKind is what a sequence step shows.
type StepKind int

const (
	StepPicture StepKind = iota // Layers over the page
	StepCredits                 // credits page Page
	StepAnim                    // a full-screen animation (Anim)
	StepText                    // Text alone, no picture
)

// SequenceStep is one step: what it shows, its tune and how long it holds.
type SequenceStep struct {
	Kind      StepKind
	Layers    []Layer  `json:",omitempty"` // StepPicture
	Text      []string `json:",omitempty"` // StepText, StepPicture: the paragraphs said over the step (colours are gobox's)
	Page      int      `json:",omitempty"` // StepCredits
	Anim      string   `json:",omitempty"` // StepAnim: the animation's base name in the folder's art
	Song      string   `json:",omitempty"` // the tune it starts (a song name or the game's song number)
	HoldMs    int      `json:",omitempty"` // 0: until a key or click
	Lightning Flash    `json:",omitempty"` // when the picture flashes (POR TITLE 2: FlashEGA)
	Whole     bool     `json:",omitempty"` // the layers are one whole-screen page
	Credits   bool     `json:",omitempty"` // StepPicture: the picture is the game's credits (PoD TITLE 3, SotSB's Amiga card): gobox adds its remaster credit line over it, as to a StepCredits page
	Edition   Edition  `json:",omitempty"` // the art it is shown with (SotSB's Amiga credits card: EditionAmiga)
}

// Edition is which release's art a step belongs to.
type Edition uint8

const (
	EditionAny   Edition = iota // shown with every art
	EditionAmiga                // shown only while the Amiga art is up (TitleIntro)
)

// Flash is when a step's picture flashes with lightning.
type Flash int

const (
	FlashNone Flash = iota // never
	FlashEGA               // in EGA graphics only (POR: ovr02 title_screen under the EGA driver)
)

// Graphics is the run's graphics mode, as the game's config picks it.
type Graphics int

const (
	GraphicsOther Graphics = iota // any mode without its own rule here (VGA, Tandy, CGA)
	GraphicsEGA
)

// On reports whether f flashes in graphics mode g.
func (f Flash) On(g Graphics) bool { return f == FlashEGA && g == GraphicsEGA }

// Art is a picture in the folder's art: an archive's block (TITLE 2), or
// with Kind ArtPNG the PNG File in the folder's art/ (POR's TITLE.PNG).
type Art struct {
	File  string
	Block int
	Kind  ArtKind `json:",omitempty"`
}

// ArtKind is what an Art's File names.
type ArtKind uint8

const (
	ArtDAX ArtKind = iota // a DAX archive, Block its block
	ArtPNG                // a PNG picture in the folder's art/
)

// Layer is a picture placed in 40x25 cells, optionally cycling frames.
type Layer struct {
	Art
	At         Cells
	Frames     []int `json:",omitempty"`
	FrameMs    int   `json:",omitempty"`
	Centred    bool  `json:",omitempty"` // the picture at its own size, centred in At (SotSB's Amiga credits card)
	MinParty   int   `json:",omitempty"` // shown only with at least this many party members (POR's finale: one picture per member)
	MsgDelayed bool  `json:",omitempty"` // FrameMs is as at the default message delay: the host's message delay slows or quickens it
}

// Cells is a box in 40x25 cells, Min and Max the corner cells (inclusive).
type Cells struct{ X1, Y1, X2, Y2 int }

// Credit is a credits line at its cell and colour, on page Page.
type Credit struct {
	X, Y   int
	Colour game.Ink
	Text   string
	Page   int `json:",omitempty"`
}

// cellPx is a 40x25 cell's size in page pixels (320x200).
const cellPx = 8

// The full-screen animations' names (game.Scene.Sequence, game.MenuItem.Sequence):
// the Start's and the End's leading StepAnim.
const (
	SeqIntro  = "intro"
	SeqEnding = "ending"
)

// ErrStepKind: a sequence step a title cannot show (a StepAnim after
// the first step; a leading one plays first, LeadingAnim).
var ErrStepKind = errors.New("step kind not in a title")

// ErrNotCarried: a Sequence field gobox does not carry yet.
var ErrNotCarried = errors.New("sequence field not carried yet")

// TitleIntro is the title of sequence s as played: step N of Pictures is
// s.Steps[N-1] (gobox reads the step from the script), -N credits page N-1.
// Without full (the host's Full intro option) only s.Short's steps play.
// Every edition's steps are in it: the front skips those of an art not
// shown (Plays), so the title follows F12 without being built again. A step whose
// Lightning is On in graphics mode g makes the intro flash. whole: the
// picture steps are whole-screen pages. A leading StepAnim is not a title
// step: it plays before the title (LeadingAnim).
func TitleIntro(s Sequence, full bool, g Graphics) (in *game.Intro, whole bool, err error) {
	in = &game.Intro{Full: true}
	pics, wholes := 0, 0
	for i, st := range s.Steps {
		n := i + 1
		if i == 0 && st.Kind == StepAnim {
			continue
		}
		if !full && s.Short != nil && !slices.Contains(s.Short, i) {
			continue
		}
		if st.HoldMs != 0 {
			return nil, false, fmt.Errorf("step %d HoldMs: %w", n, ErrNotCarried)
		}
		switch st.Kind {
		case StepPicture:
			in.Pictures = append(in.Pictures, n)
			pics++
			if st.Whole {
				wholes++
			}
		case StepText:
			if len(st.Layers) != 0 || len(st.Text) == 0 {
				return nil, false, fmt.Errorf("step %d: a text step has text and no layers: %w", n, ErrStepKind)
			}
			in.Pictures = append(in.Pictures, n)
		case StepCredits:
			in.Pictures = append(in.Pictures, -(st.Page + 1))
		default:
			return nil, false, fmt.Errorf("step %d: %w", n, ErrStepKind)
		}
		in.Lightning = in.Lightning || st.Lightning.On(g)
	}
	if wholes != 0 && wholes != pics {
		return nil, false, fmt.Errorf("Whole on some picture steps only: %w", ErrNotCarried)
	}
	return in, wholes != 0, nil
}

// Plays reports whether title step n (TitleIntro's numbering: > 0 a
// step, < 0 a credits page) plays with the Amiga art shown (amiga) or not:
// an EditionAmiga step only with it.
func (s *Sequence) Plays(n int, amiga bool) bool {
	st := s.Step(n)
	return n < 0 || st == nil || st.Edition != EditionAmiga || amiga
}

// LeadingAnim is s's leading StepAnim (nil: s is nil or starts with no animation).
func (s *Sequence) LeadingAnim() *SequenceStep {
	if s == nil || len(s.Steps) == 0 || s.Steps[0].Kind != StepAnim {
		return nil
	}
	return &s.Steps[0]
}

// Step is the step N of a game.Intro's Pictures (nil: none such).
func (s *Sequence) Step(n int) *SequenceStep {
	if s == nil || n < 1 || n > len(s.Steps) {
		return nil
	}
	return &s.Steps[n-1]
}

// pixRect is cell box c in page pixels, Max exclusive.
func pixRect(c Cells) image.Rectangle {
	return image.Rect(c.X1*cellPx, c.Y1*cellPx, (c.X2+1)*cellPx, (c.Y2+1)*cellPx)
}

// Placed is layer l's box in page pixels for a picture of w x h pixels:
// its cells, or with Centred the picture centred in them.
func (l Layer) Placed(w, h int) image.Rectangle {
	box := pixRect(l.At)
	if !l.Centred {
		return box
	}
	at := box.Min.Add(image.Pt((box.Dx()-w)/2, (box.Dy()-h)/2))
	return image.Rectangle{Min: at, Max: at.Add(image.Pt(w, h))}
}
