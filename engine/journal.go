package engine

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gobox/ecl"
	"gobox/game"
)

// Journal texts (extra): the game often only names an
// entry of the printed Adventurer's Journal ("SEE JOURNAL ENTRY 26", "YOU
// OVERHEAR TAVERN TALE 5", the proclamations at City Hall). The import's
// special cases (scriptfixes.go) name it with Journal instead, and the
// pager shows it right there, with its pictures. The books and pictures
// (pre-scaled to the screen, one shared 16-colour palette) are the game's
// own, in the adventure folder's JournalDir, written by its import, in gobox's
// journal format; gobox (game/journal.go) reads, finds and lays them out.

// JournalDir is the game's journal books and pictures in the adventure folder.
const JournalDir = "journal"

// book is the game's journal book name (a file of JournalDir), loaded once;
// nil if the game has none.
func (e *Engine) book(name string) *game.JournalBook {
	if b, ok := e.books[name]; ok {
		return b
	}
	b, err := game.LoadJournal(os.DirFS(filepath.Join(e.adv, JournalDir)), name)
	if errors.Is(err, fs.ErrNotExist) {
		b = nil
	} else if err != nil {
		panic(err) // a broken game folder
	}
	if e.books == nil {
		e.books = map[string]*game.JournalBook{}
	}
	e.books[name] = b
	return b
}

// journalLayout is the texts titled titles of book name laid out for the
// pager (game.Book.Layout's markers); voiced marks each paragraph with
// its own recording's key.
func (e *Engine) journalLayout(name string, titles []string, voiced bool) []string {
	var voice func(string) string
	if voiced {
		voice = voiceKey
	}
	b := e.book(name)
	if b == nil {
		panic(fmt.Sprintf("script: no journal book %s in %s", name, filepath.Join(e.adv, JournalDir))) // a broken game folder
	}
	return b.Layout(titles, jrWidth, jrRows(), voice)
}

// journalPicture is a picture of the game's journal (Data.Scans); nil if missing.
func (e *Engine) journalPicture(name string) image.Image {
	if b := e.book("JOURNAL.TXT"); b != nil {
		return b.Picture(name)
	}
	return nil
}

// WrapWords breaks s into lines of at most w characters at spaces.
func WrapWords(s string, w int) []string { return game.WrapText(s, w) }

// The pager: the text inside a full-screen box, jrRows lines a page from
// row jrTop (under the journal's title and a blank row) to two rows above
// the "press a key" line (messageRow); a taller page holds more lines.
const (
	journalBox = "Adventurer's Journal"
	jrWidth    = 0x24
	jrTop      = 3
)

func jrRows() int { return messageRow - 1 - jrTop }

// showJournal has the player read lines under the box title box.
func (e *Engine) showJournal(box string, lines []string) {
	e.read(&Read{Title: box, Lines: lines})
}

// read publishes r (State.Read) with the next Seq; nothing waits on it
// (gobox shows, pages and closes it), and the next decision clears it.
func (e *Engine) read(r *Read) {
	e.readSeq++
	r.Seq = e.readSeq
	e.ext.read = r
}

// opJournal is the pager command (gobox ecl Journal, Felix 2026-11-11): the
// long text entry n of the adventure's text data (ecl.Entry) in the
// pager. A voiced entry is voiced every time it is asked for.
func (e *Engine) opJournal() {
	e.vm.ReadArgs(1)
	e.showEntry(ecl.EntryID(e.vm.Arg(1, e.host)))
}

// showEntry shows entry id in the pager, after the entries already up in the
// same box (one script's Journal commands are read together), and notes a
// journal text (noteJournal). The text window stays as the script left it.
func (e *Engine) showEntry(id ecl.EntryID) {
	en, ok := e.entries()[id]
	if !ok {
		panic(fmt.Sprintf("script: no text entry %d (%s)", id, ecl.EntriesFile)) // a script naming no entry: a broken game folder
	}
	lines := e.journalLayout(en.Book, []string{en.Title}, en.Voiced)
	if en.Voiced {
		e.ext.again = map[string]bool{}
		for _, l := range lines {
			switch {
			case strings.HasPrefix(l, "\x00"):
				e.ext.again[strings.ReplaceAll(l[1:], " ", "_")] = true
			case strings.HasPrefix(l, "\x03"):
				key, _, _ := strings.Cut(l[1:], "\x03")
				e.ext.again[key] = true
			}
		}
	}
	if en.Book == "JOURNAL.TXT" {
		e.noteJournal(en.Title)
	}
	if r := e.ext.read; r != nil && r.Title == en.Box {
		r.Lines = append(append(r.Lines, ""), lines...) // as game.Book.Layout of both titles
		return
	}
	e.showJournal(en.Box, lines)
}

// entryNamed is the id of the entry the scripts call name (a dev view's).
func (e *Engine) entryNamed(name string) ecl.EntryID {
	for id, en := range e.entries() {
		if en.Name == name {
			return id
		}
	}
	panic("no text entry " + name)
}

// entries are the adventure's long texts by id, read once.
func (e *Engine) entries() map[ecl.EntryID]ecl.Entry {
	if e.entryTable == nil {
		es, err := ecl.ReadEntries(os.DirFS(e.scriptDir()))
		if err != nil {
			panic(err) // a broken game folder
		}
		e.entryTable = map[ecl.EntryID]ecl.Entry{}
		for _, en := range es {
			e.entryTable[en.ID] = en
		}
	}
	return e.entryTable
}
