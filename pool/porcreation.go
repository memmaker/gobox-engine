package pool

import (
	"gobox/importer/pictures"
	"slices"

	"gobox/core"
	"gobox/core/goldbox"
)

// creationHost is the Engine side of POR's creation (the ruleset's Host):
// the roster files, the picture table and Data's quick party.
type creationHost struct{ e *Engine }

// Creation is POR's creation rules.
func (e *Engine) Creation() core.CreationRules {
	e.ruleset()
	return e.create
}

func (h creationHost) Selected() int                              { return h.e.sel }
func (h creationHost) Keep(id core.CharID) error                  { return h.e.SaveCharacter(id, "") }
func (h creationHost) Join(id core.RosterID) (core.CharID, error) { return h.e.joinRoster(id) }

func (h creationHost) Portrait(p goldbox.Pics) (head, body core.PicID) {
	return pictures.Pick(p.Pool.PortraitHead, h.e.portPic("HEAD", p.PortHead)), pictures.Pick(p.Pool.PortraitBody, h.e.portPic("BODY", p.PortBody))
}

func (h creationHost) IconPics(p goldbox.Pics) (head, body core.PicID) {
	return pictures.Pick(p.Pool.IconHead, h.e.iconPic("CHEAD", p.IconHead, p.IconSize)), pictures.Pick(p.Pool.IconBody, h.e.iconPic("CBODY", p.IconBody, p.IconSize))
}

// SetPics sets p's pictures from d; p.IconSize is already d's.
func (h creationHost) SetPics(d core.Draft, p *goldbox.Pics) {
	e := h.e
	pictures.Set(d.Head, &p.PortHead, 0, &p.Pool.PortraitHead, func(b byte) core.PicID { return e.portPic("HEAD", b) }, func(id core.PicID) (byte, bool) { return e.portByte("HEAD", id) })
	pictures.Set(d.Body, &p.PortBody, 0, &p.Pool.PortraitBody, func(b byte) core.PicID { return e.portPic("BODY", b) }, func(id core.PicID) (byte, bool) { return e.portByte("BODY", id) })
	pictures.Set(d.Icon.Head, &p.IconHead, 0, &p.Pool.IconHead, func(b byte) core.PicID { return e.iconPic("CHEAD", b, p.IconSize) }, func(id core.PicID) (byte, bool) { return e.iconByte("CHEAD", id, p.IconSize) })
	pictures.Set(d.Icon.Body, &p.IconBody, 0, &p.Pool.IconBody, func(b byte) core.PicID { return e.iconPic("CBODY", b, p.IconSize) }, func(id core.PicID) (byte, bool) { return e.iconByte("CBODY", id, p.IconSize) })
}

func (h creationHost) QuickParty() []core.Template {
	var out []core.Template
	for _, qp := range h.e.gameData.QuickParty {
		need := map[core.Ability]int32{}
		for key, v := range qp.Need {
			if i := slices.Index(statKeys, key); i >= 0 {
				need[abilities[i]] = int32(v)
			}
		}
		out = append(out, core.Template{
			Picks: h.e.quickPicks(qp),
			Name:  qp.Name, Head: core.PicID(qp.Head), Body: core.PicID(qp.Body), Need: need,
		})
	}
	return out
}

// statKeys are Data's QuickPick.Need keys in stat order (game.StatKeys).
var statKeys = []string{"str", "int", "wis", "dex", "con", "cha"}

// abilities are the abilities in stat order, STR .. CHA (statKeys, the creation's StatNames).
var abilities = [6]core.Ability{core.Strength, core.Intelligence, core.Wisdom, core.Dexterity, core.Constitution, core.Charisma}
