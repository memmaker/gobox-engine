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

// ErrNotCarried: a Sequence field the title Scene does not carry yet.
var ErrNotCarried = errors.New("sequence field not carried by the title scene")

// TitleScene is the title Scene of sequence s: step i (from 1) its layers,
// loaded by pic, and its tune; the credits steps show their page. Lightning
// on any step makes the intro flash.
func TitleScene(s importer.Sequence, cm Common, pic func(importer.Art) (game.Picture, error)) (Scene, error) {
	switch {
	case len(s.Credits) > 0:
		return Scene{}, fmt.Errorf("Credits: %w", ErrNotCarried)
	case len(s.Boxes) > 0:
		return Scene{}, fmt.Errorf("Boxes: %w", ErrNotCarried)
	case s.Backdrop != nil:
		return Scene{}, fmt.Errorf("Backdrop: %w", ErrNotCarried)
	case s.Short != nil:
		return Scene{}, fmt.Errorf("Short: %w", ErrNotCarried)
	}
	in := &game.Intro{Full: true}
	steps := map[int]TitleStep{}
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
				at := image.Rect(l.At.X1*cellPx, l.At.Y1*cellPx, (l.At.X2+1)*cellPx, (l.At.Y2+1)*cellPx)
				ts.Layers = append(ts.Layers, game.Layer{Picture: p, At: at, Frames: l.Frames, FrameMs: l.FrameMs})
			}
			steps[n] = ts
			in.Pictures = append(in.Pictures, n)
		case importer.StepCredits:
			in.Pictures = append(in.Pictures, -(st.Page + 1))
		default:
			return Scene{}, fmt.Errorf("step %d: %w", n, ErrStepKind)
		}
		in.Lightning = in.Lightning || st.Lightning
	}
	return Scene{Common: cm, Kind: SceneTitle, Intro: in, Full: true, Steps: steps}, nil
}
