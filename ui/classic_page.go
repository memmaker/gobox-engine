package ui

import (
	"image"
	"path/filepath"

	"gobox/adventure"
	"image/color"

	"gobox/asset"
	"gobox/game"
)

// TextPage is the text page the state shows: the pane's lines, the prompt
// row, the lines over the view window (Scene.Placed) and the pictures,
// bottom first.
type TextPage struct {
	Pane   [][]game.TextRun
	Prompt []game.TextRun
	View   [][]game.TextRun
	Layers []Picture // placed: At on the 320x200 page
}

// drawnPicture is p where At says, as the page draws it (its DAX reference kept).
func drawnPicture(p game.Picture, at image.Rectangle) Picture {
	f := pix(p)
	f.File, f.Block, f.Frame, f.At = p.File, p.Block, p.Frame, at
	return f
}

// msgDelayTicks is the original's frame delay factor: a frame lasts its
// Delay times (message delay + 3); a MsgDelayed layer's FrameMs is stated
// at the default message delay, defaultMsgDelay (POR's finale).
const (
	msgDelayTicks   = 3
	defaultMsgDelay = 1
)

// scriptPictures are step st's layers as intro in shows them: a layer
// wanting more party members than in.Party left out, a MsgDelayed one's
// frames paced by in.MsgDelay.
func scriptPictures(st *adventure.SequenceStep, in *game.Intro) []Picture {
	var out []Picture
	for _, l := range st.Layers {
		if l.MinParty > in.Party {
			continue
		}
		f := scriptPicture(l)
		if l.MsgDelayed {
			f.FrameMs = l.FrameMs * (in.MsgDelay + msgDelayTicks) / (defaultMsgDelay + msgDelayTicks)
		}
		out = append(out, f)
	}
	return out
}

// scriptPicture is a script's layer l as the page draws it: its art (a DAX
// reference, or the PNG in the adventure's art/), placed in its cells.
func scriptPicture(l adventure.Layer) Picture {
	f := Picture{File: l.File, Block: l.Block, Frames: l.Frames, FrameMs: l.FrameMs}
	if b := artBlock(l.File, l.Block); b != nil && b.Kind == asset.Still && len(b.Frames) > 0 && !amigaShown() { // a DOS EGA still as its colour indices: the title's lightning recolours them (IntroUI.Flash)
		im := b.Frames[0].Image
		f.Pix, f.PixW = im.Data, im.Width*8
	}
	if l.Kind == adventure.ArtPNG {
		f = Picture{File: l.File, Img: ReadPNG(filepath.Join(advDir, adventure.ArtDir, l.File))} // File: its name, no DAX reference (Img set)
	}
	var size image.Point
	if l.Centred { // its own size: the picture's
		m := f.Img
		if m == nil {
			m = daxMasked(l.File, l.Block, 0, -1)
		}
		if m != nil {
			size = m.Bounds().Size()
		}
	}
	f.At = l.Placed(size.X, size.Y)
	return f
}

// paneRuns lays line l out from the line's start in ink: an unset ink is
// the body text's colour (gobox's, as every font colour) on no background.
func paneRuns(l []game.TextRun, ink [16]color.RGBA) []PaneRun {
	var r []PaneRun
	x := 0
	for _, t := range l {
		fg, bg := asset.EGARGBA(10), color.RGBA{}
		if n, ok := t.Fg.EGA(); ok {
			fg = ink[n]
		}
		if n, ok := t.Bg.EGA(); ok {
			bg = ink[n]
		}
		r = append(r, PaneRun{X: x, Text: t.Text, Fg: fg, Bg: bg, Upcase: t.Upcase, Blink: t.Blink})
		x += len(t.Text)
	}
	return r
}

func paneLines(ls [][]game.TextRun, ink [16]color.RGBA) [][]PaneRun {
	if ls == nil {
		return nil
	}
	r := make([][]PaneRun, len(ls))
	for i, l := range ls {
		r[i] = paneRuns(l, ink)
	}
	return r
}

// story is a script portrait in the view window: the HEAD at cells 3,3, the
// BODY five rows below (the Gold Box's ShowPortrait); other pictures stack on.
func story(pics []game.Picture) []Picture {
	var r []Picture
	y := 24
	for i, pc := range pics {
		h := []int{40, 48}[min(i, 1)]
		f := pix(pc)
		f.File, f.Block, f.Frame = pc.File, pc.Block, pc.Frame
		f.At = image.Rect(24, y, 24+88, y+h)
		r = append(r, f)
		y += h
	}
	return r
}
