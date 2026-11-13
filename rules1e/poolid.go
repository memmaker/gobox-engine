package rules1e

import (
	"gobox/rules"
)

// PoolIDs are gobox's opaque shared-pool picks for a character: the game only
// stores them and states them back (Member.PortraitHeadID, PortraitBodyID,
// IconHeadID, IconBodyID); "" is none.
type PoolIDs struct {
	PortraitHeadID, PortraitBodyID, IconHeadID, IconBodyID rules.PicID
	Roster                                                 rules.RosterID // the roster character's own number ("rosterid"; NoRoster: none yet)
}

// PoolMagic opens the Go-side trailer after the original record bytes of a
// .cha/.sav: PoolMagic, then keyed entries, each a key and a value as a
// uint16 LE length and its bytes. Keys are the Create/SetIcon Field names
// (poolKeys); unknown keys are skipped and "" values are not written, so more
// keys can be added later. Written only when an ID is set, so the original
// bytes of every other save stay identical; old saves have no trailer and
// load "".
const PoolMagic = "PID2"

const RosterKey = "rosterid" // the trailer's RosterID, decimal

var PoolKeys = []string{"portraithead", "portraitbody", "iconhead", "iconbody"} // the write order

func (p *PoolIDs) Fields() map[string]*rules.PicID {
	return map[string]*rules.PicID{"portraithead": &p.PortraitHeadID, "portraitbody": &p.PortraitBodyID, "iconhead": &p.IconHeadID, "iconbody": &p.IconBodyID}
}

// SetPoolID stores id as c's pool id of the field (a poolKeys name).
func (c *Character) SetPoolID(field string, id rules.PicID) {
	if dst := c.PoolIDs.Fields()[field]; dst != nil {
		*dst = id
	}
}

// pic is the pool id of pick k.
func (p *PoolIDs) pic(k rules.PoolPic) *rules.PicID {
	return [...]*rules.PicID{rules.PoolPortraitHead: &p.PortraitHeadID, rules.PoolPortraitBody: &p.PortraitBodyID, rules.PoolIconHead: &p.IconHeadID, rules.PoolIconBody: &p.IconBodyID}[k]
}
