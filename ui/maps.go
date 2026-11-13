package ui

import (
	"fmt"
	"image"
	"slices"
	"strconv"
	"strings"

	"gobox/game"
)

// ahead is the step towards each facing (N E S W).
var ahead = [4]image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// labelKinds are the label kinds by game.Label.Kind: their colours are
// Data.LabelInks[kind]; the player's ("") are 0.
var labelKinds = map[string]int{"": 0, "shop": 1, "temple": 2, "inn": 3, "civic": 4}

// label is a label on a square of the map, the map's or the player's.
type label struct {
	game.Label
	mine bool
}

// labels are a's labels, the map's and the player's, square by square
// (row by row): the player's are a.Marked with a.Notes.
func labels(a game.Automap) []label {
	var l []label
	for _, b := range a.Labels {
		l = append(l, label{b, false})
	}
	for i, at := range a.Marked {
		if i < len(a.Notes) {
			k := ""
			if i < len(a.NoteKinds) {
				k = a.NoteKinds[i]
			}
			l = append(l, label{game.Label{At: at, Text: a.Notes[i], Kind: k}, true})
		}
	}
	slices.SortStableFunc(l, func(p, q label) int { return (p.At.Y<<8 | p.At.X) - (q.At.Y<<8 | q.At.X) })
	return l
}

// letters gives each label text its legend letter: A..Z, then a..z, in the
// texts' order ignoring case.
func letters(ls []label) (texts []string, l map[string]byte) {
	for _, b := range ls {
		texts = append(texts, b.Text)
	}
	slices.SortFunc(texts, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	texts = slices.Compact(texts)
	l = map[string]byte{}
	for i, t := range texts {
		l[t] = byte('A' + i)
		if i >= 26 {
			l[t] = byte('a' + i - 26)
		}
	}
	return texts, l
}

// Map is m (the game's) with what s.Map gives in its place: how far each
// square is seen, the party's square and step ahead, the walls and doors
// and the labels' kinds and letters. nil where m is.
func Map(m *FrontMapSquares, a game.Automap, where game.Where) *FrontMapSquares {
	if m == nil || a.Size != image.Pt(16, 16) {
		return m
	}
	c := FrontMapSquares{Party: image.Pt(where.X, where.Y), Ahead: ahead[where.Facing/2&3]}
	on := image.Rect(0, 0, 16, 16)
	for _, p := range a.Known {
		if p.In(on) {
			c.Squares[p.Y][p.X].Seen = FrontSeenFar
		}
	}
	for _, p := range a.Walked {
		if p.In(on) {
			c.Squares[p.Y][p.X].Seen = FrontVisited
		}
	}
	set := func(ws []game.Wall, k func(game.Wall) MapWall) {
		for _, w := range ws {
			if w.At.In(on) {
				c.Squares[w.At.Y][w.At.X].Walls[w.Side/2&3] = k(w)
			}
		}
	}
	set(a.Walls, func(game.Wall) MapWall { return FrontWall })
	set(a.Doors, func(w game.Wall) MapWall {
		if w.State == game.DoorLocked || w.State == game.DoorWizardLocked {
			return FrontLockedDoor
		}
		return FrontDoor
	})
	ls := labels(a)
	_, l := letters(ls)
	for _, b := range ls {
		if b.At.In(on) {
			q := &c.Squares[b.At.Y][b.At.X]
			q.Kind, q.Letter = labelKinds[b.Kind], l[b.Text]
		}
	}
	return &c
}

// pix is p (the game's) as the front's pixels, or as its reference without any.
func pix(p game.Picture) Picture {
	if p.W == 0 || p.Pixels == nil { // a ref (DAX File, Block, Frame, or a pool Name) the presenter resolves
		return Picture{File: p.File, Block: p.Block, Frame: p.Frame, Name: p.Name}
	}
	return PixelsOf(p)
}

// AreaMap is m (the game's) with the overland's squares and marker from
// s.Map.Terrain and art (Data.Art: "terrain/<id>", "marker/<facing>"), and
// the squares seen from s.Map.Known. m where it is nil, s has no overland
// or art has none of it.
func AreaMap(m *FrontAreaMap, a game.Automap, where game.Where, art map[string]game.Picture) *FrontAreaMap {
	if m == nil || len(a.Terrain) == 0 || len(art) == 0 || len(a.Terrain) != a.Size.X*a.Size.Y {
		return m
	}
	c := *m
	c.SquarePix = make([][]Picture, a.Size.Y)
	for y := range c.SquarePix {
		c.SquarePix[y] = make([]Picture, a.Size.X)
		for x := range c.SquarePix[y] {
			p := art["terrain/"+strconv.Itoa(a.Terrain[y*a.Size.X+x])]
			c.SquarePix[y][x] = pix(p)
			c.Tile = max(c.Tile, p.W)
		}
	}
	c.Party = image.Pt(where.X, where.Y)
	mk := a.Marker
	if mk == "" {
		mk = "marker"
	}
	c.MarkerPix = pix(art[mk+"/"+strconv.Itoa(where.Facing)])
	c.Seen = nil
	if len(a.Known) < len(a.Terrain) { // else every square is seen
		c.Seen = make([][]bool, a.Size.Y)
		for y := range c.Seen {
			c.Seen[y] = make([]bool, a.Size.X)
		}
		for _, p := range a.Known {
			if p.In(image.Rectangle{Max: a.Size}) {
				c.Seen[p.Y][p.X] = true
			}
		}
	}
	return &c
}

// FullMap is f (the game's) with its map (Map), name, legend, and the
// cursor's square and label from s. nil where f is.
func FullMap(f *FrontFullMap, a game.Automap, where game.Where) *FrontFullMap {
	if f == nil || a.Size != image.Pt(16, 16) {
		return nil // no 16x16 map up: nothing to draw (a FullMap without Map crashed the presenter)
	}
	c := *f
	if c.Map == nil { // the front's own M asks with no squares yet (Map keeps nil as nil)
		c.Map = &FrontMapSquares{}
	}
	c.Name, c.Map, c.Labelled, c.Mine, c.Legend = a.Area, Map(c.Map, a, where), false, false, nil
	c.Info = fmt.Sprintf("%d,%d", f.Cursor.X, f.Cursor.Y)
	ls := labels(a)
	texts, l := letters(ls)
	for _, b := range ls {
		if b.At == f.Cursor {
			c.Info += " " + b.Text
			c.Labelled, c.Mine = true, b.mine
		}
	}
	for _, t := range texts {
		i := slices.IndexFunc(ls, func(b label) bool { return b.Text == t })
		c.Legend = append(c.Legend, MapLabel{Letter: l[t], Text: t, Kind: labelKinds[ls[i].Kind], First: ls[i].At})
	}
	return &c
}

// automap is the automap of the state up (zero with no game).
func (p *Presenter) automap() game.Automap {
	a, _ := p.minimap()
	return a
}
