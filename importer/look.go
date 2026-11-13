package importer

import (
	"errors"
	"gobox/adventure"
	"gobox/asset"
	"image/color"
	"path/filepath"
)

// writeLook writes l as adventure.LookFile in the adventure's data folder
// data; a zero Look writes nothing (the front's own values stand).
func writeLook(data string, l adventure.Look) error {
	if l.ReadyMark == "" && l.Door == nil && l.AmigaGround == nil {
		return nil
	}
	return WriteJSON(filepath.Join(data, adventure.LookFile), l)
}

// groundColour is the colour the combat tiles in block of file paint the
// bare ground in: their most common one (black on DOS, a grey on the Amiga).
func groundColour(art *asset.Art, v asset.Version, file string, block int) (color.RGBA, error) {
	b := art.Block(v, file, block)
	if b == nil {
		return color.RGBA{}, errors.New("no " + file + " block for the combat ground")
	}
	im := b.Image(asset.NoMask)
	if im == nil {
		return color.RGBA{}, errors.New(file + " is no picture")
	}
	var n [256]int
	for _, p := range im.Data {
		n[p]++
	}
	best := 0
	for c := range n {
		if c != asset.Clear && n[c] > n[best] {
			best = c
		}
	}
	return asset.EGAIdentity.RGBA(uint8(best)), nil
}

// porLook is POR's: its doors on the maps are the Amiga's light brown
// (palette 0xA84), its Amiga combat tiles stand on the ground grey of SQRPACI.
func porLook(artDir string) (adventure.Look, error) {
	g, err := groundColour(asset.OpenArt(artDir), asset.Amiga, "SQRPACI", 1)
	if err != nil {
		return adventure.Look{}, err
	}
	door := asset.AmigaRGBA(asset.AmigaBase + 16*asset.AmigaPortrait + 6)
	return adventure.Look{Door: &door, AmigaGround: &g}, nil
}

// cotabLook is COTAB's: a readied item's name is marked "Yes ".
var cotabLook = adventure.Look{ReadyMark: "Yes "}
