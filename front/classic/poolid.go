package classic

import (
	"image"
	"image/draw"
	"strings"

	"gobox/sprites"
)

// poolID is the image of an opaque pool id "A+B" (nil if it names nothing):
// a portrait "head_N+body_M" is the head over the body, an icon
// "combat_head_N+combat_body_M" the head laid on the body's top.
// ponytail: pool combat PNGs are not recoloured by palette index, add via origins when a game needs it.
func poolID(kind, id string) image.Image {
	h, b, ok := strings.Cut(id, "+")
	if !ok {
		return nil
	}
	hi, bi := sprites.Image("pool", kind, h), sprites.Image("pool", kind, b)
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
