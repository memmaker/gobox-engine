package adventure

import (
	"errors"
	"image"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
)

func TestTitleScene(t *testing.T) {
	seq := Sequence{Steps: []SequenceStep{
		{Kind: StepPicture, Song: "3", Lightning: FlashEGA, Whole: true, Layers: []Layer{{Art: Art{File: "TITLE", Block: 1}, At: Cells{X2: 39, Y2: 24}}}},
		{Kind: StepPicture, Whole: true, Layers: []Layer{{Art: Art{File: "TITLE", Block: 2}, At: Cells{X1: 1, Y1: 2, X2: 3, Y2: 4}, Frames: []int{0, 1}, FrameMs: 100}}},
		{Kind: StepCredits, Page: 0},
	}}
	pic := func(a Art) (game.Picture, error) { return game.Picture{File: a.File, Block: a.Block}, nil }
	got, err := TitleScene(seq, true, GraphicsEGA, game.Common{}, pic)
	if err != nil {
		t.Fatal(err)
	}
	want := game.Scene{Kind: game.SceneTitle, Full: true, Intro: &game.Intro{Full: true, Pictures: []int{1, 2, -1}, Lightning: true}, Steps: map[int]game.TitleStep{
		1: {Song: "3", Layers: []game.Layer{{Picture: game.Picture{File: "TITLE", Block: 1}, At: image.Rect(0, 0, 320, 200)}}},
		2: {Layers: []game.Layer{{Picture: game.Picture{File: "TITLE", Block: 2}, At: image.Rect(8, 16, 32, 40), Frames: []int{0, 1}, FrameMs: 100}}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	seq.Short = []int{0} // the short intro: the first step alone
	short, err := TitleScene(seq, false, GraphicsEGA, game.Common{}, pic)
	if err != nil || !slices.Equal(short.Intro.Pictures, []int{1}) || len(short.Steps) != 1 {
		t.Fatalf("short %+v, %v", short.Intro, err)
	}
	if _, err := TitleScene(Sequence{Steps: []SequenceStep{{Kind: StepCredits}, {Kind: StepAnim}}}, true, GraphicsEGA, game.Common{}, pic); !errors.Is(err, ErrStepKind) {
		t.Fatalf("anim step: %v", err)
	}
	lead := Sequence{Steps: append([]SequenceStep{{Kind: StepAnim, Anim: "INTRO", Song: "P1"}}, seq.Steps...)}
	led, err := TitleScene(lead, true, GraphicsEGA, game.Common{}, pic)
	if err != nil || !slices.Equal(led.Intro.Pictures, []int{2, 3, -1}) || !reflect.DeepEqual(led.Steps[2], want.Steps[1]) {
		t.Fatalf("leading anim: %+v, %v", led.Intro, err)
	}
	if q := LeadingAnim(lead, "intro"); q == nil || !reflect.DeepEqual(*q, game.Sequence{Name: "intro", Anim: "INTRO", Music: []string{"P1"}}) {
		t.Fatalf("LeadingAnim %+v", q)
	}
	if LeadingAnim(seq, "intro") != nil {
		t.Fatal("LeadingAnim without an anim")
	}
	if _, err := TitleScene(Sequence{Steps: []SequenceStep{{HoldMs: 1}}}, true, GraphicsEGA, game.Common{}, pic); !errors.Is(err, ErrNotCarried) {
		t.Errorf("hold: %v", err)
	}
	if _, err := TitleScene(Sequence{Steps: []SequenceStep{{Whole: true}, {}}}, true, GraphicsEGA, game.Common{}, pic); !errors.Is(err, ErrNotCarried) {
		t.Errorf("mixed whole: %v", err)
	}
	lines, boxes, bd, err := TitleData(Sequence{Credits: []Credit{{X: 2, Y: 3, Colour: 4, Text: "a", Page: 1}}, Boxes: []Cells{{X1: 1, Y1: 1, X2: 38, Y2: 3}},
		Backdrop: []Layer{{Art: Art{File: "TITLE", Block: 3}, At: Cells{X1: 1, Y1: 2, X2: 3, Y2: 4}}}}, pic)
	if err != nil || !reflect.DeepEqual(lines, []game.Credit{{X: 2, Y: 3, Colour: 4, Text: "a", Page: 1}}) || !reflect.DeepEqual(boxes, []image.Rectangle{{Min: image.Pt(1, 1), Max: image.Pt(38, 3)}}) ||
		len(bd) != 1 || bd[0].Block != 3 || bd[0].At != image.Rect(8, 16, 32, 40) {
		t.Errorf("TitleData %v %v %+v %v", lines, boxes, bd, err)
	}
}

// TestFlashEGAOnly: FlashEGA flashes in EGA graphics only.
func TestFlashEGAOnly(t *testing.T) {
	if !FlashEGA.On(GraphicsEGA) || FlashEGA.On(GraphicsOther) || FlashNone.On(GraphicsEGA) {
		t.Error("FlashEGA must flash in EGA only")
	}
}

// TestTitleSceneText: a text step is a step of its own with no picture, and a picture
// step carries the words said over it; the layers keep what the game filters them by.
func TestTitleSceneText(t *testing.T) {
	seq := Sequence{Steps: []SequenceStep{
		{Kind: StepText, Text: []string{"one"}},
		{Kind: StepPicture, Text: []string{"two"}, Layers: []Layer{{Art: Art{File: "F", Block: 1}, MinParty: 2, MsgDelayed: true}}},
	}}
	pic := func(a Art) (game.Picture, error) { return game.Picture{File: a.File, Block: a.Block}, nil }
	sc, err := TitleScene(seq, true, GraphicsEGA, game.Common{}, pic)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(sc.Intro.Pictures, []int{1, 2}) || !slices.Equal(sc.Steps[1].Text, []string{"one"}) || len(sc.Steps[1].Layers) != 0 {
		t.Fatalf("text step: %+v", sc)
	}
	if l := sc.Steps[2].Layers[0]; !slices.Equal(sc.Steps[2].Text, []string{"two"}) || l.MinParty != 2 || !l.MsgDelayed {
		t.Fatalf("picture step: %+v", sc.Steps[2])
	}
	if _, err := TitleScene(Sequence{Steps: []SequenceStep{{Kind: StepText}}}, true, GraphicsEGA, game.Common{}, pic); !errors.Is(err, ErrStepKind) {
		t.Errorf("empty text step: %v", err)
	}
}
