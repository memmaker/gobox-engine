package importcotab

import (
	"embed"
	"io/fs"
	"strings"
	"sync"

	eclscript "gobox/ecl/script"
	"gobox/game/journal"
)

// The Adventurer's Journal's texts (journal/JOURNAL.TXT): the scripts name
// them by number only ("YOU RECORD IT IN JOURNAL ENTRY 31", a tavern tale
// by its number); the import names each through Journal. gobox
// shows it to be read (State.Read) and the journal's notes keep it
// (State.Lore; the slot's NOTES).
//
//go:embed journal
var journalFS embed.FS

// JournalFS is the journal's books and pictures (the import writes them to the data folder).
func JournalFS() fs.FS {
	sub, _ := fs.Sub(journalFS, "journal")
	return sub
}

// journalBook is JOURNAL.TXT, loaded once.
var journalBook = sync.OnceValue(func() *journal.Book {
	sub, _ := fs.Sub(journalFS, "journal")
	b, err := journal.Load(sub, "JOURNAL.TXT")
	if err != nil {
		panic(err) // built in: never
	}
	return b
})

// The reading box: its title, width and rows a page.
const (
	journalBox   = "Adventurer's Journal"
	journalWidth = 36
	journalRows  = 18
)

// journalEntries is the scripts' entries (scripts/names/entries.json) by id, read once.
var journalEntries = sync.OnceValue(func() map[eclscript.EntryID]eclscript.Entry {
	es, err := eclscript.ReadEntries(scriptData())
	if err != nil {
		panic(err) // built in: never
	}
	m := map[eclscript.EntryID]eclscript.Entry{}
	for _, en := range es {
		m[en.ID] = en
	}
	return m
})

// journalParas is title's text as paragraphs, pictures left out.
func journalParas(title string) []string {
	var out []string
	cur := ""
	for _, p := range journalBook().Text(title) {
		switch {
		case strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">"):
		case p == "":
			if cur != "" {
				out, cur = append(out, cur), ""
			}
		default:
			cur = strings.TrimSpace(cur + " " + p)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
