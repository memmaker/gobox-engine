package play

import (
	"errors"
	"image"
	"reflect"
	"testing"

	"gobox/game"
	"gobox/importer"
)

func TestTitleScene(t *testing.T) {
	seq := importer.Sequence{Steps: []importer.SequenceStep{
		{Kind: importer.StepPicture, Song: "3", Lightning: importer.FlashEGA, Whole: true, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 1}, At: importer.Cells{X2: 39, Y2: 24}}}},
		{Kind: importer.StepPicture, Whole: true, Layers: []importer.Layer{{Art: importer.Art{File: "TITLE", Block: 2}, At: importer.Cells{X1: 1, Y1: 2, X2: 3, Y2: 4}, Frames: []int{0, 1}, FrameMs: 100}}},
		{Kind: importer.StepCredits, Page: 0},
	}}
	pic := func(a importer.Art) (game.Picture, error) { return game.Picture{File: a.File, Block: a.Block}, nil }
	got, err := TitleScene(seq, importer.GraphicsEGA, Common{}, pic)
	if err != nil {
		t.Fatal(err)
	}
	want := Scene{Kind: SceneTitle, Full: true, Intro: &game.Intro{Full: true, Pictures: []int{1, 2, -1}, Lightning: true}, Steps: map[int]TitleStep{
		1: {Song: "3", Layers: []game.Layer{{Picture: game.Picture{File: "TITLE", Block: 1}, At: image.Rect(0, 0, 320, 200)}}},
		2: {Layers: []game.Layer{{Picture: game.Picture{File: "TITLE", Block: 2}, At: image.Rect(8, 16, 32, 40), Frames: []int{0, 1}, FrameMs: 100}}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	if _, err := TitleScene(importer.Sequence{Steps: []importer.SequenceStep{{Kind: importer.StepAnim}}}, importer.GraphicsEGA, Common{}, pic); !errors.Is(err, ErrStepKind) {
		t.Fatalf("anim step: %v", err)
	}
	if _, err := TitleScene(importer.Sequence{Steps: []importer.SequenceStep{{HoldMs: 1}}}, importer.GraphicsEGA, Common{}, pic); !errors.Is(err, ErrNotCarried) {
		t.Errorf("hold: %v", err)
	}
	if _, err := TitleScene(importer.Sequence{Steps: []importer.SequenceStep{{Whole: true}, {}}}, importer.GraphicsEGA, Common{}, pic); !errors.Is(err, ErrNotCarried) {
		t.Errorf("mixed whole: %v", err)
	}
	for name, bad := range map[string]importer.Sequence{"backdrop": {Backdrop: []importer.Layer{{}}}, "short": {Short: []int{}}} {
		if _, _, err := TitleData(bad); !errors.Is(err, ErrNotCarried) {
			t.Errorf("%s: %v", name, err)
		}
	}
	lines, boxes, err := TitleData(importer.Sequence{Credits: []importer.Credit{{X: 2, Y: 3, Colour: 4, Text: "a", Page: 1}}, Boxes: []importer.Cells{{X1: 1, Y1: 1, X2: 38, Y2: 3}}})
	if err != nil || !reflect.DeepEqual(lines, []game.Credit{{X: 2, Y: 3, Colour: 4, Text: "a", Page: 1}}) || !reflect.DeepEqual(boxes, []image.Rectangle{{Min: image.Pt(1, 1), Max: image.Pt(38, 3)}}) {
		t.Errorf("TitleData %v %v %v", lines, boxes, err)
	}
}

// TestFlashEGAOnly: FlashEGA flashes in EGA graphics only.
func TestFlashEGAOnly(t *testing.T) {
	if !importer.FlashEGA.On(importer.GraphicsEGA) || importer.FlashEGA.On(importer.GraphicsOther) || importer.FlashNone.On(importer.GraphicsEGA) {
		t.Error("FlashEGA must flash in EGA only")
	}
}
