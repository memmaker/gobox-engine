package engine

import (
	"errors"
	"gobox/game"
	"slices"
	"strconv"
	"strings"
)

// Notes (extra): the journal's second tab, published as Lore, newest
// first. The game adds the printed journal's texts the text window names
// (journal entries, proclamations, tavern tales: showEntry); the
// player adds what the text window shows (the text box's "> Notes"),
// kept with the map and square it was read on, and may remove those again
// (game.RemoveNote). Kept with the save (NOTES.GOB). A save from before the
// notes lists the journal texts it has heard (VOICED.GOB).

// noteRec is a note: a journal text by Title, or the player's (User) Text
// read at Place.
type noteRec struct {
	Title, Place string
	Text         []string
	User         bool
}

// journalTexts are JOURNAL.TXT's texts by title, paragraphs joined and
// pictures left out.
func (e *Engine) journalTexts() map[string][]string {
	if e.jtexts != nil {
		return e.jtexts
	}
	out := map[string][]string{}
	j := e.book("JOURNAL.TXT")
	if j == nil {
		return out // the game has no journal texts
	}
	for _, t := range j.Titles() {
		var paras []string
		for _, p := range j.Text(t) {
			switch {
			case strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">"):
			case p == "" || len(paras) == 0:
				paras = append(paras, p)
			default:
				paras[len(paras)-1] = strings.TrimSpace(paras[len(paras)-1] + " " + p)
			}
		}
		out[t] = slices.DeleteFunc(paras, func(p string) bool { return p == "" })
	}
	e.jtexts = out
	return out
}

// noteJournal adds the journal text titled t, once, and says so.
func (e *Engine) noteJournal(t string) {
	if _, ok := e.journalTexts()[t]; !ok || slices.ContainsFunc(e.ext.notes, func(n noteRec) bool { return !n.User && n.Title == t }) {
		return
	}
	e.ext.notes = append(e.ext.notes, noteRec{Title: t}) // gobox says it was added (State.Lore grew)
}

// paneNote is the text pane's text as one paragraph ("" none).
func (e *Engine) paneNote() string {
	var w []string
	for _, l := range e.pane().Lines {
		for _, r := range l {
			w = append(w, strings.Fields(r.Text)...)
		}
	}
	return strings.Join(w, " ")
}

// errNoted is a note already in the journal.
var errNoted = errors.New("Already in the journal")

// noteAdd adds s as the player's note, with the map and square, once.
func (e *Engine) noteAdd(s string) error {
	if s == "" {
		return &game.Refusal{} // nothing typed: refused, no toast
	}
	if slices.ContainsFunc(e.ext.notes, func(n noteRec) bool { return n.User && slices.Equal(n.Text, []string{s}) }) {
		return errNoted
	}
	m := e.mapName()
	e.ext.notes = append(e.ext.notes, noteRec{Title: m, Place: m + ", " + strconv.Itoa(int(e.posX)) + "," + strconv.Itoa(int(e.posY)), Text: []string{s}, User: true})
	return nil
}

// noteRemove removes the player's note id.
func (e *Engine) noteRemove(id int) {
	if id >= 0 && id < len(e.ext.notes) && e.ext.notes[id].User {
		e.ext.notes = slices.Delete(e.ext.notes, id, id+1)
	}
}

// loreView is the Notes, newest first, then the journal texts an older save
// has heard.
func (e *Engine) loreView() []Lore {
	var out []Lore
	for i, n := range slices.Backward(e.ext.notes) {
		if n.User {
			out = append(out, Lore{Title: n.Title, Place: n.Place, Text: n.Text, ID: i})
		} else {
			out = append(out, Lore{Title: n.Title, Place: "", Text: e.journalTexts()[n.Title], ID: -1})
		}
	}
	var old []string
	for t := range e.journalTexts() {
		if e.ext.voiced[strings.ReplaceAll(t, " ", "_")] && !slices.ContainsFunc(e.ext.notes, func(n noteRec) bool { return !n.User && n.Title == t }) {
			old = append(old, t)
		}
	}
	slices.Sort(old)
	for _, t := range old {
		out = append(out, Lore{Title: t, Place: "", Text: e.journalTexts()[t], ID: -1})
	}
	return out
}
