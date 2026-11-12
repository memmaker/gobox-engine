package pool

// The clue book (extra): TSR's Pool of Radiance clue book (
// CLUEBOOK.TXT of the game's journal books, its maps CLUE*.PNG). Its chapter list and its reading are
// the presenter's (B while exploring); the engine gives the text.

// clueChapters are CLUEBOOK.TXT's titles in the book's order.
func (e *Engine) clueChapters() []string {
	if b := e.book("CLUEBOOK.TXT"); b != nil {
		return b.Titles()
	}
	return nil
}

// ClueRead is the clue book from chapter i to its end, for the pager
// (JournalPage).
func (e *Engine) ClueRead(i int) *Read {
	titles := e.clueChapters()
	i = min(max(i, 0), len(titles)-1)
	return &Read{Title: "Clue Book", Lines: e.journalLayout("CLUEBOOK.TXT", titles[i:], false)}
}
