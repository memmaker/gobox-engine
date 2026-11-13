package importsotsb

import (
	"image"

	"gobox/game"
)

// titleAmigaCredits is the Amiga TITLE.DAX's own block, the port's credits
// card ("Amiga version by MicroMagic", 256x161): the last step on the Amiga,
// before the credits page.
const titleAmigaCredits = 5

// titleLayers are each step's TITLE pictures and where they go on the screen.
// titleBanner is the step kept as the menu's backdrop (Data.TitleBackdrop).
const titleBanner = 4

var titleLayers = map[int][]game.Layer{
	1: {{Picture: game.Picture{File: "TITLE", Block: 1}, At: image.Rect(0, 0, 320, 200)}},
	2: {{Picture: game.Picture{File: "TITLE", Block: 2}, At: image.Rect(0, 0, 320, 200)}},
	3: {{Picture: game.Picture{File: "TITLE", Block: 2}, At: image.Rect(0, 0, 320, 200)},
		{Picture: game.Picture{File: "TITLE", Block: 3}, At: image.Rect(16, 40, 16+288, 40+88)}},
	4: {{Picture: game.Picture{File: "TITLE", Block: 2}, At: image.Rect(0, 0, 320, 200)},
		{Picture: game.Picture{File: "TITLE", Block: 4}, At: image.Rect(40, 40, 40+248, 40+88)}},
	titleAmigaCredits: {{Picture: game.Picture{File: "TITLE", Block: titleAmigaCredits}, At: image.Rect(32, 19, 32+256, 19+161)}}, // centred
}

// Credits are the title's credits page (0x36e): each line at its cell in
// its colour, as the original prints them.
var Credits = []game.Credit{
	{X: 7, Y: 0, Colour: 5, Text: "secret of the silver blades"},
	{X: 2, Y: 3, Colour: 5, Text: "game created by:    story and"},
	{X: 2, Y: 4, Colour: 14, Text: "ssi special"},
	{X: 22, Y: 4, Colour: 5, Text: "development by:"},
	{X: 2, Y: 5, Colour: 14, Text: "projects team       Dave Shelley"},
	{X: 2, Y: 7, Colour: 5, Text: "programming:        graphic arts:"},
	{X: 2, Y: 8, Colour: 14, Text: "ken nicholson       mark johnson"},
	{X: 2, Y: 9, Colour: 14, Text: "russ brown          laura bowen"},
	{X: 22, Y: 10, Colour: 14, Text: "fred butts"},
	{X: 2, Y: 11, Colour: 5, Text: "Encounter Design:"},
	{X: 22, Y: 11, Colour: 14, Text: "cyrus lum"},
	{X: 2, Y: 12, Colour: 14, Text: "dave shelley        susan manley"},
	{X: 2, Y: 13, Colour: 14, Text: "michael mancuso     mike provenza"},
	{X: 2, Y: 14, Colour: 14, Text: "ken humphries       tom wahl"},
	{X: 2, Y: 15, Colour: 14, Text: "graeme Bayless"},
	{X: 2, Y: 16, Colour: 14, Text: "rick white"},
	{X: 22, Y: 16, Colour: 5, Text: "Musical Score:"},
	{X: 22, Y: 17, Colour: 14, Text: "john halbleib"},
	{X: 2, Y: 19, Colour: 5, Text: "playtest:"},
	{X: 12, Y: 19, Colour: 14, Text: "rick white,don mc clure,"},
	{X: 2, Y: 20, Colour: 14, Text: "dave lucca,cliff mann,rick wilson,"},
	{X: 2, Y: 21, Colour: 14, Text: "erik flom,james young,larry webber"},
}
