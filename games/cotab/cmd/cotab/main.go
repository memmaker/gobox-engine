// Command cotab plays Curse of the Azure Bonds from its imported adventure
// folder (gobox import cotab; playcotab, as cmd/gobox does), and converts
// the Amiga disks' graphics for that import:
//
//	cotab [-adv DIR] [-scale N] [-seed N] [app.Flags' -shot, -keys, -slot ...]
//	cotab -import-disk2 ADFDIR [-amiga DIR]
//
// DIR is the adventure folder (default adventures/cotab). It goes once
// cmd/gobox has every option it has.
package main

import (
	"flag"
	"log"

	"gobox/amigaimport"
	"gobox/app"

	"gobox/games/cotab/playcotab"
)

func main() {
	adv := flag.String("adv", "../../data/cotab", "the imported adventure folder (gobox import cotab)")
	imp2 := flag.String("import-disk2", "", "convert the Amiga graphics from the .adf disks under this directory into -amiga, then exit (import.toml)")
	amigaDir := flag.String("amiga", "../../data/cotab/sources/amiga", "where -import-disk2 writes the Amiga archives (beside the DOS originals gobox import reads)")
	scale := flag.Int("scale", 2, "window scale")
	seed := flag.Int64("seed", -1, "a fixed dice seed, so the same keys give the same run (scripted tests; -1: the clock)")
	captureOf := app.Flags(map[byte]app.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00, 'F': 0x4400}) // -keys \r Enter, \e Escape, \U \D \L \R arrows, \F F10 (save/load)
	flag.Parse()
	if *imp2 != "" {
		// the two Curse disks among the .adf images (loose or zipped): their graphics as DOS-layout
		// archives, which gobox import cotab reads beside the DOS originals
		spec := amigaimport.Spec{Gfx: amigaGfx, Disks: []string{"CurseA", "CurseB"}, Exe: "Curse", Files: []string{"SOUNDS.GLB"}} // Exe: the program holding the fixed palette (Gfx Pal); SOUNDS.GLB the sampled effects (rules.AmigaMusic)
		if err := amigaimport.Import(*imp2, *amigaDir, spec); err != nil {
			log.Fatal(err)
		}
		return
	}
	s := app.Start{Adventure: *adv, Data: "../../data/cotab/sources", Settings: app.Settings{Scale: *scale}, Capture: captureOf(), Seed: *seed}
	if err := playcotab.Play(s); err != nil {
		log.Fatal(err)
	}
}

// pal is the offset of Curse's fixed Amiga palette in its program (Spec.Exe);
// TITLE and BIGPIC keep their own.
const pal = 0x47410

// amigaGfx is how the Amiga disks' archives decode (gobox amigaimport's
// table for the later Pool games); Mask is cotab's transparent colour: 0xD
// for the 3D view's wall pieces and sky, 0 for the combat pictures.
var amigaGfx = map[string]amigaimport.Gfx{
	"8x8d": {Mask: 0xD, Pal: pal}, "bigpic": {Pal: pal}, "body": {Pal: pal}, "head": {Pal: pal}, "cbody": {Pal: pal}, "chead": {Pal: pal},
	"cpic": {Pal: pal}, "comspr": {Pal: pal}, "tiles": {Pal: pal}, "back": {Mask: 0xD, Pal: pal}, "skygrnd": {Mask: 0xD, Pal: pal}, "gen": {Pal: pal},
	"dungcom": {Frames: true, Base: 1, Pal: pal}, "wildcom": {Frames: true, Base: 1, Pal: pal}, "randcom": {Frames: true, Base: 1, Pal: pal},
	"sky":   {Base: 0xfa, Mask: 0xD, Pal: pal},
	"title": {Pal: pal}, "intro": {Pal: pal},
	"pic": {Sprite: true, XOR: true, Pal: pal}, "final": {Sprite: true, XOR: true, Pal: pal}, "frame": {Sprite: true, Pal: pal},
	"sprit": {Sprite: true, Pal: pal},
}
