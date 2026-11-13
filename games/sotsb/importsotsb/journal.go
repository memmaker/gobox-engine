package importsotsb

import (
	"embed"
	"io/fs"
	"sync"

	"gobox/game/journal"
)

//go:embed journal
var journalFS embed.FS

var book = sync.OnceValue(func() *journal.Book {
	sub, _ := fs.Sub(journalFS, "journal")
	b, err := journal.Load(sub, "JOURNAL.TXT")
	if err != nil {
		panic(err) // built in: never
	}
	return b
})
