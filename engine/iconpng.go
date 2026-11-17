package engine

import (
	"errors"

	"gobox/asset"
	"gobox/rules"
)

// errWholeIcons: the game's combat icons are its own whole pictures, picked from its sheet (artNames.MemberIcons).
var errWholeIcons = errors.New("this game's combat icons are picked from its own pictures")

// The combat icon from a PNG (extra): a file dropped on the icon editor,
// saved as NAME.icon.png beside the record. A square picture is both
// poses; one twice as wide as high is cut in the middle, the left half the
// ready pose, the right the action pose. The battle draws it in place of
// the built icon; the page's icons (the overview) stay as built. The game
// hands the front the bytes: ui decodes them and cuts the poses (ownpng.go).

// takeIconPNG makes the PNG b c's combat icon, or says why not (the icon
// editor, a PNG dropped on the creation sheet's icon).
func (e *Engine) takeIconPNG(cID rules.CharID, b []byte) bool {
	if e.artNames().MemberIcons != "" { // whole-picture icons (BR1): the rules hold a sheet block, no PNG
		e.say(errWholeIcons.Error())
		return false
	}
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
