package engine

import "strings"

// Punct is the set of characters that end a word for Flow.
func Punct(c byte) bool { return strings.IndexByte("!,-.:;?", c) >= 0 }

// Flow is the Gold Box DisplayTextBlock: it lays s out word wrapped in the box x1..x2, y1..y2 from the cursor
// (*col, *row): put gets each character's cell; when the box is full, full
// is called and the box starts over.
func Flow(s string, x1, y1, x2, y2 int, col, row *int, put func(x, y int, c byte), full func()) {
	n := min(len(s), 255)
	if n == 0 {
		return
	}
	at := func(k int) byte { return s[k-1] } // 1-based
	i := 1
	printSeg := func(j int) {
		for ; i <= j; i++ {
			put(*col, *row, at(i))
			*col++
		}
	}
	for {
		j := i
		for j < n && Punct(at(j)) {
			j++
		}
		for j < n && !Punct(at(j)) && at(j) != ' ' {
			j++
		}
		if at(j) != ' ' {
			for j+1 < n && Punct(at(j+1)) {
				j++
			}
		}
		if j-i+*col > x2 {
			*col = x1
			*row++
			for i < n && at(i) == ' ' {
				i++
			}
			if *row > y2 && i < n {
				*col, *row = x1, y1
				full()
				printSeg(j)
			}
		} else {
			printSeg(j)
		}
		if i > n {
			break
		}
	}
	if *col > x2 {
		*col = x1
		*row++
	}
}
