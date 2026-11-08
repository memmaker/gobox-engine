package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

func TestStoryPortrait(t *testing.T) {
	s := game.State{Page: &game.Page{}, Story: []game.Picture{{File: "HEAD1", Block: 4}, {File: "BODY1", Block: 7}}}
	v := Page(front.View{}, s.Page, s.Story)
	if len(v.Pictures) != 2 || v.Pictures[0].At != image.Rect(24, 24, 112, 64) || v.Pictures[1].At != image.Rect(24, 64, 112, 112) || v.Pictures[1].File != "BODY1" {
		t.Fatalf("story portrait: %+v", v.Pictures)
	}
}
