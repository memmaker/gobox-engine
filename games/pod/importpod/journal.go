package importpod

import (
	"embed"
	"gobox/game/journal"
	"io/fs"
	"sync"
)

// The Adventurer's Journal's texts (journal/JOURNAL.TXT), which PoD's
// JOURNAL command names by number only ("recorded as journal entry N").
//
//go:embed journal
var journalFS embed.FS

// journalBook is JOURNAL.TXT, loaded once.
var journalBook = sync.OnceValue(func() *journal.Book {
	sub, _ := fs.Sub(journalFS, "journal")
	b, err := journal.Load(sub, journalFile)
	if err != nil {
		panic(err)
	}
	return b
})

// journalFile is the book's file; journalEntryMax the last entry number
// the printed journal has.
const (
	journalFile     = "JOURNAL.TXT"
	journalEntryMax = 100
)

// The reading box: its title, width and rows a page.
const (
	journalBox   = "Adventurer's Journal"
	journalWidth = 36
	journalRows  = 18
)

// journalTitle is entry n's title in JOURNAL.TXT.
func journalTitle(n byte) string { return "Journal Entry " + itoa(int(n)) }

// Read is journal text the game waits to have read.
type Read struct {
	Title string
	Lines []string
}
