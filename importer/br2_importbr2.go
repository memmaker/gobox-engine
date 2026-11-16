package importer

import (
	"os"
	"path/filepath"

	"gobox/adventure"
	"gobox/game"
	"gobox/rules/xxvc"
)

// init registers Buck Rogers: Matrix Cubed with gobox/importer (scripts, art,
// maps, records, rule tables into adventures/br2/).
func init() {
	registerBR("br2", "Buck Rogers: Matrix Cubed", BR2EXE, br2Start)
}

// BR2EXE is where BR2's START.EXE keeps what the import reads.
var BR2EXE = BREXE{DS: 0xE87, ItemWords: 0x1EEC, NumWords: 45, AmmoPrices: 0xD86, ShopLists: 0x18, ShipWeapons: 0x242A, WeaponShort: 0x1177, ShipTypes: 0x124E, Explosion: xxvc.ExplosionFixed,
	Space: 0x25E8, MapSteps: 0x0B46} // buckrogers2-go rules/restspace_data.go spaceAt: Rows 0x25E8 (orbitPositions[0]), DX 0x0B46 (mapStepX); the orbit tables after the rows as in BR1 (PlanetBase 0x2696 = Rows+58*3)

// creditAt is pO025_04F1's credits page (English v1.0 O030): each
// displayString's cell, colour and its string's GAME.OVR offset.
var creditAt = []struct{ X, Y, Colour, At int }{
	{2, 1, 14, 393}, {2, 3, 15, 430}, {21, 3, 15, 448}, {2, 4, 10, 461}, {21, 4, 10, 476},
	{21, 5, 10, 487}, {2, 6, 15, 499}, {2, 7, 10, 515}, {20, 7, 15, 528}, {2, 8, 10, 542},
	{20, 8, 10, 553}, {2, 9, 10, 565}, {20, 9, 10, 573}, {2, 10, 10, 584}, {20, 10, 10, 595},
	{2, 11, 10, 611}, {20, 11, 10, 620}, {2, 12, 10, 634}, {20, 12, 10, 648}, {2, 13, 10, 661},
	{20, 13, 10, 675}, {2, 14, 10, 685}, {20, 14, 10, 701}, {2, 16, 15, 714}, {9, 16, 10, 721},
	{2, 18, 15, 748}, {13, 18, 10, 759}, {2, 19, 10, 786}, {2, 20, 10, 816}, {2, 21, 10, 852},
	{2, 22, 10, 889}, {2, 23, 10, 918},
}

// br2Start is BR2's title (pO025_0C36, rules/title.go; cells from draw_picture,
// sizes from each TITLE block's VGA header): TITLE 1 at cell 0, 7; TITLE 2 at
// 2, 1 with TITLE 3 at 3, 21; the credits page, its text GAME.OVR's
// (creditAt); TITLE 4 at 0, 0 stays under the team menu.
func br2Start(dos string) (*adventure.Sequence, error) {
	ovr, err := os.ReadFile(filepath.Join(dos, "GAME.OVR"))
	if err != nil {
		return nil, err
	}
	pic := func(b, x, y, w, h int) adventure.Layer {
		return adventure.Layer{Art: adventure.Art{File: "TITLE", Block: b}, At: adventure.Cells{X1: x, Y1: y, X2: x + w - 1, Y2: y + h - 1}}
	}
	q := &adventure.Sequence{
		Steps: []adventure.SequenceStep{
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{pic(1, 0, 7, 40, 9)}},
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{pic(2, 2, 1, 36, 18), pic(3, 3, 21, 34, 4)}},
			{Kind: adventure.StepCredits},
		},
		Backdrop: []adventure.Layer{pic(4, 0, 0, 40, 23)},
	}
	for _, c := range creditAt {
		q.Credits = append(q.Credits, adventure.Credit{X: c.X, Y: c.Y, Colour: game.EGAInk(c.Colour), Text: pascal(ovr, c.At)})
	}
	return q, nil
}

// pascal is the Pascal string at offset o (its length byte) in b, "" past its end.
func pascal(b []byte, o int) string {
	if o < 0 || o >= len(b) || o+1+int(b[o]) > len(b) {
		return ""
	}
	return string(b[o+1 : o+1+int(b[o])])
}
