// pictures.go is an adventure's picture table (data/pictures.json): each
// original DAX picture's entry in the shared portrait and icon collection,
// from the series' origins (origins/<series>.json, written by sprites'
// cmd/poolsprites).

package asset

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"gobox/rules"
)

// DAXRef is one original picture: its DAX file ("HEAD2.DAX") and block (and
// frame, for animated blocks; NoFrame otherwise), in its Platform's archives
// (whose names can repeat across platforms: CBODY.DAX).
type DAXRef struct {
	File     string
	Block    int
	Frame    int
	Platform Platform `json:",omitempty"`
}

// Platform is the version of a game an archive comes from.
type Platform uint8

const (
	PlatformDOS   Platform = iota // origins.json key <game>
	PlatformAmiga                 // origins.json key <game>-amiga
)

// originSuffix is each Platform's origins.json key suffix.
var originSuffix = map[Platform]string{PlatformDOS: "", PlatformAmiga: "-amiga"}

// PlatformID is game id's key for platform p (origins.json, the font pool's games.json).
func PlatformID(id string, p Platform) string { return id + originSuffix[p] }

// NoFrame is a DAXRef of a whole block.
const NoFrame = -1

// Pictures maps a game's original pictures to the collection entries
// (portraits head and body, combat icon head and body). It is written into
// the adventure as data/pictures.json; old saves and the creation grid
// convert their DAX numbers through it.
type Pictures map[DAXRef]rules.PicID

// PicturesFile is the adventure data path of the table.
const PicturesFile = "data/pictures.json"

type pictureEntry struct {
	DAXRef
	ID rules.PicID
}

// GamePictures is game id's table, built from its series' origins.json:
// the DOS archives' pictures (origins key game) and the Amiga's (game-amiga).
func GamePictures(game string) Pictures {
	p := Pictures{}
	for pl, suffix := range originSuffix {
		for file, blocks := range Origins()[game+suffix] {
			for k, name := range blocks {
				r := parseRef(file, k)
				r.Platform = pl
				p[r] = rules.PicID(name)
			}
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
func (p Pictures) ID(ref DAXRef) rules.PicID {
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
		if a.Platform != b.Platform {
			return a.Platform < b.Platform
		}
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

//go:embed origins/*.json
var originsFS embed.FS

// Origins is every series' origins merged (game ids are unique across
// series): game id, DAX file, "block" or "block.frame" -> its series' sprite name.
func Origins() map[string]map[string]map[string]string {
	o := map[string]map[string]map[string]string{}
	files, err := fs.Glob(originsFS, "origins/*.json")
	if err != nil {
		panic(err) // a malformed pattern, fixed at compile time
	}
	for _, f := range files {
		var m map[string]map[string]map[string]string
		b, err := originsFS.ReadFile(f)
		if err == nil {
			err = json.Unmarshal(b, &m)
		}
		if err != nil {
			panic(fmt.Sprintf("embedded %s: %v", f, err)) // embedded data, checked by TestPicturesRoundTrip
		}
		for k, v := range m {
			o[k] = v
		}
	}
	return o
}

// Pick is one picture of a character as a draft holds it: its pool pick,
// else the PicID of the game's own picture number (own).
func Pick(pool, own rules.PicID) rules.PicID {
	if pool != "" {
		return pool
	}
	return own
}

// Set is Pick's inverse into a character: a picture the game has a number
// for sets that number (kept when it already shows id) and clears the pool
// pick; any other is the pool pick, and the number becomes none so the
// record holds no second, stale source for the picture. of is a number's
// PicID, inv its inverse.
func Set(id rules.PicID, num *byte, none byte, pool *rules.PicID, of func(byte) rules.PicID, inv func(rules.PicID) (byte, bool)) {
	*pool = ""
	if of(*num) == id {
		return
	}
	if b, ok := inv(id); ok {
		*num = b
		return
	}
	if id != "" { // "": no pick, the number stays the picture
		*num = none
	}
	*pool = id
}
