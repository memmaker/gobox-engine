package importer

import (
	"gobox/adventure"

	"gobox/game"
)

// titleAmigaCredits is the Amiga TITLE.DAX's own block, the port's credits
// card ("Amiga version by MicroMagic", 256x161): the last step on the Amiga,
// before the credits page.
const titleAmigaCredits = 5

// Credits are the title's credits page (0x36e): each line at its cell in
// its colour, as the original prints them.
var Credits = []adventure.Credit{
	{X: 7, Y: 0, Colour: game.InkMagenta, Text: "secret of the silver blades"},
	{X: 2, Y: 3, Colour: game.InkMagenta, Text: "game created by:    story and"},
	{X: 2, Y: 4, Colour: game.InkYellow, Text: "ssi special"},
	{X: 22, Y: 4, Colour: game.InkMagenta, Text: "development by:"},
	{X: 2, Y: 5, Colour: game.InkYellow, Text: "projects team       Dave Shelley"},
	{X: 2, Y: 7, Colour: game.InkMagenta, Text: "programming:        graphic arts:"},
	{X: 2, Y: 8, Colour: game.InkYellow, Text: "ken nicholson       mark johnson"},
	{X: 2, Y: 9, Colour: game.InkYellow, Text: "russ brown          laura bowen"},
	{X: 22, Y: 10, Colour: game.InkYellow, Text: "fred butts"},
	{X: 2, Y: 11, Colour: game.InkMagenta, Text: "Encounter Design:"},
	{X: 22, Y: 11, Colour: game.InkYellow, Text: "cyrus lum"},
	{X: 2, Y: 12, Colour: game.InkYellow, Text: "dave shelley        susan manley"},
	{X: 2, Y: 13, Colour: game.InkYellow, Text: "michael mancuso     mike provenza"},
	{X: 2, Y: 14, Colour: game.InkYellow, Text: "ken humphries       tom wahl"},
	{X: 2, Y: 15, Colour: game.InkYellow, Text: "graeme Bayless"},
	{X: 2, Y: 16, Colour: game.InkYellow, Text: "rick white"},
	{X: 22, Y: 16, Colour: game.InkMagenta, Text: "Musical Score:"},
	{X: 22, Y: 17, Colour: game.InkYellow, Text: "john halbleib"},
	{X: 2, Y: 19, Colour: game.InkMagenta, Text: "playtest:"},
	{X: 12, Y: 19, Colour: game.InkYellow, Text: "rick white,don mc clure,"},
	{X: 2, Y: 20, Colour: game.InkYellow, Text: "dave lucca,cliff mann,rick wilson,"},
	{X: 2, Y: 21, Colour: game.InkYellow, Text: "erik flom,james young,larry webber"},
}
