package games

import (
	"embed"
	"io/fs"
	"sync"

	"gobox/game/journal"
)

//go:embed sotsb/journal
var sotsbJournalFS embed.FS

var book = sync.OnceValue(func() *journal.Book {
	sub, _ := fs.Sub(sotsbJournalFS, "sotsb/journal")
	b, err := journal.Load(sub, "JOURNAL.TXT")
	if err != nil {
		panic(err) // built in: never
	}
	return b
})
