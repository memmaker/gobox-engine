package engine

import (
	"gobox/game"
	"gobox/rules"
)

// state puts the pool picks in a State member.
func poolState(p rules.PoolPics, m *game.Member) {
	m.PortraitHeadID, m.PortraitBodyID, m.IconHeadID, m.IconBodyID = string(p.PortraitHead), string(p.PortraitBody), string(p.IconHead), string(p.IconBody)
}
