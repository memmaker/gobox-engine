package classic

import (
	"image"
	"image/draw"

	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// poolID is the pool portrait of head and body names (nil if either is
// missing): the head over the body, as the creation sheet stacks its Head
// and Body boxes (88x40 over 88x48).
func poolID(kind, head, body string) image.Image {
	hi, bi := sprites.Image(Series, kind, head), sprites.Image(Series, kind, body)
	if hi == nil || bi == nil {
		return nil
	}
	hb, bb := hi.Bounds(), bi.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, max(hb.Dx(), bb.Dx()), hb.Dy()+bb.Dy()))
	draw.Draw(out, image.Rect(0, 0, hb.Dx(), hb.Dy()), hi, hb.Min, draw.Src)
	draw.Draw(out, image.Rect(0, hb.Dy(), bb.Dx(), hb.Dy()+bb.Dy()), bi, bb.Min, draw.Src)
	return out
}

// PoolPortrait is v with the pool head and body of the creation draft in the sheet's Head and Body
// boxes, in place of the game's pictures there (v itself when the draft names none).
func PoolPortrait(v front.View, s game.State) front.View {
	cr, m := v.Creation, subject(s)
	if cr == nil || m.PortraitHeadID == "" || m.PortraitBodyID == "" {
		return v
	}
	head, body := sprites.Image(Series, "portraits", m.PortraitHeadID), sprites.Image(Series, "portraits", m.PortraitBodyID)
	if head == nil || body == nil {
		return v
	}
	var keep []front.Picture
	port := cr.Head.Pixels().Union(cr.Body.Pixels())
	for _, pc := range v.Pictures {
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
	v.Pictures = keep
	return v
}
