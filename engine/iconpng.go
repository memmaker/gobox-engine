package engine

import (
	"fmt"
	"gobox/asset"
	"gobox/rules"
	"image"
)

// The combat icon from a PNG (extra): a file dropped on the icon editor,
// saved as NAME.icon.png beside the record. A square picture is both
// poses; one twice as wide as high is cut in the middle, the left half the
// ready pose, the right the action pose. The battle draws it in place of
// the built icon; the page's icons (the overview) stay as built.

// pngKey is a PNG's bytes as a map key: the decoded pictures are cached
// by content, so a changed PNG never shows a stale picture.
type pngKey string

// iconPoseKey is one cut of a PNG icon: its pose (0 ready, 1 action) and facing.
type iconPoseKey struct {
	png        pngKey
	pose, flip int
}

// pngImage is the picture of b, a PNG a character holds (checked when it
// came in: asset.ReadPNG, asset.IconShape); nil for none.
func (e *Engine) pngImage(b []byte) image.Image {
	if b == nil {
		return nil
	}
	if im, ok := e.pngs[pngKey(b)]; ok {
		return im
	}
	im, err := asset.DecodePNG(b)
	if err != nil {
		panic(fmt.Sprintf("a character's PNG was not checked when it came in: %v", err))
	}
	if e.pngs == nil {
		e.pngs = map[pngKey]image.Image{}
	}
	e.pngs[pngKey(b)] = im
	return im
}

// takeIconPNG makes the PNG b c's combat icon, or says why not (the icon
// editor, a PNG dropped on the creation sheet's icon).
func (e *Engine) takeIconPNG(cID rules.CharID, b []byte) bool {
	if !e.iconOK(b) {
		return false
	}
	p := e.iconPics(cID)
	p.PNG = b
	e.do(rules.SetIconPics{Who: cID, Pics: p})
	return true
}

// iconOK is whether b is a combat icon's PNG, or says why not.
func (e *Engine) iconOK(b []byte) bool {
	if err := asset.IconShape(b); err != nil {
		e.say(asset.ErrIconShape.Error())
		return false
	}
	return true
}

// iconPNG is c's PNG icon for pose (0 ready, 1 action), flipped to face
// left when flip; nil without a PNG. The cuts are kept: the battle's
// figures are cached by image.
func (e *Engine) iconPNG(cID rules.CharID, pose int, flip bool) image.Image {
	png := e.iconPics(cID).PNG
	im := e.pngImage(png)
	if im == nil {
		return nil
	}
	f := 0
	if flip {
		f = 1
	}
	if b := im.Bounds(); b.Dx() == b.Dy() { // one picture for both poses
		pose = 0
	}
	k := iconPoseKey{pngKey(png), pose, f}
	if p, ok := e.iconPoses[k]; ok {
		return p
	}
	if e.iconPoses == nil {
		e.iconPoses = map[iconPoseKey]image.Image{}
	}
	e.iconPoses[k] = asset.IconPose(im, pose, flip)
	return e.iconPoses[k]
}
