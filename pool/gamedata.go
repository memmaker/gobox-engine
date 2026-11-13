package pool

import (
	"encoding/json"
	"fmt"
	"gobox/core"
	"os"
	"path/filepath"
	"strconv"

	"gobox/asset"
	"gobox/game"
)

// GameDataFile is the game's declarative data (game.Data without what the
// engine builds at run time) in the adventure's data folder (DataDir), written by the
// game's import: its id, party menu, verbs, Quick Start party, font, border,
// rests, spell tips and the rest of game.Data's static fields.
const GameDataFile = "gamedata.json"

// WriteGameData writes d (its static fields) as GameDataFile into the data folder dir.
func WriteGameData(dir string, d game.Data) error {
	b, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, GameDataFile), b, 0o644)
}

// ReadGameData reads the game's data from the adventure folder adv, as its
// import wrote it. An Amiga music set (AmigaMusic) whose files the folder's sound/
// lacks is dropped; with the files, the AdLib music gives way to it.
func ReadGameData(adv string) (game.Data, error) {
	b, err := os.ReadFile(filepath.Join(adv, DataDir, GameDataFile))
	if err != nil {
		return game.Data{}, fmt.Errorf("%w (run the game's import)", err)
	}
	var d game.Data
	if err := json.Unmarshal(b, &d); err != nil {
		return game.Data{}, fmt.Errorf("%s: %w", GameDataFile, err)
	}
	if am := d.AmigaMusic; am != nil {
		have := true
		for _, f := range []string{am.Songs, am.Effects} {
			if f == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(adv, SoundDir, f)); err != nil {
				have = false
			}
		}
		if have {
			d.AdLib = nil
		} else {
			d.AmigaMusic = nil
		}
	}
	return d, nil
}

// overlandArt is the overland's and the title's pictures for game.Data.Art: "terrain/<id>"
// for each terrain tile id, "border/<place>" for the window border and "marker/<facing>" for the party marker, "title/logo" for the party menu.
func (e *Engine) overlandArt() map[string]game.Picture {
	logo := e.titleLogo()
	border := e.borderTiles()
	var heads []byte
	for _, id := range e.partySide() {
		heads = append(heads, byte(e.fact(id, core.PortraitFact{}).(core.Portrait).Head))
	}
	k := artKey{e.sqrpaci, e.icons[9][0], DOSArt.Load(), logo, border[0].Pix != nil, string(heads)}
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
	if logo != nil { // the title under the party menu
		art["title/logo"] = gamePicture(*logo)
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
	logo   *Picture
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
		d.Clues = append(d.Clues, game.Clue{Title: t, Read: game.Read{Title: r.Title, Lines: r.Lines, Page: r.Page}})
	}
	d.Art = e.overlandArt()
	d.CreditLines, d.CreditBoxes, d.TitleBackdrop = e.title.credits, e.title.boxes, e.title.backdrop // the adventure's Start (SetStart)
	d.Sequences = e.title.seqs
	d.Fonts = e.fontFaces()
	d.SoundTables = e.sounds // read by SoundData before the presenter starts
	return d
}
