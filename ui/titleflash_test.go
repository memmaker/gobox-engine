package ui

import (
	"image"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// TestTitleFlash: while the title's lightning flashes (IntroUI.Flash), the
// title picture's palette entry 1 shows the flash step's EGA colour, as
// c70cbee's title_screen set it (SetPaletteColor(c, 1)).
func TestTitleFlash(t *testing.T) {
	pic := game.Drawn{Picture: game.Picture{W: 1, H: 1, Pixels: []byte{1}, Clear: -1}, At: image.Rect(0, 0, 1, 1)}
	for _, c := range []struct {
		flash []int
		want  int
	}{{nil, 1}, {[]int{15, 11}, 15}} {
		p, _ := presenterOn(game.Scene{Kind: game.SceneTitle, Intro: &game.Intro{}, Common: game.Common{Layers: []game.Drawn{pic}}})
		p.intro.Flash = c.flash
		if got := p.pictures()[0].Img.At(0, 0); got != asset.EGARGBA(c.want) {
			t.Errorf("flash %v: %v, want EGA %d", c.flash, got, c.want)
		}
	}
}
