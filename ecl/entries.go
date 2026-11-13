package ecl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
)

// EntryID names a long text of the adventure's text data, the operand of
// the pager command (Journal).
type EntryID int32

// Entry is one long text: a journal entry, a tavern tale, a history
// chapter. Book and Title find it in the adventure's texts; Voiced: each
// paragraph has its recording.
type Entry struct {
	ID     EntryID
	Name   string // its value name in the scripts (Journal(Name))
	Book   string // the adventure's text file
	Title  string // the text's title in it
	Box    string // the pager's title
	Voiced bool
}

// EntriesFile is the adventure's long texts, in its script data.
const EntriesFile = "names/entries.json"

// ReadEntries reads EntriesFile; none when the adventure has none.
func ReadEntries(data fs.FS) ([]Entry, error) {
	b, err := fs.ReadFile(data, EntriesFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var es []Entry
	if err := json.Unmarshal(b, &es); err != nil {
		return nil, fmt.Errorf("%s: %w", EntriesFile, err)
	}
	seen := map[EntryID]bool{}
	for _, e := range es {
		if seen[e.ID] {
			return nil, fmt.Errorf("%s: entry %d twice", EntriesFile, e.ID)
		}
		seen[e.ID] = true
	}
	return es, nil
}

// checkOperands is the load-time check of a parsed script against its data:
// every immediate fits its encoding (a byte operand is read as its low byte)
// and every Journal names an entry of EntriesFile.
func checkOperands(s *Script, data fs.FS) error {
	es, err := ReadEntries(data)
	if err != nil {
		return err
	}
	ids := map[EntryID]bool{}
	for _, e := range es {
		ids[e.ID] = true
	}
	for _, in := range s.Code {
		for _, a := range in.Args {
			if a.Type == ArgByte && !a.Label && (a.N < 0 || a.N > 0xFF) {
				return fmt.Errorf("%d: byte operand %d out of range", in.Line, a.N)
			}
		}
		if in.Op == OpJournal && len(in.Args) == 1 && (in.Args[0].Type == ArgByte || in.Args[0].Type == ArgWord) && !ids[EntryID(in.Args[0].N)] {
			return fmt.Errorf("%d: Journal(%d): no such entry in %s", in.Line, in.Args[0].N, EntriesFile)
		}
	}
	return nil
}
