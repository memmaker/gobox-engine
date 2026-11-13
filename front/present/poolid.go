package present

import (
	"image"
	"image/draw"

	"gobox/front"
	"gobox/game"
)

// poolID is the pool portrait of head and body names (nil if either is
// missing): the head over the body, as the creation sheet stacks its Head
// and Body boxes (88x40 over 88x48).
func poolID(kind, head, body string) image.Image {
	hi, bi := poolPicture(kind, head), poolPicture(kind, body)
	if hi == nil || bi == nil {
		return nil
	}
	hb, bb := hi.Bounds(), bi.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, max(hb.Dx(), bb.Dx()), hb.Dy()+bb.Dy()))
	draw.Draw(out, image.Rect(0, 0, hb.Dx(), hb.Dy()), hi, hb.Min, draw.Src)
	draw.Draw(out, image.Rect(0, hb.Dy(), bb.Dx(), hb.Dy()+bb.Dy()), bi, bb.Min, draw.Src)
	return out
}

// withPoolPortrait is pics with the pool head and body of the creation draft in the sheet's Head and Body
// boxes, in place of the game's pictures there (pics itself when the draft names none).
func withPoolPortrait(pics []front.Picture, cr *front.Creation, own game.Member) []front.Picture {
	m := subject(own)
	if cr == nil || m.PortraitHeadID == "" || m.PortraitBodyID == "" {
		return pics
	}
	head, body := poolPicture("portraits", m.PortraitHeadID), poolPicture("portraits", m.PortraitBodyID)
	if head == nil || body == nil {
		return pics
	}
	var keep []front.Picture
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
		box  front.CellRect
		name string
	}{{head, cr.Head, m.PortraitHeadID}, {body, cr.Body, m.PortraitBodyID}} {
		at := p.box.Pixels().Min
		keep = append(keep, front.Picture{Name: p.name, Img: p.im, At: image.Rectangle{at, at.Add(p.im.Bounds().Size())}})
	}
	return keep
}
