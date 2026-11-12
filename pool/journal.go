package pool

import (
	"embed"
	"fmt"
	"image"
	"io/fs"
	"os"
	"strings"
	"sync"

	"gobox/ecl/script"
	"gobox/game/journal"
)

// Journal texts (extra): the game often only names an
// entry of the printed Adventurer's Journal ("SEE JOURNAL ENTRY 26", "YOU
// OVERHEAR TAVERN TALE 5", the proclamations at City Hall). The import's
// special cases (scriptfixes.go) name it with Journal instead, and the
// pager shows it right there, with its pictures. The
// texts and pictures (pre-scaled to the screen, one shared 16-colour
// palette) are built in from journal/, in gobox's journal format; gobox
// (game/journal) reads, finds and lays them out.

//go:embed journal
var journalFS embed.FS

// books are journal/'s files, each a journal.Book, loaded once.
var books = sync.OnceValue(func() map[string]*journal.Book {
	sub, _ := fs.Sub(journalFS, "journal")
	out := map[string]*journal.Book{}
	for _, f := range []string{"JOURNAL.TXT", "HISTORY.TXT", "CLUEBOOK.TXT"} {
		b, err := journal.Load(sub, f)
		if err != nil {
			panic(err) // built in: never
		}
		out[f] = b
	}
	return out
})

// book is journal/name's Book.
func book(name string) *journal.Book { return books()[name] }

// journalLayout is the texts titled titles of journal/name laid out for the
// pager (journal.Book.Layout's markers); voiced marks each paragraph with
// its own recording's key.
func journalLayout(name string, titles []string, voiced bool) []string {
	var voice func(string) string
	if voiced {
		voice = voiceKey
	}
	return book(name).Layout(titles, jrWidth, jrRows(), voice)
}

// journalPicture is a picture of journal/ (Data.Scans); nil if missing.
func journalPicture(name string) image.Image { return book("JOURNAL.TXT").Picture(name) }

// WrapWords breaks s into lines of at most w characters at spaces.
func WrapWords(s string, w int) []string { return journal.Wrap(s, w) }

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
// long text entry n of the adventure's text data (script.Entry) in the
// pager. A voiced entry is voiced every time it is asked for.
func (e *Engine) opJournal() {
	e.vm.ReadArgs(1)
	e.showEntry(script.EntryID(e.vm.Arg(1, e.host)))
}

// showEntry shows entry id in the pager, after the entries already up in the
// same box (one script's Journal commands are read together), and notes a
// journal text (noteJournal). The text window stays as the script left it.
func (e *Engine) showEntry(id script.EntryID) {
	en, ok := e.entries()[id]
	if !ok {
		panic(fmt.Sprintf("script: no text entry %d (%s)", id, script.EntriesFile)) // a script naming no entry: a broken game folder
	}
	lines := journalLayout(en.Book, []string{en.Title}, en.Voiced)
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
		r.Lines = append(append(r.Lines, ""), lines...) // as journal.Book.Layout of both titles
		return
	}
	e.showJournal(en.Box, lines)
}

// entryNamed is the id of the entry the scripts call name (a dev view's).
func (e *Engine) entryNamed(name string) script.EntryID {
	for id, en := range e.entries() {
		if en.Name == name {
			return id
		}
	}
	panic("no text entry " + name)
}

// entries are the adventure's long texts by id, read once.
func (e *Engine) entries() map[script.EntryID]script.Entry {
	if e.entryTable == nil {
		es, err := script.ReadEntries(os.DirFS(e.scriptDir()))
		if err != nil {
			panic(err) // a broken game folder
		}
		e.entryTable = map[script.EntryID]script.Entry{}
		for _, en := range es {
			e.entryTable[en.ID] = en
		}
	}
	return e.entryTable
}
