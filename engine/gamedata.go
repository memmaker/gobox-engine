package engine

import (
	"gobox/rules"
	"strconv"

	"gobox/asset"
	"gobox/game"
)

// overlandArt is the overland's and the title's pictures for game.Data.Art: "terrain/<id>"
// for each terrain tile id, "border/<place>" for the window border and "marker/<facing>" for the party marker.
func (e *Engine) overlandArt() map[string]game.Picture {
	border := e.borderTiles()
	var heads []byte
	for _, id := range e.partySide() {
		heads = append(heads, byte(e.fact(id, rules.PortraitFact{}).(rules.Portrait).Head))
	}
	k := artKey{e.sqrpaci, e.icons[9][0], DOSArt.Load(), border[0].Pix != nil, string(heads)}
	if e.ext.art != nil && e.ext.artKey == k {
		return e.ext.art
	}
	art := map[string]game.Picture{}
	for id := range 256 {
		if p := e.terrainPix(byte(id)); p.Pix != nil {
			art["terrain/"+strconv.Itoa(id)] = gamePicture(p)
		}
	}
	if im := e.icons[9][0]; im != nil {
		for f := 0; f < 8; f += 2 {
			art["marker/"+strconv.Itoa(f)] = gamePicture(e.picPix(im, 0, f > 3, -1))
		}
	}
	for i, p := range border { // the window border's pieces
		if p.Pix != nil {
			art["border/"+strconv.Itoa(i)] = gamePicture(p)
		}
	}
	for _, id := range e.partySide() { // the party's heads on the cards
		if n := e.headArt(id); n != "" {
			art[n] = gamePicture(e.charHead(id))
		}
	}
	e.ext.art, e.ext.artKey = art, k
	return art
}

// artKey is what overlandArt builds from: the tiles, the marker and the art set.
type artKey struct {
	tiles  [2]*asset.Image
	marker *asset.Image
	dos    bool
	border bool   // the border's tiles are loaded
	heads  string // the party's PortHead numbers
}

// Data is the game's data (GameDataFile) with the overland's art, as
// loaded now.
func (e *Engine) Data() game.Data {
	d := e.gameData
	d.Adventure = e.adv // the presenter's art (Picture File and Block, Theme.Border) and sound: Adventure is json:"-", so never in gameData
	d.Scans = e.journalPicture
	for i, t := range e.clueChapters() {
		r := e.ClueRead(i)
		d.Clues = append(d.Clues, game.Clue{Title: t, Read: game.Read{Title: r.Title, Texts: r.Texts, Page: r.Page}})
	}
	d.Art = e.overlandArt()
	d.CreditLines, d.CreditBoxes, d.TitleBackdrop = e.title.credits, e.title.boxes, e.title.backdrop // the adventure's Start (SetStart)
	d.Sequences = e.title.seqs
	d.Fonts = e.fontFaces()
	d.SoundTables = e.sounds // read by SoundData before the presenter starts
	return d
}
