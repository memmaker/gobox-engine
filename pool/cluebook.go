package pool

// The clue book (extra): TSR's Pool of Radiance clue book (journal/
// CLUEBOOK.TXT, its maps CLUE*.PNG). Its chapter list and its reading are
// the presenter's (B while exploring); the engine gives the text.

// clueChapters are CLUEBOOK.TXT's titles in the book's order.
func clueChapters() []string { return book("CLUEBOOK.TXT").Titles() }

// ClueRead is the clue book from chapter i to its end, for the pager
// (JournalPage).
func ClueRead(i int) *Read {
	titles := clueChapters()
	i = min(max(i, 0), len(titles)-1)
	return &Read{Title: "Clue Book", Lines: journalLayout("CLUEBOOK.TXT", titles[i:], false)}
}
