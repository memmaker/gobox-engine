package adventure

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
)

func TestTitleIntro(t *testing.T) {
	seq := Sequence{Steps: []SequenceStep{
		{Kind: StepPicture, Song: "3", Lightning: FlashEGA, Whole: true, Layers: []Layer{{Art: Art{File: "TITLE", Block: 1}, At: Cells{X2: 39, Y2: 24}}}},
		{Kind: StepPicture, Whole: true, Edition: EditionAmiga},
		{Kind: StepCredits, Page: 0},
	}}
	in, whole, err := TitleIntro(seq, true, true, GraphicsEGA)
	if err != nil || !whole || !reflect.DeepEqual(*in, game.Intro{Full: true, Pictures: []int{1, 2, -1}, Lightning: true}) {
		t.Fatalf("got %+v %v %v", in, whole, err)
	}
	if dos, _, _ := TitleIntro(seq, true, false, GraphicsOther); !slices.Equal(dos.Pictures, []int{1, -1}) || dos.Lightning {
		t.Fatalf("DOS art: %+v", dos)
	}
	seq.Short = []int{0} // the short intro: the first step alone
	if short, _, err := TitleIntro(seq, false, true, GraphicsEGA); err != nil || !slices.Equal(short.Pictures, []int{1}) {
		t.Fatalf("short %+v, %v", short, err)
	}
	if _, _, err := TitleIntro(Sequence{Steps: []SequenceStep{{Kind: StepCredits}, {Kind: StepAnim}}}, true, false, GraphicsEGA); !errors.Is(err, ErrStepKind) {
		t.Fatalf("anim step: %v", err)
	}
	lead := Sequence{Steps: append([]SequenceStep{{Kind: StepAnim, Anim: "INTRO", Song: "P1"}}, seq.Steps...)}
	if led, _, err := TitleIntro(lead, true, true, GraphicsEGA); err != nil || !slices.Equal(led.Pictures, []int{2, 3, -1}) || lead.Step(2).Song != "3" {
		t.Fatalf("leading anim: %+v, %v", led, err)
	}
	if q := lead.LeadingAnim(); q == nil || q.Anim != "INTRO" {
		t.Fatalf("LeadingAnim %+v", q)
	}
	if seq.LeadingAnim() != nil || (*Sequence)(nil).LeadingAnim() != nil || seq.Step(4) != nil {
		t.Fatal("LeadingAnim without an anim, Step out of range")
	}
	if _, _, err := TitleIntro(Sequence{Steps: []SequenceStep{{HoldMs: 1}}}, true, false, GraphicsEGA); !errors.Is(err, ErrNotCarried) {
		t.Errorf("hold: %v", err)
	}
	if _, _, err := TitleIntro(Sequence{Steps: []SequenceStep{{Whole: true}, {}}}, true, false, GraphicsEGA); !errors.Is(err, ErrNotCarried) {
		t.Errorf("mixed whole: %v", err)
	}
	if _, _, err := TitleIntro(Sequence{Steps: []SequenceStep{{Kind: StepText}}}, true, false, GraphicsEGA); !errors.Is(err, ErrStepKind) {
		t.Errorf("empty text step: %v", err)
	}
}

// TestFlashEGAOnly: FlashEGA flashes in EGA graphics only.
func TestFlashEGAOnly(t *testing.T) {
	if !FlashEGA.On(GraphicsEGA) || FlashEGA.On(GraphicsOther) || FlashNone.On(GraphicsEGA) {
		t.Error("FlashEGA must flash in EGA only")
	}
}
