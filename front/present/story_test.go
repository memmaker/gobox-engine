package present

import (
	"image"
	"image/color"
	"testing"

	"gobox/game"
)

func TestStoryPortrait(t *testing.T) {
	story := []game.Picture{{File: "HEAD1", Block: 4}, {File: "BODY1", Block: 7}}
	l := layPage(&game.Page{}, story, [16]color.RGBA{})
	if len(l.pictures) != 2 || l.pictures[0].At != image.Rect(24, 24, 112, 64) || l.pictures[1].At != image.Rect(24, 64, 112, 112) || l.pictures[1].File != "BODY1" {
		t.Fatalf("story portrait: %+v", l.pictures)
	}
}
