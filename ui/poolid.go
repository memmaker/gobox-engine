package ui

import (
	"image"
	"image/draw"

	"gobox/game"
)

// poolID is the pool portrait of head and body names (nil if either is
// missing): the head over the body, as the creation sheet stacks its Head
// and Body boxes (88x40 over 88x48). Cached by the two pictures, so the
// same pair is the same image (hiFace's cache holds).
func poolID(kind, head, body string) image.Image {
	hi, bi := poolPicture(kind, head), poolPicture(kind, body)
	if hi == nil || bi == nil {
		return nil
	}
	k := [2]image.Image{hi, bi}
	if out, ok := poolIDs[k]; ok {
		return out
	}
	hb, bb := hi.Bounds(), bi.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, max(hb.Dx(), bb.Dx()), hb.Dy()+bb.Dy()))
	draw.Draw(out, image.Rect(0, 0, hb.Dx(), hb.Dy()), hi, hb.Min, draw.Src)
	draw.Draw(out, image.Rect(0, hb.Dy(), bb.Dx(), hb.Dy()+bb.Dy()), bi, bb.Min, draw.Src)
	poolIDs[k] = out
	return out
}

// poolIDs are poolID's pictures by their head and body.
var poolIDs = map[[2]image.Image]image.Image{}

// withPoolPortrait is pics with the pool head and body of the creation draft in the sheet's Head and Body
// boxes, in place of the game's pictures there (pics itself when the draft names none).
func withPoolPortrait(pics []Picture, cr *FrontCreation, own game.Member) []Picture {
	m := subject(own)
	if cr == nil || m.PortraitHeadID == "" || m.PortraitBodyID == "" {
		return pics
	}
	head, body := poolPicture("portraits", m.PortraitHeadID), poolPicture("portraits", m.PortraitBodyID)
	if head == nil || body == nil {
		return pics
	}
	var keep []Picture
	port := cr.Head.Pixels().Union(cr.Body.Pixels())
	if own.PortraitHeadID == "" && pictured(pics, port) { // a default fills an empty box only
		return pics
	}
	for _, pc := range pics {
		if !pc.Fill && pc.At.Overlaps(port) {
			continue
		}
		keep = append(keep, pc)
	}
	for _, p := range []struct {
		im   image.Image
		box  CellRect
		name string
	}{{head, cr.Head, m.PortraitHeadID}, {body, cr.Body, m.PortraitBodyID}} {
		at := port.Min.Add(p.box.Pixels().Min.Sub(port.Min).Div(max(1, cr.Scale))) // at 1x from the portrait's corner: scaled scales it
		keep = append(keep, Picture{Name: p.name, Img: p.im, At: image.Rectangle{at, at.Add(p.im.Bounds().Size())}})
	}
	return keep
}
