// Command pod runs Pools of Darkness from its imported adventure folder
// (gobox import pod) through the shared pool host (playpod).
//
//	pod [-adv DIR] [-scale N] [-seed N] [app.Flags: -shot, -keys ...]
//	pod -setup [-data DIR]
//	pod -import-disk2 ADFDIR [-data DIR]
//
// -setup is gameimport's step (import.toml): it writes DIR/../COVER.PNG
// from the originals' DIR/TITLE.DAX and exits; -import-disk2 converts the
// Amiga disks into amiga/ beside DIR, which gobox import pod reads.
package main

import (
	"flag"
	"log"
	"path/filepath"

	"gobox/amigaimport"
	"gobox/app"
	"gobox/games/pod/playpod"
)

func main() {
	adv := flag.String("adv", "../../data/pod", "the imported adventure folder (gobox import pod)")
	dir := flag.String("data", "../../data/pod/sources/dos", "the original DOS game files (-setup, -import-disk2 only)")
	scale := flag.Int("scale", 2, "window scale")
	imp2 := flag.String("import-disk2", "", "convert the Amiga graphics from the .adf disks under this directory into amiga/ beside -data (data/amiga), then exit (import.toml)")
	setup := flag.Bool("setup", false, "one-time setup after gameimport copied the game into -data: write COVER.PNG beside it, then exit")
	seed := flag.Int64("seed", -1, "a fixed dice seed, so the same keys give the same run (scripted tests; -1: the clock)")
	captureOf := app.Flags(map[byte]app.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00, 'F': 0x4400}) // -keys escapes: \r Enter, \e Escape, \U \D \L \R the arrows, \F F10 (save/load)
	flag.Parse()
	if *setup {
		if err := app.WriteCover(*dir, filepath.Join(filepath.Dir(filepath.Clean(*dir)), "COVER.PNG")); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *imp2 != "" {
		// the three POD disks among the .adf images (loose or zipped): their graphics as DOS-layout
		// archives in amiga/ beside -data (data/amiga), where gobox import pod looks
		// Exe: the program holding the fixed palette (Gfx Pal); the music (pod.mx) and the sampled effects (SOUNDS.GLB) are copied as they are
		spec := amigaimport.Spec{Gfx: amigaGfx, Disks: []string{"POD_Disk1_v.1.00", "POD_Disk2_v.1.00", "POD_Disk3_v.1.00"}, Exe: "Pools of Darkness",
			Files: []string{"pod.mx", "SOUNDS.GLB"}, Alias: map[string]string{"8x8da": "8x8d"}}
		if err := amigaimport.Import(*imp2, filepath.Join(filepath.Dir(filepath.Clean(*dir)), "amiga"), spec); err != nil {
			log.Fatal(err)
		}
		return
	}
	s := app.Start{Adventure: *adv, Data: "../../data/pod/sources", Settings: app.Settings{Scale: *scale, Window: true}, Capture: captureOf(), Seed: *seed}
	if err := playpod.Play(s); err != nil {
		log.Fatal(err)
	}
}

// pal is the offset of Pools of Darkness' fixed Amiga palette in its program
// (Spec.Exe; the same 16 colours as Curse's and Silver Blades'): the pictures
// that carry none of their own (CBODY, CHEAD, COMSPR, CPIC) are drawn in it.
const pal = 0x51354

// amigaGfx is how the Amiga disks' archives decode (gobox amigaimport's table for the Pool games).
var amigaGfx = map[string]amigaimport.Gfx{
	"8x8d": {Pal: pal}, "bigpic": {Pal: pal}, "body": {Pal: pal}, "head": {Pal: pal}, "cbody": {Pal: pal}, "chead": {Pal: pal},
	"cpic": {Pal: pal}, "comspr": {Pal: pal}, "tiles": {Pal: pal}, "back": {Pal: pal}, "skygrnd": {Pal: pal}, "gen": {Pal: pal},
	"dungcom": {Frames: true, Base: 1, Pal: pal}, "wildcom": {Frames: true, Base: 1, Pal: pal}, "randcom": {Frames: true, Base: 1, Pal: pal},
	"sky":   {Base: 0xfa, Pal: pal},
	"title": {Pal: pal}, "intro": {Pal: pal},
	"pic": {Sprite: true, XOR: true, Pal: pal}, "final": {Sprite: true, XOR: true, Pal: pal}, "frame": {Sprite: true, Pal: pal},
	"sprit": {Sprite: true, Pal: pal},
}
