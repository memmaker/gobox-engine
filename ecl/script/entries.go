package script

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
