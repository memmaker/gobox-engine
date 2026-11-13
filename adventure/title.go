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
	EditionAmiga                // shown only while the Amiga art is up
)

// Shown is s as played with the Amiga art up (amiga) or not: the steps of
// another edition dropped, Short's indexes moved with them, every Edition
// EditionAny. play.TitleScene takes only a sequence passed through Shown.
func (s Sequence) Shown(amiga bool) Sequence {
	out := s
	out.Steps, out.Short = nil, nil
	at := map[int]int{} // old step index -> new
	for i, st := range s.Steps {
		if st.Edition == EditionAmiga && !amiga {
			continue
		}
		at[i] = len(out.Steps)
		st.Edition = EditionAny
		out.Steps = append(out.Steps, st)
	}
	for _, i := range s.Short {
		if n, ok := at[i]; ok {
			out.Short = append(out.Short, n)
		}
	}
	if s.Short != nil && out.Short == nil {
		out.Short = []int{}
	}
	return out
}

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

// Credit is a credits line at its cell and colour (an EGA index), on page Page.
type Credit struct {
	X, Y, Colour int
	Text         string
	Page         int `json:",omitempty"`
}

// cellPx is a 40x25 cell's size in page pixels (320x200).
const cellPx = 8

// ErrStepKind: a sequence step a title Scene cannot show (a StepAnim after
// the first step; a leading one plays first as SceneSequence, LeadingAnim).
var ErrStepKind = errors.New("step kind not in a title scene")

// ErrNotCarried: a Sequence field gobox does not carry yet.
var ErrNotCarried = errors.New("sequence field not carried yet")

// TitleScene is the title Scene of sequence s's steps: step i (from 1) its
// layers, loaded by pic, and its tune; the credits steps show their page.
// Without full (the host's Full intro option) only s.Short's steps play.
// A step whose Lightning is On in graphics mode g makes the intro flash. The rest of s goes to the
// game's Data (TitleData). A leading StepAnim is not a title step (step 1
// stays its number): it plays before the title as LeadingAnim's sequence.
func TitleScene(s Sequence, full bool, g Graphics, cm game.Common, pic func(Art) (game.Picture, error)) (game.Scene, error) {
	in := &game.Intro{Full: true}
	steps := map[int]game.TitleStep{}
	pics := 0
	whole := 0
	for i, st := range s.Steps {
		n := i + 1
		if i == 0 && st.Kind == StepAnim {
			continue
		}
		if !full && s.Short != nil && !slices.Contains(s.Short, i) {
			continue
		}
		if st.Edition != EditionAny {
			return game.Scene{}, fmt.Errorf("step %d Edition %d (not passed through Sequence.Shown): %w", n, st.Edition, ErrNotCarried)
		}
		if st.HoldMs != 0 {
			return game.Scene{}, fmt.Errorf("step %d HoldMs: %w", n, ErrNotCarried)
		}
		switch st.Kind {
		case StepPicture:
			ts := game.TitleStep{Song: st.Song, Text: st.Text, Credits: st.Credits}
			for _, l := range st.Layers {
				p, err := pic(l.Art)
				if err != nil {
					return game.Scene{}, fmt.Errorf("step %d: %w", n, err)
				}
				ts.Layers = append(ts.Layers, game.Layer{Picture: p, At: l.Placed(p.W, p.H), Frames: l.Frames, FrameMs: l.FrameMs, MinParty: l.MinParty, MsgDelayed: l.MsgDelayed})
			}
			steps[n] = ts
			in.Pictures = append(in.Pictures, n)
			pics++
			if st.Whole {
				whole++
			}
		case StepText:
			if len(st.Layers) != 0 || len(st.Text) == 0 {
				return game.Scene{}, fmt.Errorf("step %d: a text step has text and no layers: %w", n, ErrStepKind)
			}
			steps[n] = game.TitleStep{Song: st.Song, Text: st.Text}
			in.Pictures = append(in.Pictures, n)
		case StepCredits:
			in.Pictures = append(in.Pictures, -(st.Page + 1))
		default:
			return game.Scene{}, fmt.Errorf("step %d: %w", n, ErrStepKind)
		}
		in.Lightning = in.Lightning || st.Lightning.On(g)
	}
	if whole != 0 && whole != pics {
		return game.Scene{}, fmt.Errorf("Whole on some picture steps only: %w", ErrNotCarried)
	}
	return game.Scene{Common: cm, Kind: game.SceneTitle, Intro: in, Full: whole != 0, Steps: steps}, nil
}

// LeadingAnim is s's leading StepAnim as the game.Sequence name plays
// (Anim, its Song as the Music), for game.Data.Sequences; nil: s starts with
// no animation.
func LeadingAnim(s Sequence, name string) *game.Sequence {
	if len(s.Steps) == 0 || s.Steps[0].Kind != StepAnim {
		return nil
	}
	st := s.Steps[0]
	q := &game.Sequence{Name: name, Anim: st.Anim}
	if st.Song != "" {
		q.Music = []string{st.Song}
	}
	return q
}

// TitleData is s's credits lines, boxes and backdrop as game.Data's
// CreditLines, CreditBoxes and TitleBackdrop, the backdrop's pictures loaded
// by pic.
func TitleData(s Sequence, pic func(Art) (game.Picture, error)) (lines []game.Credit, boxes []image.Rectangle, backdrop []game.Layer, err error) {
	for _, c := range s.Credits {
		lines = append(lines, game.Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
	}
	for _, b := range s.Boxes {
		boxes = append(boxes, image.Rect(b.X1, b.Y1, b.X2, b.Y2))
	}
	for _, l := range s.Backdrop {
		p, err := pic(l.Art)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("backdrop: %w", err)
		}
		backdrop = append(backdrop, game.Layer{Picture: p, At: l.Placed(p.W, p.H), Frames: l.Frames, FrameMs: l.FrameMs})
	}
	return lines, boxes, backdrop, nil
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
