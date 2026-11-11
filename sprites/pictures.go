package sprites

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"gobox/core"
)

// DAXRef is one original picture: its DAX file ("HEAD2.DAX") and block (and
// frame, for animated blocks; NoFrame otherwise).
type DAXRef struct {
	File  string
	Block int
	Frame int
}

// NoFrame is a DAXRef of a whole block.
const NoFrame = -1

// Pictures maps a game's original pictures to the collection entries
// (portraits head and body, combat icon head and body). It is written into
// the adventure as data/pictures.json; old saves and the creation grid
// convert their DAX numbers through it.
type Pictures map[DAXRef]core.PicID

// PicturesFile is the adventure data path of the table.
const PicturesFile = "data/pictures.json"

type pictureEntry struct {
	DAXRef
	ID core.PicID
}

// GamePictures is game id's table, built from its series' origins.json.
func GamePictures(game string) Pictures {
	p := Pictures{}
	for file, blocks := range Origins()[game] {
		for key, name := range blocks {
			p[parseRef(file, key)] = core.PicID(name)
		}
	}
	return p
}

func parseRef(file, key string) DAXRef {
	r := DAXRef{File: file, Frame: NoFrame}
	for i := range key {
		if key[i] == '.' {
			r.Block, _ = strconv.Atoi(key[:i])
			r.Frame, _ = strconv.Atoi(key[i+1:])
			return r
		}
	}
	r.Block, _ = strconv.Atoi(key)
	return r
}

// ID is ref's entry: the frame's, else the whole block's ("" if none).
func (p Pictures) ID(ref DAXRef) core.PicID {
	if id, ok := p[ref]; ok {
		return id
	}
	ref.Frame = NoFrame
	return p[ref]
}

// MarshalJSON writes the table as a list (JSON keys are strings only).
func (p Pictures) MarshalJSON() ([]byte, error) {
	es := make([]pictureEntry, 0, len(p))
	for r, id := range p {
		es = append(es, pictureEntry{r, id})
	}
	sortEntries(es)
	return json.Marshal(es)
}

// UnmarshalJSON reads MarshalJSON's list.
func (p *Pictures) UnmarshalJSON(b []byte) error {
	var es []pictureEntry
	if err := json.Unmarshal(b, &es); err != nil {
		return err
	}
	*p = make(Pictures, len(es))
	for _, e := range es {
		(*p)[e.DAXRef] = e.ID
	}
	return nil
}

// ReadPictures reads an adventure folder's PicturesFile.
func ReadPictures(adventure fs.FS) (Pictures, error) {
	b, err := fs.ReadFile(adventure, PicturesFile)
	if err != nil {
		return nil, err
	}
	var p Pictures
	return p, json.Unmarshal(b, &p)
}

func sortEntries(es []pictureEntry) {
	sort.Slice(es, func(i, j int) bool {
		a, b := es[i].DAXRef, es[j].DAXRef
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Block != b.Block {
			return a.Block < b.Block
		}
		return a.Frame < b.Frame
	})
}

// WritePictures writes game id's table into the adventure folder dst.
func WritePictures(game, dst string) error {
	b, err := json.MarshalIndent(GamePictures(game), "", " ")
	if err != nil {
		return err
	}
	path := filepath.Join(dst, PicturesFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
