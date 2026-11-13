package present

import (
	"image"
	"image/color"
	"testing"

	"gobox/front"
	"gobox/game/play"
)

// The pool's portrait and icon on the creation sheet (a Name, no File) are
// content the screen draws, as the game's DAX pictures there are.
func TestContentPoolPictures(t *testing.T) {
	v := front.View{Pictures: []front.Picture{
		{Name: "head_000", Img: image.NewRGBA(image.Rect(0, 0, 88, 40)), At: image.Rect(224, 8, 312, 48)},
		{Name: "combat_head_000", Pix: make([]uint8, 576), PixW: 24, At: image.Rect(232, 104, 256, 128)},
		{Pix: make([]uint8, 576), PixW: 24, At: image.Rect(232, 104, 256, 128)}, // a frame: not content
	}}
	p, _ := draftOn(&play.Draft{}) // the page's default layout: the portrait at 28,1, the icon at 29,13
	if n := len(p.contentPictures(v)); n != 2 {
		t.Fatalf("%d content pictures, want the pool's 2", n)
	}
}

// A game naming no SelectInk still gets a visible focus frame.
func TestSelectOutlineInk(t *testing.T) {
	old := selectInk
	defer func() { selectInk = old }()
	selectInk = color.RGBA{}
	selectOutline(image.Rect(0, 0, 8, 8), 1, func(_ image.Rectangle, c color.RGBA) {
		if c.A == 0 {
			t.Fatal("the focus frame has no colour")
		}
	})
}
