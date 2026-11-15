package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gobox/importer"
)

// The Buck Rogers originals the br1/br2 tests import from; without them
// those tests skip (rules.New fails on a folder that is not there).
var (
	br1Dos = filepath.Join("..", "..", "..", "buckrogers-go", "data", "dos")
	br2Dos = filepath.Join(os.Getenv("HOME"), "Projects", "buckrogers2-re", "gog_en")
)

// brAdvs is the temporary root of the imported adventure folders (TestMain removes it).
var brAdvs = sync.OnceValue(func() string {
	dir, err := os.MkdirTemp("", "br-adv")
	if err != nil {
		log.Fatal(err)
	}
	return dir
})

// brAdv imports game id from the originals in dos once per test binary,
// as gobox import writes it, and is its folder.
func brAdv(id, dos string) func() string {
	return sync.OnceValue(func() string {
		if _, err := os.Stat(filepath.Join(dos, "ECL1.DAX")); err == nil {
			if err := importer.Run(id, dos, brAdvs()); err != nil {
				log.Fatal(err)
			}
		}
		return filepath.Join(brAdvs(), id)
	})
}

var br1Adv, br2Adv = brAdv("br1", br1Dos), brAdv("br2", br2Dos)

// br1BareAdv is br1Adv without its saves (a fresh roster): its other parts linked.
func br1BareAdv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ents, err := os.ReadDir(br1Adv())
	if err != nil {
		t.Skip(err)
	}
	for _, e := range ents {
		if e.Name() == "saves" {
			continue
		}
		if err := os.Symlink(filepath.Join(br1Adv(), e.Name()), filepath.Join(dir, e.Name())); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
