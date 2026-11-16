package importer

import (
	"gobox/adventure"
	"gobox/game"
)

// BR1EXE is where BR1's START.EXE keeps what the import reads.
var BR1EXE = BREXE{DS: 0xDAF, ItemWords: 0x1DB4, NumWords: 44, AmmoPrices: 0xC50, ShopLists: 0x18, ShipWeapons: 0x22DC, WeaponShort: 0x103F, ShipTypes: 0x1116}

// init registers Buck Rogers: Countdown to Doomsday with gobox/importer (scripts,
// art, maps, records, rule tables into adventures/br1/).
func init() {
	registerBR("br1", "Buck Rogers: Countdown to Doomsday", BR1EXE, Fixed(br1Start()))
}

// br1Start is BR1's title (pO025_0C36, rules/title.go; cells from pR267_159A,
// sizes from each TITLE block's VGA header): TITLE 1 at cell 0, 7; TITLE 2 at
// 2, 1 with TITLE 3 at 3, 21; the credits page (pO030_02E1,
// BR1Credits); TITLE 4 stays under the team menu.
func br1Start() *adventure.Sequence {
	pic := func(b, x, y, w, h int) adventure.Layer {
		return adventure.Layer{Art: adventure.Art{File: "TITLE", Block: b}, At: adventure.Cells{X1: x, Y1: y, X2: x + w - 1, Y2: y + h - 1}}
	}
	q := &adventure.Sequence{
		Steps: []adventure.SequenceStep{
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{pic(1, 0, 7, 40, 9)}},
			{Kind: adventure.StepPicture, Layers: []adventure.Layer{pic(2, 2, 1, 36, 18), pic(3, 3, 21, 34, 4)}},
			{Kind: adventure.StepCredits},
		},
		Backdrop: []adventure.Layer{pic(4, 6, 9, 29, 8)},
	}
	q.Credits = BR1Credits
	return q
}

// BR1Credits is the title's credits page (pO030_02E1): each line at its cell
// in its colour, as pR653_0424 prints them; the blank rows 2, 5 and 8
// closed up so the remaster's lines fit in the two rows above the frame.
var BR1Credits = []adventure.Credit{
	{X: 2, Y: 1, Colour: game.InkYellow, Text: "CREATED BY:SSI SPECIAL PROJECTS TEAM"},
	{X: 11, Y: 2, Colour: game.InkWhite, Text: "GAME DEVELOPMENT:"},
	{X: 14, Y: 3, Colour: game.InkLightGreen, Text: "BRET BERRY"},
	{X: 13, Y: 4, Colour: game.InkWhite, Text: "PROGRAMMING:"},
	{X: 4, Y: 5, Colour: game.InkLightGreen, Text: "MIKE MCNALLY      RUSS BROWN"},
	{X: 1, Y: 6, Colour: game.InkWhite, Text: "ENCOUNTER CODE:      MUSICAL SCORE:"},
	{X: 4, Y: 7, Colour: game.InkLightGreen, Text: "DAVID SHELLEY      TOM WAHL"},
	{X: 4, Y: 8, Colour: game.InkLightGreen, Text: "KEN HUMPHRIES"},
	{X: 22, Y: 8, Colour: game.InkWhite, Text: "PLAYTEST:"},
	{X: 4, Y: 9, Colour: game.InkLightGreen, Text: "RHONDA GILBERT    LARRY WEBBER"},
	{X: 4, Y: 10, Colour: game.InkLightGreen, Text: "KEN EKLUND        ERIK FLOM"},
	{X: 1, Y: 11, Colour: game.InkWhite, Text: "GRAPHIC ARTS:"},
	{X: 22, Y: 11, Colour: game.InkLightGreen, Text: "JOHN KIRK"},
	{X: 4, Y: 12, Colour: game.InkLightGreen, Text: "TOM WAHL          RICK WHITE"},
	{X: 4, Y: 13, Colour: game.InkLightGreen, Text: "LAURA BOWEN       DAVE LUCCA"},
	{X: 4, Y: 14, Colour: game.InkLightGreen, Text: "FRED BUTTS        CYRUS HARRIS"},
	{X: 4, Y: 15, Colour: game.InkLightGreen, Text: "MARK JOHNSON      GLEN CURETON"},
	{X: 4, Y: 16, Colour: game.InkLightGreen, Text: "CYRUS LUM         CHAD CASELLI"},
	{X: 4, Y: 17, Colour: game.InkLightGreen, Text: "SUSAN MANLEY      JEFF SHOTWELL"},
	{X: 4, Y: 18, Colour: game.InkLightGreen, Text: "MIKE PROVENZA     MIKE GILMARTIN"},
	{X: 4, Y: 19, Colour: game.InkLightGreen, Text: "MAURINE STARKEY   NATHANIEL TRUHER"},
}
