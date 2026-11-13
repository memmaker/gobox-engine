package ui

import (
	"fmt"
	"image"

	"gobox/asset"
)

// A member's own PNGs (a portrait or combat icon dropped on the sheet) come
// from the game as their bytes; gobox decodes them and cuts the icon's
// poses. Both are kept by content, so a picture keeps its identity (the
// cards' and figures' caches go by image) and a changed PNG never shows a
// stale one.

// ownPNGs are the PNGs decoded, by their bytes.
var ownPNGs = map[string]image.Image{}

// ownPoseKey is one cut of a PNG icon: its bytes, pose (0 ready, 1 action) and facing.
type ownPoseKey struct {
	png  string
	pose int
	flip bool
}

// ownPoses are the icon PNGs' cuts (ownPose).
var ownPoses = map[ownPoseKey]image.Image{}

// ownPNG is the picture of b, a PNG the game checked when it came in
// (asset.ReadPNG, asset.IconShape); nil for none.
func ownPNG(b []byte) image.Image {
	if b == nil {
		return nil
	}
	if im, ok := ownPNGs[string(b)]; ok {
		return im
	}
	im, err := asset.DecodePNG(b)
	if err != nil {
		panic(fmt.Sprintf("a member's PNG was not checked when it came in: %v", err))
	}
	ownPNGs[string(b)] = im
	return im
}

// ownPose is the icon PNG b's pose (0 ready, 1 action; a square PNG is
// both), mirrored to face left when flip; nil without a PNG.
func ownPose(b []byte, pose int, flip bool) image.Image {
	im := ownPNG(b)
	if im == nil {
		return nil
	}
	if r := im.Bounds(); r.Dx() == r.Dy() {
		pose = 0
	}
	k := ownPoseKey{string(b), pose, flip}
	if p, ok := ownPoses[k]; ok {
		return p
	}
	ownPoses[k] = asset.IconPose(im, pose, flip)
	return ownPoses[k]
}
