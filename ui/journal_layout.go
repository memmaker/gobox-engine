package ui

import (
	"image"
	"strings"

	"gobox/game"
)

// layout lays out the texts of a Read for the pager, jrWidth characters
// wide: "\x00" marks a title (a blank line under it), "\x01" a picture and
// "\x02" the rows it covers, "\x03key\x03" a voiced paragraph's first line.
// scan gives the pictures (one it lacks is left out).
func (p pageCut) layout(texts []game.ReadText, scan func(string) image.Image) []string {
	var lines []string
	for _, t := range texts {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "\x00"+t.Title, "")
		for _, para := range t.Paras {
			from := len(lines)
			switch {
			case para.Picture != "":
				if scan == nil || scan(para.Picture) == nil {
					continue
				}
				lines = append(lines, "\x01"+para.Picture)
				_, h := pictureSize(scan(para.Picture))
				for range min((h+7)/8, p.jrRows()) - 1 {
					lines = append(lines, "\x02")
				}
			case para.Text == "":
				lines = append(lines, "")
			case len(para.Text) <= jrWidth: // as written: keeps the spacing of tables
				lines = append(lines, para.Text)
			default:
				lines = append(lines, wrapText(para.Text, jrWidth)...)
			}
			if para.Voice != "" {
				lines[from] = "\x03" + para.Voice + "\x03" + lines[from]
			}
		}
	}
	return lines
}

// wrapText breaks s into lines of at most w characters at spaces.
func wrapText(s string, w int) []string {
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
