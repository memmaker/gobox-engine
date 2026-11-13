package games

import (
	"sync"

	"gobox/game/journal"
)

var book = sync.OnceValue(func() *journal.Book {
	b, err := journal.Load(embeddedJournal("sotsb"), "JOURNAL.TXT")
	if err != nil {
		panic(err) // built in: never
	}
	return b
})
