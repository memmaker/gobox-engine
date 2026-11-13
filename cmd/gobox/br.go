package main

import (
	"flag"
	"path/filepath"

	"gobox/game"
	"gobox/importer"
	"gobox/ui"

	"buckrogers/playbr1"
	"buckrogers2/playbr2"
)

// br1Dev, br2Dev are the Buck Rogers games' test-only starts, filled from
// the command line (brFlags).
var (
	br1Dev playbr1.Dev
	br2Dev playbr2.Dev
)

// brFlags registers the Buck Rogers test starts (call before flag.Parse).
func brFlags() {
	flag.StringVar(&br1Dev.Start, "teststart", "", "test only (br1): a test start after the opening (rules/debug.go: e.g. fight)")
	flag.IntVar(&br2Dev.Block, "block", 0, "test only (br2): a new game starts in this ECL block instead of 16")
	flag.IntVar(&br2Dev.Credits, "credits", 0, "test only (br2): every member starts a new game with these credits")
	flag.IntVar(&br2Dev.HP, "hp", 0, "test only (br2): every member starts a new game with at most these hit points")
	flag.IntVar(&br2Dev.Space, "space", 0, "test only (br2): a new game opens with a space fight against this ship, 1-5")
	flag.BoolVar(&br2Dev.Rest, "rest", false, "test only (br2): a new game opens the rest menu first")
	flag.IntVar(&br2Dev.FightN, "fightn", 0, "test only (br2): how many of -fight's monster (0: one)")
	flag.IntVar(&br2Dev.Shop, "shop", 0, "test only (br2): a new game opens in this shop (ECL SHOP n)")
}

// players start the games that are not on rules1e's pool host (playPool
// reads everything else of a rules1e game from its manifest): an engine and
// its data. ponytail: one engine per game until the engine loop is gobox's
// (Track R4/R5).
var players = map[string]func(s start) (game.Game, game.Data, error){
	"br1": func(s start) (game.Game, game.Data, error) {
		return playbr1.Play(s.Adventure, s.Capture.Dev.Saves, s.Seed, br1Dev)
	},
	"br2": func(s start) (game.Game, game.Data, error) {
		d := br2Dev
		d.Fight = s.Capture.Dev.Fight
		return playbr2.Play(s.Adventure, s.Capture.Dev.Saves, s.Seed, d)
	},
}

// playOther starts f's game from players and puts gobox's front on it.
func playOther(s start, f importer.Folder) (ui.Launch, error) {
	g, d, err := players[f.ID](s)
	if err != nil {
		return ui.Launch{}, err
	}
	d.OwnMusic = filepath.Join(s.Data, "music")     // the player's own music, by trigger
	ui.SetScripts(f.Manifest.Start, f.Manifest.End) // the title, credits and end gobox composes
	w, h := window(f)
	return ui.Launch{Front: ui.NewFront(ui.NewGame(g, d), d), Title: f.Title, W: w, H: h, Apply: ui.Tune(g)}, nil
}

// defaultWindow is the window's size at scale 1 for a manifest without one.
var defaultWindow = importer.Size{W: 320, H: 200}

// window is f's window size at scale 1.
func window(f importer.Folder) (w, h int) {
	if f.Window == (importer.Size{}) {
		return defaultWindow.W, defaultWindow.H
	}
	return f.Window.W, f.Window.H
}
