package classic

import (
	"image"
	"image/draw"

	"gobox/sprites"
)

// poolID is the pool picture of head and body names (nil if either is
// missing): a portrait is the head over the body, as the creation sheet
// stacks its Head and Body boxes (88x40 over 88x48), a combat icon the head
// laid on the body's top. kind is "portraits" or "combat".
// ponytail: pool combat PNGs are RGBA, so no recolouring by palette index, facing or pose yet.
func poolID(kind, head, body string) image.Image {
	hi, bi := sprites.Image("pool", kind, head), sprites.Image("pool", kind, body)
	if hi == nil || bi == nil {
		return nil
	}
	hb, bb := hi.Bounds(), bi.Bounds()
	if kind == "combat" {
		out := image.NewRGBA(image.Rect(0, 0, bb.Dx(), bb.Dy()))
		draw.Draw(out, out.Bounds(), bi, bb.Min, draw.Src)
		draw.Draw(out, image.Rect(0, 0, hb.Dx(), hb.Dy()), hi, hb.Min, draw.Over)
		return out
	}
	out := image.NewRGBA(image.Rect(0, 0, max(hb.Dx(), bb.Dx()), hb.Dy()+bb.Dy()))
	draw.Draw(out, image.Rect(0, 0, hb.Dx(), hb.Dy()), hi, hb.Min, draw.Src)
	draw.Draw(out, image.Rect(0, hb.Dy(), bb.Dx(), hb.Dy()+bb.Dy()), bi, bb.Min, draw.Src)
	return out
}
