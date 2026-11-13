package engine

import (
	"gobox/asset"
	"slices"

	"gobox/rules"
)

// creationHost is the Engine side of POR's creation (the ruleset's Host):
// the roster files, the picture table and Data's quick party.
type creationHost struct{ e *Engine }

// Creation is POR's creation rules.
func (e *Engine) Creation() rules.CreationRules {
	e.ruleset()
	return e.create
}

func (h creationHost) Selected() int                                { return h.e.sel }
func (h creationHost) Keep(id rules.CharID) error                   { return h.e.SaveCharacter(id, "") }
func (h creationHost) Join(id rules.RosterID) (rules.CharID, error) { return h.e.joinRoster(id) }

func (h creationHost) Portrait(p rules.Pics) (head, body rules.PicID) {
	return asset.Pick(p.Pool.PortraitHead, h.e.portPic("HEAD", p.PortHead)), asset.Pick(p.Pool.PortraitBody, h.e.portPic("BODY", p.PortBody))
}

func (h creationHost) IconPics(p rules.Pics) (head, body rules.PicID) {
	return asset.Pick(p.Pool.IconHead, h.e.iconPic("CHEAD", p.IconHead, p.IconSize)), asset.Pick(p.Pool.IconBody, h.e.iconPic("CBODY", p.IconBody, p.IconSize))
}

// SetPics sets p's pictures from d; p.IconSize is already d's.
func (h creationHost) SetPics(d rules.Draft, p *rules.Pics) {
	e := h.e
	asset.Set(d.Head, &p.PortHead, 0, &p.Pool.PortraitHead, func(b byte) rules.PicID { return e.portPic("HEAD", b) }, func(id rules.PicID) (byte, bool) { return e.portByte("HEAD", id) })
	asset.Set(d.Body, &p.PortBody, 0, &p.Pool.PortraitBody, func(b byte) rules.PicID { return e.portPic("BODY", b) }, func(id rules.PicID) (byte, bool) { return e.portByte("BODY", id) })
	asset.Set(d.Icon.Head, &p.IconHead, 0, &p.Pool.IconHead, func(b byte) rules.PicID { return e.iconPic("CHEAD", b, p.IconSize) }, func(id rules.PicID) (byte, bool) { return e.iconByte("CHEAD", id, p.IconSize) })
	asset.Set(d.Icon.Body, &p.IconBody, 0, &p.Pool.IconBody, func(b byte) rules.PicID { return e.iconPic("CBODY", b, p.IconSize) }, func(id rules.PicID) (byte, bool) { return e.iconByte("CBODY", id, p.IconSize) })
}

func (h creationHost) QuickParty() []rules.Template {
	var out []rules.Template
	for _, qp := range h.e.gameData.QuickParty {
		need := map[rules.Ability]int32{}
		for key, v := range qp.Need {
			if i := slices.Index(statKeys, key); i >= 0 {
				need[abilities[i]] = int32(v)
			}
		}
		out = append(out, rules.Template{
			Picks: h.e.quickPicks(qp),
			Name:  qp.Name, Head: rules.PicID(qp.Head), Body: rules.PicID(qp.Body), Need: need,
		})
	}
	return out
}

// statKeys are Data's QuickPick.Need keys in stat order (game.StatKeys).
var statKeys = []string{"str", "int", "wis", "dex", "con", "cha"}

// abilities are the abilities in stat order, STR .. CHA (statKeys, the creation's StatNames).
var abilities = [6]rules.Ability{rules.Strength, rules.Intelligence, rules.Wisdom, rules.Dexterity, rules.Constitution, rules.Charisma}
