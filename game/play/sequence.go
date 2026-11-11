package play

import (
	"errors"
	"fmt"
	"image"
	"reflect"
	"slices"

	"gobox/game"
	"gobox/importer"
)

// cellPx is a 40x25 cell's size in page pixels (320x200).
const cellPx = 8

// ErrStepKind: a sequence step a title Scene cannot show (StepAnim plays as
// SceneSequence, not in a title).
var ErrStepKind = errors.New("step kind not in a title scene")

// ErrNotCarried: a Sequence field gobox does not carry yet.
var ErrNotCarried = errors.New("sequence field not carried yet")

// TitleScene is the title Scene of sequence s's steps: step i (from 1) its
// layers, loaded by pic, and its tune; the credits steps show their page.
// Without full (the host's Full intro option) only s.Short's steps play.
// A step whose Lightning is On in graphics mode g makes the intro flash. The rest of s goes to the
// game's Data (TitleData).
func TitleScene(s importer.Sequence, full bool, g importer.Graphics, cm Common, pic func(importer.Art) (game.Picture, error)) (Scene, error) {
	in := &game.Intro{Full: true}
	steps := map[int]TitleStep{}
	pics := 0
	whole := 0
	for i, st := range s.Steps {
		n := i + 1
		if !full && s.Short != nil && !slices.Contains(s.Short, i) {
			continue
		}
		if st.HoldMs != 0 {
			return Scene{}, fmt.Errorf("step %d HoldMs: %w", n, ErrNotCarried)
		}
		switch st.Kind {
		case importer.StepPicture:
			ts := TitleStep{Song: st.Song}
			for _, l := range st.Layers {
				p, err := pic(l.Art)
				if err != nil {
					return Scene{}, fmt.Errorf("step %d: %w", n, err)
				}
				ts.Layers = append(ts.Layers, game.Layer{Picture: p, At: pixRect(l.At), Frames: l.Frames, FrameMs: l.FrameMs})
			}
			steps[n] = ts
			in.Pictures = append(in.Pictures, n)
			pics++
			if st.Whole {
				whole++
			}
		case importer.StepCredits:
			in.Pictures = append(in.Pictures, -(st.Page + 1))
		default:
			return Scene{}, fmt.Errorf("step %d: %w", n, ErrStepKind)
		}
		in.Lightning = in.Lightning || st.Lightning.On(g)
	}
	if whole != 0 && whole != pics {
		return Scene{}, fmt.Errorf("Whole on some picture steps only: %w", ErrNotCarried)
	}
	return Scene{Common: cm, Kind: SceneTitle, Intro: in, Full: whole != 0, Steps: steps}, nil
}

// TitleData is s's credits lines, boxes and backdrop as game.Data's
// CreditLines, CreditBoxes and TitleBackdrop, the backdrop's pictures loaded
// by pic.
func TitleData(s importer.Sequence, pic func(importer.Art) (game.Picture, error)) (lines []game.Credit, boxes []image.Rectangle, backdrop []game.Layer, err error) {
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
		backdrop = append(backdrop, game.Layer{Picture: p, At: pixRect(l.At), Frames: l.Frames, FrameMs: l.FrameMs})
	}
	return lines, boxes, backdrop, nil
}

// picSource is each importer.ArtKind's game.PicSource.
var picSource = map[importer.ArtKind]game.PicSource{importer.ArtDAX: game.SourceDAX, importer.ArtPNG: game.SourcePNG}

// pixRect is cell box c in page pixels, Max exclusive.
func pixRect(c importer.Cells) image.Rectangle {
	return image.Rect(c.X1*cellPx, c.Y1*cellPx, (c.X2+1)*cellPx, (c.Y2+1)*cellPx)
}

// TitleDiffs lists where sequence s differs from a game's title as it states
// it today (Scene today, Data d): picture step i has the layers of
// today.Steps[today.Intro.Pictures[i]] (with the credits page 0 gobox adds
// after two or more pictures, front/title.go), a credits step its page, Backdrop
// d's TitleBackdrop, Credits and Remaster d's
// CreditLines and RemasterCredit. Pictures compare by File and Block (pixels
// unloaded, Clear -1). Songs are not compared. For the importers' tests.
func TitleDiffs(s importer.Sequence, today Scene, d game.Data) (diffs []string) {
	layers := func(ls []importer.Layer) (out []game.Layer) {
		for _, l := range ls {
			out = append(out, game.Layer{Picture: game.Picture{Source: picSource[l.Kind], File: l.File, Block: l.Block, Clear: -1}, At: pixRect(l.At), Frames: l.Frames, FrameMs: l.FrameMs})
		}
		return out
	}
	bare := func(ls []game.Layer) []game.Layer {
		out := slices.Clone(ls)
		for i := range out {
			out[i].Picture = game.Picture{Source: out[i].Source, File: out[i].Picture.File, Block: out[i].Picture.Block, Clear: -1}
		}
		return out
	}
	if today.Intro == nil {
		return []string{"no intro today"}
	}
	pics := today.Intro.Pictures
	if len(pics) > 1 && len(d.CreditLines) != 0 && !slices.ContainsFunc(pics, func(n int) bool { return n < 0 }) {
		pics = append(slices.Clone(pics), -1) // front/title.go (NoCredits: no CreditLines): gobox's credits page 0 after the pictures
	}
	if len(s.Steps) != len(pics) {
		return []string{fmt.Sprintf("%d steps, today's %v", len(s.Steps), pics)}
	}
	for i, st := range s.Steps {
		n := pics[i]
		switch st.Kind {
		case importer.StepPicture:
			if got, want := layers(st.Layers), bare(today.Steps[n].Layers); !reflect.DeepEqual(got, want) {
				diffs = append(diffs, fmt.Sprintf("step %d:\n got %+v\nwant %+v", i, got, want))
			}
		case importer.StepCredits:
			if n != -(st.Page + 1) {
				diffs = append(diffs, fmt.Sprintf("step %d: credits page %d, today %d", i, st.Page, n))
			}
		default:
			diffs = append(diffs, fmt.Sprintf("step %d: kind %d", i, st.Kind))
		}
	}
	if got, want := layers(s.Backdrop), bare(d.TitleBackdrop); len(s.Backdrop)+len(d.TitleBackdrop) != 0 && !reflect.DeepEqual(got, want) {
		diffs = append(diffs, fmt.Sprintf("backdrop:\n got %+v\nwant %+v", got, want))
	}
	var lines []game.Credit
	for _, c := range s.Credits {
		lines = append(lines, game.Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
	}
	if !reflect.DeepEqual(lines, d.CreditLines) || s.Remaster != d.RemasterCredit {
		diffs = append(diffs, fmt.Sprintf("credits %q %+v\nwant %q %+v", s.Remaster, lines, d.RemasterCredit, d.CreditLines))
	}
	return diffs
}
