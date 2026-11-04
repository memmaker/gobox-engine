package classic

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"gobox/front"
	"gobox/game"
)

// ahead is the step towards each facing (N E S W).
var ahead = [4]image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// letters gives each label text its legend letter: A..Z, then a..z, in the
// texts' order ignoring case.
func letters(notes []string) map[string]byte {
	texts := slices.Clone(notes)
	slices.SortFunc(texts, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	texts = slices.Compact(texts)
	l := map[string]byte{}
	for i, t := range texts {
		l[t] = byte('A' + i)
		if i >= 26 {
			l[t] = byte('a' + i - 26)
		}
	}
	return l
}

// Map is m (the game's) with what s.Map gives in its place: the party's
// square and step ahead, the walls and doors (door Kind 2 and up: locked)
// and the labels' letters. How far each square is seen and the labels'
// kinds stay the game's: game.Automap has neither. nil where m is.
func Map(m *front.MapSquares, s game.State) *front.MapSquares {
	a := s.Map
	if m == nil || a.Size == (image.Point{}) {
		return m
	}
	c := *m
	c.Party, c.Ahead = image.Pt(s.Where.X, s.Where.Y), ahead[s.Where.Facing/2&3]
	for y := range c.Squares {
		for x := range c.Squares[y] {
			c.Squares[y][x].Walls, c.Squares[y][x].Letter = [4]front.MapWall{}, 0
		}
	}
	set := func(ws []game.Wall, k func(game.Wall) front.MapWall) {
		for _, w := range ws {
			if w.At.In(image.Rect(0, 0, 16, 16)) {
				c.Squares[w.At.Y][w.At.X].Walls[w.Side/2&3] = k(w)
			}
		}
	}
	set(a.Walls, func(game.Wall) front.MapWall { return front.Wall })
	set(a.Doors, func(w game.Wall) front.MapWall {
		if w.Kind >= 2 {
			return front.LockedDoor
		}
		return front.Door
	})
	l := letters(a.Notes)
	for i, at := range a.Marked {
		if i < len(a.Notes) && at.In(image.Rect(0, 0, 16, 16)) {
			c.Squares[at.Y][at.X].Letter = l[a.Notes[i]]
		}
	}
	return &c
}

// FullMap is f (the game's) with its map (Map), name, and the cursor's
// square and label from s. The legend's kinds and whether a label is the
// player's (Mine) stay the game's. nil where f is.
func FullMap(f *front.FullMap, s game.State) *front.FullMap {
	if f == nil || s.Map.Size == (image.Point{}) {
		return f
	}
	c := *f
	c.Name, c.Map, c.Labelled = s.Map.Area, Map(f.Map, s), false
	c.Info = fmt.Sprintf("%d,%d", f.Cursor.X, f.Cursor.Y)
	if i := slices.Index(s.Map.Marked, f.Cursor); i >= 0 && i < len(s.Map.Notes) {
		c.Info += " " + s.Map.Notes[i]
		c.Labelled = true
	}
	return &c
}
