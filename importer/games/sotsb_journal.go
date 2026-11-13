package games

import (
	"sync"

	"gobox/game"
)

var book = sync.OnceValue(func() *game.JournalBook {
	b, err := game.LoadJournal(embeddedJournal("sotsb"), "JOURNAL.TXT")
	if err != nil {
		panic(err) // built in: never
	}
	return b
})
