package goldbox

import "gobox/core"

// Pics are a Gold Box record's own picture numbers and its shared-pool picks:
// what a ruleset's creation hands its engine to name the pictures (the
// engine holds the picture files).
type Pics struct {
	PortHead, PortBody byte
	IconHead, IconBody byte
	IconSize           byte // 1 small, 2 large
	Pool               core.PoolPics
}

// PortraitPic is a portrait head or body: its picture id in HEADn/BODYn.DAX of set n.
type PortraitPic struct {
	Set Set
	Id  byte
}
