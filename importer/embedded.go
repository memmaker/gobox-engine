package importer

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

// gameFiles are the games' embedded data: each game's script names and
// layout (scripts/layout/global.txt, made from the game's Pascal; the data
// of gobox ecl Library.Data) and its Adventurer's Journal books with
// the pictures they show (gobox's journal format).
//
//go:embed cotab/scripts pod/scripts sotsb/scripts cotab/journal pod/journal por/journal sotsb/journal
var gameFiles embed.FS

// embeddedScripts is game's scripts' data.
func embeddedScripts(game string) fs.FS { return embeddedDir(game + "/scripts") }

// embeddedJournal is game's journal books and pictures.
func embeddedJournal(game string) fs.FS { return embeddedDir(game + "/journal") }

func embeddedDir(dir string) fs.FS {
	d, err := fs.Sub(gameFiles, dir)
	if err != nil {
		panic(err) // a constant, valid path
	}
	return d
}

// ImportJournal writes game's journal books and pictures into dir (the
// engine reads them from the adventure folder's journal folder); the
// entries that name them are the scripts' names/entries.json.
func ImportJournal(game, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	files, err := fs.ReadDir(embeddedJournal(game), ".")
	if err != nil {
		return err
	}
	for _, f := range files {
		b, err := fs.ReadFile(embeddedJournal(game), f.Name())
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, f.Name()), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
