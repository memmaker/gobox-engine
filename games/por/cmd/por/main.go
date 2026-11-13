// Command por is Pool of Radiance's install tool for gameimport
// (import.toml); the game itself is played by gobox's cmd/gobox from the
// adventure folder `gobox import por data/dos` writes.
//
//	por -setup                         the area scripts into data/assets/scripts, data/COVER.PNG
//	por -import-disk2 DIR -data data   the Amiga graphics from the .adf disks under DIR into data/amiga
package main

import (
	"flag"
	"log"
	"path/filepath"

	"gobox/amigaimport"
	"gobox/pool"
)

func main() {
	dir := flag.String("data", "", "where -import-disk2 writes (import.toml)")
	imp2 := flag.String("import-disk2", "", "convert the Amiga graphics from the .adf disks under this directory into -data/amiga, then exit (import.toml)")
	setup := flag.Bool("setup", false, "one-time setup after the game was imported: write the area scripts into data/assets/scripts and data/COVER.PNG from data/dos/TITLE.DAX, then exit (gameimport)")
	flag.Parse()
	switch {
	case *setup:
		if err := setupData("data"); err != nil {
			log.Fatal(err)
		}
	case *imp2 != "":
		if *dir == "" {
			log.Fatal("-import-disk2 needs -data")
		}
		// the Amiga data disk among the .adf images under -import-disk2 (loose or zipped): its graphics as DOS-layout archives in -data/amiga
		spec := amigaimport.Spec{Gfx: amigaGfx, Disks: []string{"poolgame", "POOLDATA"}, Program: "program", Intro: "intro", Skip: map[string]string{"8x8d.dax": "the Amiga set is laid out for its own wall renderer and does not match WALLDEF's symbol numbers: the 3D view stays DOS (setup.go amigaGfx)"}}
		if err := amigaimport.Import(*imp2, filepath.Join(*dir, pool.AmigaDir), spec); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("por plays no more: gobox import por data/dos, then gobox (gobox-engine/cmd/gobox)")
	}
}
