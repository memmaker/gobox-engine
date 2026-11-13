// Command gobox plays the adventures imported into its adventures directory
// and imports them:
//
//	gobox [flags] [adventure]       play it (none: a numbered menu of the games in -data)
//	gobox import [-out dir] <game> <originals>
//
// An adventure is a folder <data>/<id>/ (default data/, run from
// gobox-engine) with a valid manifest.json (gobox/importer); the binary reads
// nothing else of the game at run time, no original file and no check that
// one exists. The folder's sources/ is the host's directory for that game:
// the port's own assets (fonts), the saves and the player's music.
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
	"path/filepath"
	"strings"

	"buckrogers/playbr1"
	"buckrogers2/playbr2"
	"gobox/app"
	"gobox/engine/playpool"
	"gobox/importer"
	"gobox/ui"

	_ "buckrogers/importbr1"
	_ "buckrogers2/importbr2"
	_ "gobox/importer/games"
)

// players start the games that are not on rules1e's pool host (playpool.Play
// reads everything else of a rules1e game from its manifest). ponytail: one
// engine per game until the engine loop is gobox's (Track R4/R5).
var players = map[string]func(app.Start) error{
	"br1": playbr1.Play,
	"br2": playbr2.Play,
}

// ruleset1e is the manifest Ruleset the pool host plays.
const ruleset1e = "rules1e"

// play starts the adventure folder f in the engine of its game.
func play(s app.Start, f importer.Folder) error {
	if f.Ruleset == ruleset1e {
		return playpool.Play(s)
	}
	return players[f.ID](s)
}

// escapes are the -keys escapes beside the shell's (app.Flags): \r Enter,
// \e Escape, \U \D \L \R the arrows, \1..\8 F1..F8, \0 F10.
func escapes() map[byte]app.Key {
	esc := map[byte]app.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00}
	for n := range 8 { // \1..\8: F1..F8
		esc['1'+byte(n)] = ui.KeyF1 + uint16(n)<<8
	}
	esc['0'] = ui.KeyF1 + 9<<8 // \0: F10
	return esc
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "import" {
		if err := importer.Main(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	root := flag.String("data", "data", "the games' data directory: one imported folder per game (gobox import -out writes it)")
	s := app.Start{}
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
		if f.Ruleset == ruleset1e || players[f.ID] != nil {
			playable = append(playable, f)
		}
	}
	if len(playable) == 0 {
		fmt.Fprintf(os.Stderr, "no game in %s: import one with gobox import -out %s <game> <originals>; games: %s\n", *root, *root, strings.Join(importer.IDs(), ", "))
		os.Exit(2)
	}
	f, ok := pick(playable, flag.Arg(0))
	if !ok {
		os.Exit(2)
	}
	s.Adventure = f.Dir
	s.Data = filepath.Join(f.Dir, "sources")
	if err := play(s, f); err != nil {
		log.Fatal(err)
	}
}

// pick is the folder named want, or with want empty the one whose number
// the player types into the menu printed on stdout.
func pick(playable []importer.Folder, want string) (importer.Folder, bool) {
	if want != "" {
		for _, f := range playable {
			if f.ID == want {
				return f, true
			}
		}
		fmt.Fprintf(os.Stderr, "no game %q\n", want)
		return importer.Folder{}, false
	}
	for i, f := range playable {
		fmt.Printf("%d. %s\n", i+1, f.Title)
	}
	fmt.Print("Game number: ")
	var n int
	if _, err := fmt.Scan(&n); err != nil || n < 1 || n > len(playable) {
		fmt.Fprintln(os.Stderr, "no such game")
		return importer.Folder{}, false
	}
	return playable[n-1], true
}
