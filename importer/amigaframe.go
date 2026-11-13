package importer

import "gobox/game"

// The Amiga releases' window frames (game.AmigaFrame): each game's frame
// screens as its program blits them, and the border cells read off them.
// Positions are the screen's pixels: the program passes 8000+x/2 (and
// 8000+y/2; coordinates above 6000 are in 2-pixel units from 8000, Curse
// F_3e520) and each piece's header hot spot is subtracted.

const (
	frameCols  = 40 // the screen's cells across
	frameRows  = 25 // and down
	frameCell  = 8  // a cell's side, in pixels
	ruleRow    = 16 // the row of the rule the frame screens draw under the 3D view
	dividerCol = 16 // the column of the divider right of the 3D view
	// dividerRows is the divider's length: rows 0..16, the rule's row the last
	dividerRows = ruleRow + 1
)

// frameRow is the cells c0..c1 of row row of screen s.
func frameRow(s, row, c0, c1 int) []game.FrameCell {
	var out []game.FrameCell
	for c := c0; c <= c1; c++ {
		out = append(out, game.FrameCell{Screen: s, Col: c, Row: row})
	}
	return out
}

// frameCol is the cells r0..r1 of column col of screen s.
func frameCol(s, col, r0, r1 int) []game.FrameCell {
	var out []game.FrameCell
	for r := r0; r <= r1; r++ {
		out = append(out, game.FrameCell{Screen: s, Col: col, Row: r})
	}
	return out
}

// frameEdges are the edges every frame has: the screen edges of the full
// frame (screen full), the rule of screen rule (ruleRow) and the divider of
// screen main (dividerCol, rows 0..ruleRow).
func frameEdges(full, rule, main int) map[string][]game.FrameCell {
	return map[string][]game.FrameCell{
		"top":     frameRow(full, 0, 0, frameCols-1),
		"bottom":  frameRow(full, frameRows-1, 0, frameCols-1),
		"left":    frameCol(full, 0, 0, frameRows-1),
		"right":   frameCol(full, frameCols-1, 0, frameRows-1),
		"rule":    frameRow(rule, ruleRow, 0, frameCols-1),
		"divider": frameCol(main, dividerCol, 0, dividerRows-1),
	}
}

// framePieces are the ribbon's pieces of screen full: its corners, then a
// top, bottom, left and right cell next to the top-left corner.
func framePieces(full int) [8]game.FrameCell {
	at := func(c, r int) game.FrameCell { return game.FrameCell{Screen: full, Col: c, Row: r} }
	last, bottom := frameCols-1, frameRows-1
	return [8]game.FrameCell{at(0, 0), at(last, 0), at(0, bottom), at(last, bottom), at(1, 0), at(1, bottom), at(0, 1), at(last, 1)}
}
