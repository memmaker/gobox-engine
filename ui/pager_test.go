package ui

import (
	"testing"
)

type src struct{ pages, maxOff int }

func (s src) Page(p, off int) (int, int, int, bool) {
	return min(p, s.pages-1), s.pages, s.maxOff, true
}

func TestPager(t *testing.T) {
	var pg Pager
	s := src{3, 40}
	for _, c := range []struct {
		a      Act
		p, off int
		closed bool
	}{{Enter, 0, 16, false}, {Down, 0, 32, false}, {Down, 0, 40, false}, {Down, 1, 0, false},
		{Up, 0, 0, false}, {Up, 0, 0, false}, {Esc, 0, 0, true}} {
		closed, _ := pg.Key(c.a, s)
		if closed != c.closed || pg.P != c.p || pg.Off != c.off {
			t.Fatalf("%v: %+v closed %v", c.a, pg, closed)
		}
	}
	pg = Pager{P: 2}
	if closed, _ := pg.Key(Enter, src{3, 0}); !closed {
		t.Fatal("past the last page")
	}
}

func TestHelpPages(t *testing.T) {
	var l [][2]string
	for range 3 {
		l = append(l, [2]string{"Head", ""}, [2]string{"k", "d"}, [2]string{"k", "d"}, [2]string{"", ""})
	}
	p := HelpPages(l, 5)
	if len(p) != 3 || p[1][0][0] != "Head" {
		t.Fatalf("%q", p)
	}
	if len(HelpPages(nil, 5)) != 1 {
		t.Fatal("no page")
	}
}

type tinyScroll struct{}

func (tinyScroll) Page(p, off int) (int, int, int, bool) { return p, 2, 2, true }

func TestPagerTinyScrollTurns(t *testing.T) {
	var pg Pager
	pg.Turn(tinyScroll{}, 1)
	if pg.P != 1 || pg.Off != 0 {
		t.Fatalf("pager %+v", pg)
	}
}
