// Command gobox plays the adventures imported into its adventures directory
// and imports them:
//
//	gobox [flags] [adventure]       play it (none: a numbered menu of the games in -data)
//	gobox import [-out dir] [-amiga] <game> <originals>
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
	"time"

	"gobox/importer"
	"gobox/ui"
)

// ruleset1e is the manifest Ruleset the pool host plays.
const ruleset1e = "rules1e"

// rulesetXXVc is the manifest Ruleset of the Buck Rogers games; the shared
// engine plays it only under -shared.
const rulesetXXVc = "rulesxxvc"

// launch is f's game as ui.Launch: on the shared engine (its ruleset's host:
// the pool games always, a rulesxxvc game under -shared) or on its own old one.
func launch(s start, f importer.Folder) (ui.Launch, error) {
	if hs, ok := hosts[f.Ruleset]; ok && (f.Ruleset == ruleset1e || s.Shared) {
		return playPool(s, hs)
	}
	return playOther(s, f)
}

// play runs the adventure folder f in the engine of its game until the
// window closes.
func play(s start, f importer.Folder) error {
	l, err := launch(s, f)
	if err != nil {
		return err
	}
	return ui.Play(l, s.Settings, s.Capture)
}

// escapes are the -keys escapes beside the shell's (parseKeys): \r Enter,
// \e Escape, \U \D \L \R the arrows, \1..\8 F1..F8, \0 F10.
func escapes() map[byte]ui.Key {
	esc := map[byte]ui.Key{'r': 13, 'e': 27, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00}
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
	s := start{}
	flag.IntVar(&s.Settings.Scale, "scale", 3, "window scale (with -window)")
	flag.BoolVar(&s.Settings.Window, "window", false, "a window instead of the full screen (the Esc menu's setting is left as it is)")
	flag.BoolVar(&s.Settings.Fullscreen, "fullscreen", false, "the screen at its own resolution")
	flag.StringVar(&s.Param, "param", "", "the original's command-line parameter")
	flag.BoolVar(&s.Explore, "explore", false, "debug: skip the intro, Quick Start a party if there is none and begin adventuring at once")
	flag.BoolVar(&s.Battle, "battle", false, "debug: a fight the computer plays out for both sides at the first prompt (with -explore: at once)")
	flag.StringVar(&s.Dev, "dev", "", "debug: a dev scene (e.g. combat) at the first adventuring prompt")
	seed := flag.Int64("seed", -1, "debug: the same dice every run (e.g. 1), for comparing captures; <0: the clock, but a scripted run (-keys) without it plays seed 1")
	captureOf := captureFlags(escapes())
	flag.BoolVar(&s.Shared, "shared", false, "dev, temporary until the old Buck Rogers engines go: play br1 on the shared engine from a folder imported with `gobox import -shared`")
	brFlags()
	flag.Parse()
	s.Capture = captureOf()
	s.Seed = resolveSeed(*seed, len(s.Capture.Keys) > 0, time.Now())

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

// resolveSeed is the dice seed of a run: flag (-seed) when it is >= 0, else
// 1 for a scripted run (so each prefix's shot is the same game), else the clock now.
func resolveSeed(flag int64, scripted bool, now time.Time) uint64 {
	switch {
	case flag >= 0:
		return uint64(flag)
	case scripted:
		return 1
	}
	return uint64(now.UnixNano())
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
