package engine

import (
	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// state puts the pool picks in a State member.
func poolState(p rules.PoolPics, m *game.Member) {
	m.PortraitHeadID, m.PortraitBodyID, m.IconHeadID, m.IconBodyID = string(p.PortraitHead), string(p.PortraitBody), string(p.IconHead), string(p.IconBody)
}

// poolPics is id's pictures as pool names: its pool picks, else the game's
// own portrait numbers' (a pick the game has a number for is stored as that
// number, asset.Set).
func (e *Engine) poolPics(id rules.CharID) rules.PoolPics {
	p := e.fact(id, rules.PoolPicsFact{}).(rules.PoolPics)
	pt := e.fact(id, rules.PortraitFact{}).(rules.Portrait)
	p.PortraitHead = asset.Pick(p.PortraitHead, e.portPic("HEAD", byte(pt.Head)))
	p.PortraitBody = asset.Pick(p.PortraitBody, e.portPic("BODY", byte(pt.Body)))
	return p
}
