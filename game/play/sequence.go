package play

import (
	"errors"
	"fmt"
	"image"

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
// Lightning on any step makes the intro flash. The rest of s goes to the
// game's Data (TitleData).
func TitleScene(s importer.Sequence, cm Common, pic func(importer.Art) (game.Picture, error)) (Scene, error) {
	in := &game.Intro{Full: true}
	steps := map[int]TitleStep{}
	pics := 0
	whole := 0
	for i, st := range s.Steps {
		n := i + 1
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
		in.Lightning = in.Lightning || st.Lightning
	}
	if whole != 0 && whole != pics {
		return Scene{}, fmt.Errorf("Whole on some picture steps only: %w", ErrNotCarried)
	}
	return Scene{Common: cm, Kind: SceneTitle, Intro: in, Full: whole != 0, Steps: steps}, nil
}

// TitleData is s's credits lines and boxes as game.Data's CreditLines and
// CreditBoxes.
func TitleData(s importer.Sequence) (lines []game.Credit, boxes []image.Rectangle, err error) {
	switch {
	case s.Backdrop != nil:
		return nil, nil, fmt.Errorf("Backdrop: %w", ErrNotCarried)
	case s.Short != nil:
		return nil, nil, fmt.Errorf("Short: %w", ErrNotCarried)
	}
	for _, c := range s.Credits {
		lines = append(lines, game.Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
	}
	for _, b := range s.Boxes {
		boxes = append(boxes, image.Rect(b.X1, b.Y1, b.X2, b.Y2))
	}
	return lines, boxes, nil
}

// pixRect is cell box c in page pixels, Max exclusive.
func pixRect(c importer.Cells) image.Rectangle {
	return image.Rect(c.X1*cellPx, c.Y1*cellPx, (c.X2+1)*cellPx, (c.Y2+1)*cellPx)
}
