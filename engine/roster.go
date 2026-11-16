package engine

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"gobox/adventure"
	"gobox/rules"
)

// The roster's RosterIDs (core.RosterID, approved 2026-11-11): each roster
// .cha holds its own number in the trailer ("rosterid"), never derived from
// the name; charlist.txt keeps the original's format. rosternext holds the
// next number to give. A .cha from before (no number) gets one when the
// roster loads and is rewritten once (migrateRoster).

const rosterNextFile = "rosternext"

// rosterFile is the .cha of roster entry name (charlist.txt's, "* " marks it joined).
func (e *Engine) rosterFile(name string) string {
	return e.savePath(adventure.MakeFileName(strings.TrimPrefix(name, "* ")) + ".cha")
}

// newRosterID gives the next roster number and counts it in rosternext.
// The id is used even when the counter could not be kept (the error is returned).
func (e *Engine) newRosterID() (rules.RosterID, error) {
	next := e.rosterNext()
	return next, e.setRosterNext(next + 1)
}

// rosterNext is rosternext's number (1 when there is none yet).
func (e *Engine) rosterNext() rules.RosterID {
	if b, err := os.ReadFile(e.savePath(rosterNextFile)); err == nil {
		if n, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 32); err == nil && n > 0 {
			return rules.RosterID(n)
		}
	}
	return 1
}

func (e *Engine) setRosterNext(n rules.RosterID) error {
	return e.writeFile(e.savePath(rosterNextFile), []byte(strconv.FormatUint(uint64(n), 10)))
}

// migrateRoster numbers the roster's .cha files without one: next is past
// the highest number held (and rosternext), each is rewritten once with it.
func (e *Engine) migrateRoster(names []string) {
	old := map[string][]byte{} // path -> its bytes
	most := rules.NoRoster
	for _, n := range names {
		p := e.rosterFile(n)
		b, err := os.ReadFile(p)
		if err != nil || !e.fileFormat().IsRecord(b) {
			continue
		}
		if id := e.fileFormat().RosterOf(b); id == rules.NoRoster {
			old[p] = b
		} else {
			most = max(most, id)
		}
	}
	if len(old) == 0 {
		return
	}
	if e.rosterNext() <= most {
		e.setRosterNext(most + 1) // failures show "Cannot write"
	}
	for _, n := range names { // the list's order, so the numbers follow it
		p := e.rosterFile(n)
		b, ok := old[p]
		if !ok {
			continue
		}
		delete(old, p)
		id, err := e.newRosterID()
		if err != nil {
			continue // the counter is not kept (it showed "Cannot write"): no number is handed out
		}
		e.writeFile(p, e.fileFormat().WithRoster(b, id)) // failures show "Cannot write"; BuildCharList has no error channel
	}
}

// joinRoster puts roster character id in the party (the Add command's join).
func (e *Engine) joinRoster(id rules.RosterID) (rules.CharID, error) {
	for _, n := range e.BuildCharList() {
		b, err := os.ReadFile(e.rosterFile(n))
		if err != nil || e.fileFormat().RosterOf(b) != id {
			continue
		}
		c, ok := e.LoadCharacter(strings.TrimPrefix(n, "* "), false) // ovr18:0AEC
		if !ok {
			return rules.NoChar, errors.New("roster character unreadable")
		}
		e.join(c) // ovr18:1446
		return e.partyIDs()[e.partySize()-1], nil
	}
	return rules.NoChar, errors.New("no such roster character")
}
