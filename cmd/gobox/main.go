// Command gobox plays the adventures imported into its adventures directory
// and imports them:
//
//	gobox [flags] [adventure]       play it (none: the only one, or the list)
//	gobox import [-out dir] <game> <originals>
//
// An adventure is a folder adventures/<id>/ with a valid manifest.json
// (gobox/importer); the binary reads nothing else of the game at run time,
// no original file and no check that one exists. -data is the host's own
// directory: the port's assets (fonts, speech, sound pack, soundtrack, the
// Amiga sound) and the saves.
//
// The Windows exe's icon: icon.ico (the Amiga Pool of Radiance title
// picture's hero and dragon) linked in through rsrc_windows_amd64.syso:
//
//go:generate go run ./winicon -ico icon.ico -o rsrc_windows_amd64.syso
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"buckrogers/playbr1"
	"buckrogers2/playbr2"
	"cotab/playcotab"
	"gobox/app"
	"gobox/front/present"
	"gobox/importer"
	"pod/playpod"
	"por/playpor"
	"sotsb/playsotsb"

	_ "buckrogers/importbr1"
	_ "buckrogers2/importbr2"
	_ "cotab/importcotab"
	_ "pod/importpod"
	_ "por/importpor"
	_ "sotsb/importsotsb"
)

// players start an adventure's engine by the manifest's ID. ponytail: one
// engine per game until the engine loop is gobox's (Track R4/R5); then the
// manifest's Ruleset picks the compiled-in ruleset and this table goes.
var players = map[string]func(app.Start) error{
	"por":   playpor.Play,
	"cotab": playcotab.Play,
	"pod":   playpod.Play,
	"br1":   playbr1.Play,
	"br2":   playbr2.Play,
	"sotsb": playsotsb.Play,
}

// escapes are the -keys escapes beside the shell's (app.Flags): \r Enter,
// \e Escape, \U \D \L \R the arrows, \1..\8 F1..F8, \0 F10.
func escapes() map[byte]app.Key {
	esc := map[byte]app.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00}
	for n := range 8 { // \1..\8: F1..F8
		esc['1'+byte(n)] = present.KeyF1 + uint16(n)<<8
	}
	esc['0'] = present.KeyF1 + 9<<8 // \0: F10
	return esc
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "import" {
		if err := importer.Main(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	root := flag.String("adventures", "adventures", "the adventures directory (gobox import writes it)")
	s := app.Start{}
	flag.StringVar(&s.Data, "data", "data", "the host's data directory: the port's assets and the saves")
	flag.IntVar(&s.Settings.Scale, "scale", 3, "window scale (with -window)")
	flag.BoolVar(&s.Settings.Window, "window", false, "a window instead of the full screen (the Esc menu's setting is left as it is)")
	flag.BoolVar(&s.Settings.Fullscreen, "fullscreen", false, "the screen at its own resolution")
	flag.StringVar(&s.Param, "param", "", "the original's command-line parameter")
	flag.BoolVar(&s.Explore, "explore", false, "debug: skip the intro, Quick Start a party if there is none and begin adventuring at once")
	flag.BoolVar(&s.Battle, "battle", false, "debug: a fight the computer plays out for both sides at the first prompt (with -explore: at once)")
	flag.StringVar(&s.Dev, "dev", "", "debug: a dev scene (e.g. combat) at the first adventuring prompt")
	flag.Int64Var(&s.Seed, "seed", -1, "debug: the same dice every run (e.g. 1), for comparing captures")
	flag.StringVar(&s.Open, "open", "", "debug: open the remastered map there, map:NAME:X:Y:D")
	captureOf := app.Flags(escapes())
	flag.Parse()
	s.Capture = captureOf()

	folders, skipped, err := importer.List(*root)
	if err != nil {
		log.Fatal(err)
	}
	for _, err := range skipped {
		log.Print("not listed: ", err)
	}
	var playable []importer.Folder
	for _, f := range folders {
		if players[f.ID] != nil {
			playable = append(playable, f)
		}
	}
	want := flag.Arg(0)
	if want == "" && len(playable) == 1 {
		want = playable[0].ID
	}
	for _, f := range playable {
		if f.ID == want {
			s.Adventure = f.Dir
			if err := players[f.ID](s); err != nil {
				log.Fatal(err)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "usage: gobox [flags] <adventure>; in %s:\n", *root)
	for _, f := range playable {
		fmt.Fprintf(os.Stderr, "  %-8s %s (%s)\n", f.ID, f.Title, f.Ruleset)
	}
	if len(playable) == 0 {
		fmt.Fprintf(os.Stderr, "  none: import one with gobox import <game> <originals>; games: %s\n", strings.Join(importer.IDs(), ", "))
	}
	os.Exit(2)
}
