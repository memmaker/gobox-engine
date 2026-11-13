package combat

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func glyphT(c byte) [8]byte { return [8]byte{c, 0xFF, c ^ 0x55, 0x81, c, 0xAA, 0x0F, 0xF0} }

// TestDefaultThemeIsPor: the hashes were taken from the code before Theme existed (embedded icons included).
func TestDefaultThemeIsPor(t *testing.T) {
	th := CombatDefaultTheme
	items := []barItem{{'F', "Attack", CombatActAttack, 0, 0, ""}, {'C', "Cast", CombatActCast, 0, 0, ""}, {'E', "Cancel", CombatActClose, 0, 0, ""}, {'W', "Wait", CombatActWait, 0, 0, ""}, {'Y', "Yes", CombatActYes, 0, 0, ""}}
	h := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	for got, want := range map[string]string{
		h(barPixels(glyphT, items, 1, 600, 120, 2, nil, "", th).Pix):          "9267a69e862b55ed55abf02246ce4f6ac43e8c42f39c4a02b5e43357f448e022",
		h(barPixels(glyphT, items, 1, 600, 120, 2, nil, "Casting X", th).Pix): "d36702d7d2ff8e57f169406d7d91a6a8d0735235f51c2090e96aedbacc0898af",
		h(woodTexture(64, th.Wood).Pix):                                       "00c672db0645934387f2483ebf8a3558f391253e6dc72dd193838f6e727efeb2",
		h(textImage(glyphT, "123", th.Hurt, th.TextOutline).Pix):              "d8d0cb085b1f60c1a1e845f635e8b89558a5a3a2d9d7a68b3240f726a505c390",
	} {
		if got != want {
			t.Errorf("got %s want %s", got, want)
		}
	}
}
