package importer

import (
	"fmt"
	"image"
	"reflect"
	"slices"

	"gobox/adventure"
	"gobox/game"
)

// picSource is each ArtKind's game.PicSource.
var picSource = map[adventure.ArtKind]game.PicSource{adventure.ArtDAX: game.SourceDAX, adventure.ArtPNG: game.SourcePNG}

// TitleDiffs lists where sequence s differs from a game's title as it states
// it today (Scene today, Data d): picture step i has the layers of
// today.Steps[today.Intro.Pictures[i]] (with the credits page 0 gobox adds
// after two or more pictures, front/title.go), a credits step its page, Backdrop
// d's TitleBackdrop, Credits d's
// CreditLines. Pictures compare by File and Block (pixels
// unloaded, Clear -1); a Centred layer takes today's size. Songs are not compared. For the importers' tests.
func TitleDiffs(s adventure.Sequence, today game.Scene, d game.Data) (diffs []string) {
	layers := func(ls []adventure.Layer, today []game.Layer) (out []game.Layer) {
		for i, l := range ls {
			var size image.Point
			if i < len(today) {
				size = today[i].At.Size()
			}
			out = append(out, game.Layer{Picture: game.Picture{Source: picSource[l.Kind], File: l.File, Block: l.Block, Clear: -1}, At: l.Placed(size.X, size.Y), Frames: l.Frames, FrameMs: l.FrameMs})
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
		case adventure.StepPicture:
			if got, want := layers(st.Layers, today.Steps[n].Layers), bare(today.Steps[n].Layers); !reflect.DeepEqual(got, want) {
				diffs = append(diffs, fmt.Sprintf("step %d:\n got %+v\nwant %+v", i, got, want))
			}
		case adventure.StepCredits:
			if n != -(st.Page + 1) {
				diffs = append(diffs, fmt.Sprintf("step %d: credits page %d, today %d", i, st.Page, n))
			}
		default:
			diffs = append(diffs, fmt.Sprintf("step %d: kind %d", i, st.Kind))
		}
	}
	if got, want := layers(s.Backdrop, d.TitleBackdrop), bare(d.TitleBackdrop); len(s.Backdrop)+len(d.TitleBackdrop) != 0 && !reflect.DeepEqual(got, want) {
		diffs = append(diffs, fmt.Sprintf("backdrop:\n got %+v\nwant %+v", got, want))
	}
	var lines []game.Credit
	for _, c := range s.Credits {
		lines = append(lines, game.Credit{X: c.X, Y: c.Y, Colour: c.Colour, Text: c.Text, Page: c.Page})
	}
	if !reflect.DeepEqual(lines, d.CreditLines) {
		diffs = append(diffs, fmt.Sprintf("credits %+v\nwant %+v", lines, d.CreditLines))
	}
	return diffs
}
