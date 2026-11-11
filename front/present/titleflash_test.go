package present

import (
	"testing"

	"gobox/ega"
	"gobox/front"
)

// TestTitleFlash: while the title's lightning flashes (IntroUI.Flash), the
// title picture's palette entry 1 shows the flash step's EGA colour, as
// c70cbee's title_screen set it (SetPaletteColor(c, 1)).
func TestTitleFlash(t *testing.T) {
	pic := front.Picture{Pix: []byte{1}, PixW: 1}
	for _, c := range []struct {
		flash []int
		want  int
	}{{nil, 1}, {[]int{15, 11}, 15}} {
		p := &Presenter{}
		p.intro.Flash = c.flash
		v := p.own(front.View{Intro: &front.Intro{}, Pictures: []front.Picture{pic}})
		if got := v.Pictures[0].Img.At(0, 0); got != ega.RGBA(c.want) {
			t.Errorf("flash %v: %v, want EGA %d", c.flash, got, c.want)
		}
	}
}
