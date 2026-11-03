package combat

import (
	"crypto/sha256"
	"fmt"
	"image/color"
	"testing"
)

func glyphT(c byte) [8]byte { return [8]byte{c, 0xFF, c ^ 0x55, 0x81, c, 0xAA, 0x0F, 0xF0} }

// TestDefaultThemeIsPor: the hashes were taken from the code before Theme existed.
func TestDefaultThemeIsPor(t *testing.T) {
	th := DefaultTheme
	items := []barItem{{'Y', "Yes", ActYes}, {'F', "Flee", ActFlee}, {'N', "No", ActNo}}
	h := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	for got, want := range map[string]string{
		h(barPixels(glyphT, items, 1, 400, 120, 2, nil, "", th).Pix):                   "6a11cd453e80347ed1474b8603d3f077be43bd507aec76de1742d42af85838a1",
		h(barPixels(glyphT, items, 1, 400, 120, 2, nil, "Casting X", th).Pix):          "d721542650fdefdee2cf607d19e3c57003ab957fdcd14a10bb24fc6ea149c521",
		h(woodTexture(64, th.Wood).Pix):                                                "00c672db0645934387f2483ebf8a3558f391253e6dc72dd193838f6e727efeb2",
		h(textImage(glyphT, "123", color.RGBA{255, 255, 82, 255}, th.TextOutline).Pix): "2bcba9d4952a141c22a4a9e0daa024afdce82d2ac75ed687ce166dbf65d96dec",
	} {
		if got != want {
			t.Errorf("got %s want %s", got, want)
		}
	}
}
