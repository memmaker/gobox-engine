// Journal (journal.go) is the printed handouts a game only names ("SEE JOURNAL
// ENTRY 26"): their texts for the Read pager (State.Read; ui lays them out).
// The scripts name each through Journal(entry); nothing matches printed text.
//
// Text files: "[Title]" starts a text (the title is the reference's name),
// a blank line a paragraph, a line "<FILE.PNG>" a picture from the same
// directory, "#" lines are comments.
package game

import (
	"image"
	_ "image/png"
	"io/fs"
	"regexp"
	"strings"
	"sync"
)

// JournalBook is the handout texts and their pictures.
type JournalBook struct {
	fsys  fs.FS
	texts map[string][]string
	order []string // the titles in file order
	mu    sync.Mutex
	pics  map[string]image.Image
}

// softBreak is a word the printed book broke at its line end ("men- ace"),
// kept in the transcription: the text is reflowed, so the halves join.
var softBreak = regexp.MustCompile(`([a-z])- ([a-z])`)

// LoadJournal reads files of fsys into a Book.
func LoadJournal(fsys fs.FS, files ...string) (*JournalBook, error) {
	b := &JournalBook{fsys: fsys, texts: map[string][]string{}, pics: map[string]image.Image{}}
	for _, f := range files {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		var title string
		for _, l := range strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n") {
			l = strings.TrimSpace(l)
			switch {
			case strings.HasPrefix(l, "#"):
			case strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]"):
				title = l[1 : len(l)-1]
				if _, ok := b.texts[title]; !ok {
					b.order = append(b.order, title)
				}
				b.texts[title] = nil
			case title != "":
				b.texts[title] = append(b.texts[title], softBreak.ReplaceAllString(l, "$1$2"))
			}
		}
	}
	return b, nil
}

// Has reports whether the book holds title.
func (b *JournalBook) Has(title string) bool { _, ok := b.texts[title]; return ok }

// Titles is the book's titles in file order.
func (b *JournalBook) Titles() []string { return b.order }

// Text is title's lines as written: "" between paragraphs, "<FILE.PNG>" a picture.
func (b *JournalBook) Text(title string) []string { return b.texts[title] }

// Picture is the picture name, decoded once (nil: missing). It fits Data.Scans.
func (b *JournalBook) Picture(name string) image.Image {
	b.mu.Lock()
	defer b.mu.Unlock()
	if im, ok := b.pics[name]; ok {
		return im
	}
	var im image.Image
	if f, err := b.fsys.Open(name); err == nil {
		im, _, _ = image.Decode(f)
		f.Close()
	}
	b.pics[name] = im
	return im
}

// Texts is the texts titled titles (those the book holds) as a Read shows
// them; voice (nil: none) gives each paragraph's recording key.
func (b *JournalBook) Texts(titles []string, voice func(para string) string) []ReadText {
	var out []ReadText
	for _, t := range titles {
		paras, ok := b.texts[t]
		if !ok {
			continue
		}
		rt := ReadText{Title: t}
		for _, p := range paras {
			switch {
			case strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">"):
				rt.Paras = append(rt.Paras, ReadPara{Picture: p[1 : len(p)-1]})
			case p != "" && voice != nil:
				rt.Paras = append(rt.Paras, ReadPara{Text: p, Voice: voice(p)})
			default:
				rt.Paras = append(rt.Paras, ReadPara{Text: p})
			}
		}
		out = append(out, rt)
	}
	return out
}
