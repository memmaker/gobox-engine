package games

import (
	"embed"
	"io/fs"
)

// The Adventurer's Journal's texts (journal/JOURNAL.TXT): the scripts name
// them by number only ("YOU RECORD IT IN JOURNAL ENTRY 31", a tavern tale
// by its number); the import names each through Journal. gobox
// shows it to be read (State.Read) and the journal's notes keep it
// (State.Lore; the slot's NOTES).
//
//go:embed cotab/journal
var cotabJournalFS embed.FS

// JournalFS is the journal's books and pictures (the import writes them to the data folder).
func JournalFS() fs.FS {
	sub, _ := fs.Sub(cotabJournalFS, "cotab/journal")
	return sub
}

// The reading box: its title, width and rows a page.
const ()
