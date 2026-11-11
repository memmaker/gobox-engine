// Package journal is the printed handouts a game only names ("SEE JOURNAL
// ENTRY 26"): their texts, laid out for the Read pager (State.Read). The
// scripts name each through Journal(entry); nothing matches printed text.
//
// Text files: "[Title]" starts a text (the title is the reference's name),
// a blank line a paragraph, a line "<FILE.PNG>" a picture from the same
// directory, "#" lines are comments.
package journal

import (
	"image"
	_ "image/png"
	"io/fs"
	"regexp"
	"strings"
	"sync"
)

// Book is the handout texts and their pictures.
type Book struct {
	fsys  fs.FS
	texts map[string][]string
	order []string // the titles in file order
	mu    sync.Mutex
	pics  map[string]image.Image
	sizes map[string]image.Rectangle
}

// Load reads files of fsys into a Book.
// softBreak is a word the printed book broke at its line end ("men- ace"),
// kept in the transcription: the text is reflowed, so the halves join.
var softBreak = regexp.MustCompile(`([a-z])- ([a-z])`)

func Load(fsys fs.FS, files ...string) (*Book, error) {
	b := &Book{fsys: fsys, texts: map[string][]string{}, pics: map[string]image.Image{}, sizes: map[string]image.Rectangle{}}
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
func (b *Book) Has(title string) bool { _, ok := b.texts[title]; return ok }

// Titles is the book's titles in file order.
func (b *Book) Titles() []string { return b.order }

// Text is title's lines as written: "" between paragraphs, "<FILE.PNG>" a picture.
func (b *Book) Text(title string) []string { return b.texts[title] }

// Picture is the picture name, decoded once (nil: missing). It fits Data.Scans.
func (b *Book) Picture(name string) image.Image {
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

// bounds is a picture's size from its header alone (the layout needs no pixels).
func (b *Book) bounds(name string) image.Rectangle {
	b.mu.Lock()
	defer b.mu.Unlock()
	if r, ok := b.sizes[name]; ok {
		return r
	}
	var r image.Rectangle
	if f, err := b.fsys.Open(name); err == nil {
		if c, _, err := image.DecodeConfig(f); err == nil {
			r = image.Rect(0, 0, c.Width, c.Height)
		}
		f.Close()
	}
	b.sizes[name] = r
	return r
}

// PictureSize is the screen area a picture of size r covers on a page width
// characters wide: the scans are 300 dpi, shown at about 90 pixels an inch,
// at most the text width wide and, unless taller than wide (those scroll),
// 17 rows high.
func PictureSize(r image.Rectangle, width int) (w, h int) {
	s := min(0.3, float64(width*8)/float64(r.Dx()))
	if r.Dy() <= r.Dx() {
		s = min(s, 136/float64(r.Dy()))
	}
	return int(float64(r.Dx()) * s), int(float64(r.Dy()) * s)
}

// Layout lays out the texts titled titles for the pager, width characters
// wide and rows lines a page: "\x00" marks a title (a blank line under it),
// "\x01" a picture and "\x02" the rows it covers; voice (nil: none) gives
// each paragraph's recording key, marked "\x03key\x03" on its first line.
func (b *Book) Layout(titles []string, width, rows int, voice func(para string) string) []string {
	var lines []string
	for _, t := range titles {
		paras, ok := b.texts[t]
		if !ok {
			continue
		}
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "\x00"+t, "")
		for _, p := range paras {
			switch {
			case p == "":
				lines = append(lines, "")
			case strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">"):
				if r := b.bounds(p[1 : len(p)-1]); r != (image.Rectangle{}) {
					lines = append(lines, "\x01"+p[1:len(p)-1])
					_, h := PictureSize(r, width)
					for range min((h+7)/8, rows) - 1 {
						lines = append(lines, "\x02")
					}
				}
			case len(p) <= width: // as written: keeps the spacing of tables
				lines = append(lines, p)
			default:
				lines = append(lines, Wrap(p, width)...)
			}
			if voice != nil && p != "" {
				for i := len(lines) - 1; i >= 0 && lines[i] != ""; i-- {
					if i == 0 || lines[i-1] == "" {
						lines[i] = "\x03" + voice(p) + "\x03" + lines[i]
					}
				}
			}
		}
	}
	return lines
}

// Wrap breaks s into lines of at most w characters at spaces.
func Wrap(s string, w int) []string {
	var out []string
	line := ""
	for _, word := range strings.Fields(s) {
		for len(word) > w {
			out = append(out, word[:w])
			word = word[w:]
		}
		switch {
		case line == "":
			line = word
		case len(line)+1+len(word) <= w:
			line += " " + word
		default:
			out = append(out, line)
			line = word
		}
	}
	return append(out, line)
}
