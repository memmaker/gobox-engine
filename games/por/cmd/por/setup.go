package main

import (
	"path/filepath"

	"gobox/amigaimport"
	"gobox/app"
	"gobox/asset"
	"gobox/engine"
	"gobox/importer/games"
)

// setupData makes an installation whole once gameimport has copied the
// game's files into dst: the port's own assets (the text scripts) and the
// cover picture.
func setupData(dst string) error {
	if err := app.CopyFS(games.PorAssets, "por/scripts", filepath.Join(dst, engine.AssetsDir, "scripts")); err != nil {
		return err
	}
	if err := importScripts(dst); err != nil {
		return err
	}
	return app.WriteCover(filepath.Join(dst, engine.DOSDir), filepath.Join(dst, "COVER.PNG"))
}

// importScripts decompiles the game's event scripts into the installation
// and imports its art and data (pool.ImportAssets).
func importScripts(dst string) error {
	if err := games.ImportScripts(filepath.Join(dst, engine.DOSDir), filepath.Join(dst, engine.AssetsDir, "scripts")); err != nil {
		return err
	}
	return games.ImportAssets(dst)
}

// amigaGfx: the Amiga Pool of Radiance graphics archives, with the palette
// (0x54080 for PIC/FINAL, which the Amiga sets before them; 0x54020 for
// the ground tiles and the combat icons, which the combat screen (0x39F0E)
// and the icon editor (0x2EB02) set; 0x54040 for the portraits) each is
// drawn in, whether its frames pack two tiles side by side, and whether it
// is an animation block (PIC/FINAL frames XORed against the first).
// ponytail: 8X8D (the 3D-view symbols) stays DOS; the Amiga set is laid out
// for its own wall renderer and does not match WALLDEF's symbol numbers.
var amigaGfx = map[string]amigaimport.Gfx{
	"icon": {Slot: asset.AmigaTerrain}, "comspr": {Slot: asset.AmigaTerrain}, "cpic": {Slot: asset.AmigaTerrain},
	"bacpac": {Slot: asset.AmigaTerrain, Split: true}, "sqrpaci": {Slot: asset.AmigaTerrain, Split: true}, "wildcom": {Slot: asset.AmigaTerrain, Split: true},
	"dungcom": {Slot: asset.AmigaTerrain, Split: true}, "randcom": {Slot: asset.AmigaTerrain, Split: true},
	"head": {Slot: asset.AmigaPortrait}, "body": {Slot: asset.AmigaPortrait}, "chead": {Slot: asset.AmigaTerrain}, "cbody": {Slot: asset.AmigaTerrain},
	"pic": {Slot: asset.AmigaView, Sprite: true, XOR: true}, "final": {Slot: asset.AmigaView, Sprite: true, XOR: true}, "sprit": {Slot: asset.AmigaPortrait, Sprite: true},
}
