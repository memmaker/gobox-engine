package pool

import (
	"gobox/core"
	"gobox/game"
)

// state puts the pool picks in a State member.
func poolState(p core.PoolPics, m *game.Member) {
	m.PortraitHeadID, m.PortraitBodyID, m.IconHeadID, m.IconBodyID = string(p.PortraitHead), string(p.PortraitBody), string(p.IconHead), string(p.IconBody)
}
