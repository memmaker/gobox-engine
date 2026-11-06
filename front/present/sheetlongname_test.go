package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestSheetLongSkillName: a rule-set name longer than the column (BR2's
// "Treat Light Wounds") is never written under its value.
func TestSheetLongSkillName(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	g := func(c byte) [8]byte {
		if c == '#' || c == '7' {
			return [8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		}
		return [8]byte{}
	}
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	name := strings.Repeat("#", 20)
	v := front.View{Party: []front.Member{{Name: "A", HP: 5, MaxHP: 5, Selected: true}}, SheetPage: &front.SheetPage{Rules: []front.Stat{{Name: name, Value: "7"}}, THAC0: 9}, Sheet: &front.Sheet{}}
	p := &Presenter{bars: gpu.NewImageFromImage(barArt())}
	p.drawSheet(dst, note, v)
	lit := 0
	for i := 0; i+3 < len(p.sheet.px.Pix); i += 4 {
		if p.sheet.px.Pix[i] != 0 || p.sheet.px.Pix[i+1] != 0 || p.sheet.px.Pix[i+2] != 0 {
			lit++
		}
	}
	// what a glyph covers depends on DrawText; the name and value apart cover 21 cells' worth
	want := 0
	solo := &Presenter{bars: gpu.NewImageFromImage(barArt())}
	v.SheetPage.Rules[0].Value = "x"
	solo.drawSheet(dst, note, v)
	for i := 0; i+3 < len(solo.sheet.px.Pix); i += 4 {
		if solo.sheet.px.Pix[i] != 0 || solo.sheet.px.Pix[i+1] != 0 || solo.sheet.px.Pix[i+2] != 0 {
			want++
		}
	}
	if lit-want != 64 { // an 8x8 glyph, whole
		t.Fatalf("the value lights %d pixels, not a whole glyph: drawn over the name", lit-want)
	}
}
