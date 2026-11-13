package engine

import (
	"bytes"
	"errors"
	"fmt"
	"gobox/core"
	"image"
	"image/png"
	"io/fs"
	"os"
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

// decodePNG is b as a picture, or why it is not a PNG.
func decodePNG(b []byte) (image.Image, error) {
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("not a PNG: %w", err)
	}
	return im, nil
}

// readPNG is the PNG file at path, its bytes checked to decode; nil, nil
// when there is no file.
func readPNG(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := decodePNG(b); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return b, nil
}

// pngImage is the picture of b, a PNG a character holds (checked when it
// came in: readPNG, iconShape); nil for none.
func (e *Engine) pngImage(b []byte) image.Image {
	if b == nil {
		return nil
	}
	if im, ok := e.pngs[pngKey(b)]; ok {
		return im
	}
	im, err := decodePNG(b)
	if err != nil {
		panic(fmt.Sprintf("a character's PNG was not checked when it came in: %v", err))
	}
	if e.pngs == nil {
		e.pngs = map[pngKey]image.Image{}
	}
	e.pngs[pngKey(b)] = im
	return im
}

// errIconShape: a PNG icon is square or twice as wide as high.
var errIconShape = errors.New("not a PNG, square or twice as wide as high")

// iconShape checks that b is a PNG a combat icon can be cut from.
func iconShape(b []byte) error {
	im, err := decodePNG(b)
	if err != nil {
		return err
	}
	if r := im.Bounds(); r.Dx() != r.Dy() && r.Dx() != 2*r.Dy() {
		return errIconShape
	}
	return nil
}

// takeIconPNG makes the PNG b c's combat icon, or says why not (the icon
// editor, a PNG dropped on the creation sheet's icon).
func (e *Engine) takeIconPNG(cID core.CharID, b []byte) bool {
	if !e.iconOK(b) {
		return false
	}
	p := e.iconPics(cID)
	p.PNG = b
	e.do(core.SetIconPics{Who: cID, Pics: p})
	return true
}

// iconOK is whether b is a combat icon's PNG, or says why not.
func (e *Engine) iconOK(b []byte) bool {
	if err := iconShape(b); err != nil {
		e.say(errIconShape.Error())
		return false
	}
	return true
}

// iconPNG is c's PNG icon for pose (0 ready, 1 action), flipped to face
// left when flip; nil without a PNG. The cuts are kept: the battle's
// figures are cached by image.
func (e *Engine) iconPNG(cID core.CharID, pose int, flip bool) image.Image {
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
	e.iconPoses[k] = iconPose(im, pose, flip)
	return e.iconPoses[k]
}

// iconPose cuts pose (0 left, 1 right; the same square when im is square)
// out of im, mirrored when flip.
func iconPose(im image.Image, pose int, flip bool) image.Image {
	b := im.Bounds()
	w := b.Dy()
	x0 := b.Min.X + pose*(b.Dx()-w)
	out := image.NewRGBA(image.Rect(0, 0, w, w))
	for y := 0; y < w; y++ {
		for x := 0; x < w; x++ {
			sx := x0 + x
			if flip {
				sx = x0 + w - 1 - x
			}
			out.Set(x, y, im.At(sx, b.Min.Y+y))
		}
	}
	return out
}
