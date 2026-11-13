// Command blades plays Secret of the Silver Blades from its imported
// adventure folder (gobox import sotsb; playsotsb, as cmd/gobox does), and
// converts the Amiga disks' graphics and music for that import:
//
//	blades [-adv DIR] [-scale N] [app.Flags: -shot, -keys ...]
//	blades -import-amiga ADFDIR [-amiga DIR]
//
// DIR is the adventure folder (default adventures/sotsb). The saves are
// UserConfigDir/gobox/sotsb, seeded once from the folder's saves.
package main

import (
	"flag"
	"log"

	"gobox/amigaimport"
	"gobox/app"

	"sotsb/importsotsb"
	"sotsb/playsotsb"
)

func main() {
	adv := flag.String("adv", "../../data/sotsb", "the imported adventure folder (gobox import sotsb)")
	scale := flag.Int("scale", 2, "window scale")
	seed := flag.Int64("seed", -1, "a fixed dice seed, so the same keys give the same run (scripted tests; -1: the clock)")
	captureOf := app.Flags(map[byte]app.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00, 'F': 0x4400}) // -keys escapes: \r Enter, \e Escape, \U \D \L \R the arrows, \F F10 (save, load)
	imp := flag.String("import-amiga", "", "convert the Amiga graphics and music from the .adf disks under this directory into -amiga, then exit")
	amigaDir := flag.String("amiga", "../../data/sotsb/sources/amiga", "where -import-amiga writes the Amiga archives (beside the DOS originals gobox import reads)")
	flag.Parse()
	if *imp != "" {
		// the two Silver Blades disks among the .adf images, as DOS-layout archives gobox import sotsb reads
		spec := amigaimport.Spec{Gfx: amigaGfx, Exe: "Secret", Disks: []string{"Secret 1", "Secret 2"}, Files: []string{importsotsb.AmigaMusic.Songs, importsotsb.AmigaMusic.Effects}}
		if err := amigaimport.Import(*imp, *amigaDir, spec); err != nil {
			log.Fatal(err)
		}
		return
	}
	s := app.Start{Adventure: *adv, Data: "../../data/sotsb/sources", Settings: app.Settings{Scale: *scale, Window: true}, Capture: captureOf(), Seed: *seed}
	if err := playsotsb.Play(s); err != nil {
		log.Fatal(err)
	}
}

// amigaGfx is how the Amiga disks' archives decode (gobox amigaimport's table for the later
// games); Base/Frames say how the unindexed libraries map to DOS block ids, Pal where the
// program Secret holds the fixed palette of the pictures that carry none (PIC and BIGPIC do).
const amigaPal = 0x4fdea

var amigaGfx = map[string]amigaimport.Gfx{
	"8x8d": {Pal: amigaPal}, "bigpic": {Pal: amigaPal}, "body": {Pal: amigaPal}, "head": {Pal: amigaPal}, "cbody": {Pal: amigaPal}, "chead": {Pal: amigaPal},
	"cpic": {Pal: amigaPal}, "comspr": {Pal: amigaPal}, "tiles": {Pal: amigaPal}, "back": {Pal: amigaPal}, "skygrnd": {Pal: amigaPal}, "gen": {Pal: amigaPal},
	"dungcom": {Frames: true, Base: 1, Pal: amigaPal}, "wildcom": {Frames: true, Base: 1, Pal: amigaPal}, "randcom": {Frames: true, Base: 1, Pal: amigaPal},
	"sky":   {Base: 0xfa, Pal: amigaPal},
	"title": {Pal: amigaPal}, "intro": {Pal: amigaPal},
	"pic": {Sprite: true, XOR: true, Pal: amigaPal}, "final": {Sprite: true, XOR: true, Pal: amigaPal}, "frame": {Sprite: true, Pal: amigaPal},
	"sprit": {Sprite: true, Pal: amigaPal},
}
