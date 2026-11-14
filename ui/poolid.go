package ui

import (
	"image"
	"image/draw"
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
