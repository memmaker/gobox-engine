package main

import (
	"path/filepath"

	"gobox/amigaimport"
	"gobox/app"
	"gobox/ega"
	"gobox/pool"
	"por"
	"por/importpor"
)

// setupData makes an installation whole once gameimport has copied the
// game's files into dst: the port's own assets (the text scripts) and the
// cover picture.
func setupData(dst string) error {
	if err := app.CopyFS(por.Assets, "scripts", filepath.Join(dst, pool.AssetsDir, "scripts")); err != nil {
		return err
	}
	if err := importScripts(dst); err != nil {
		return err
	}
	return app.WriteCover(filepath.Join(dst, pool.DOSDir), filepath.Join(dst, "COVER.PNG"))
}

// importScripts decompiles the game's event scripts into the installation
// and imports its art and data (pool.ImportAssets).
func importScripts(dst string) error {
	if err := importpor.ImportScripts(filepath.Join(dst, pool.DOSDir), filepath.Join(dst, pool.AssetsDir, "scripts")); err != nil {
		return err
	}
	return importpor.ImportAssets(dst)
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
	"icon": {Slot: ega.AmigaTerrain}, "comspr": {Slot: ega.AmigaTerrain}, "cpic": {Slot: ega.AmigaTerrain},
	"bacpac": {Slot: ega.AmigaTerrain, Split: true}, "sqrpaci": {Slot: ega.AmigaTerrain, Split: true}, "wildcom": {Slot: ega.AmigaTerrain, Split: true},
	"dungcom": {Slot: ega.AmigaTerrain, Split: true}, "randcom": {Slot: ega.AmigaTerrain, Split: true},
	"head": {Slot: ega.AmigaPortrait}, "body": {Slot: ega.AmigaPortrait}, "chead": {Slot: ega.AmigaTerrain}, "cbody": {Slot: ega.AmigaTerrain},
	"pic": {Slot: ega.AmigaView, Sprite: true, XOR: true}, "final": {Slot: ega.AmigaView, Sprite: true, XOR: true}, "sprit": {Slot: ega.AmigaPortrait, Sprite: true},
}
