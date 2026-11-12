package char

// bytesOf copies src's values into dst (the typed Pool fields into a later
// layout's bytes), as copy does.
func bytesOf[T ~uint8](dst []byte, src []T) {
	for i := range min(len(dst), len(src)) {
		dst[i] = byte(src[i])
	}
}

// From285 is a Pool of Radiance record's fields moved into the 422 layout:
// Curse's load_player (O0A1:14B3) moves after the name and abilities, which
// the importer sets itself (race limits), as are race and sex. Memorised
// spells, coins, the order and the pointers are not moved.
func From285(p Record285) Record422 {
	var r Record422
	r.THAC0Base = p.Thac0Base
	r.Class = byte(p.Class)
	r.Age = p.Age
	r.MaxHP = p.MaxHP
	bytesOf(r.SpellBook[:], p.SpellBook[:]) // spells known 1..56
	r.FightLevel = p.FightLevel
	r.Size = p.Size
	bytesOf(r.Saves[:], p.Saves[:])
	r.BaseMove = p.BaseMove
	r.MaxLevel = p.MaxLevel
	r.Kept74 = p.Kept74
	r.Kept75 = p.Kept75
	r.Undead = p.Undead
	bytesOf(r.Thief[:], p.Thief[:])
	r.Pad83 = p.Pad83
	r.Control = p.Control
	r.Share = p.Share
	r.Pad86 = p.Pad86
	bytesOf(r.Levels[:], p.Levels[:])
	r.Kind = p.Kind
	r.Alignment = p.Alignment
	bytesOf(r.BaseAttack[:], p.BaseAttack[:])
	bytesOf(r.BaseDice[:], p.BaseDice[:])
	bytesOf(r.BaseSides[:], p.BaseSides[:])
	bytesOf(r.BaseBonus[:], p.BaseBonus[:])
	r.BaseAC = p.BaseAC
	r.StrBonus = p.StrBonus
	r.RandomID = p.RandomID
	r.Exp = p.Exp
	r.ClassMask = p.ClassMask
	r.HPRoll = p.HPRoll
	bytesOf(r.SpellSlots[0:3], p.SpellSlots[0:3])   // cleric 1..3
	bytesOf(r.SpellSlots[10:13], p.SpellSlots[3:6]) // magic-user 1..3
	r.ExpValue = p.ExpValue
	r.ExpPerHP = p.ExpPerHP
	r.PortHead, r.PortBody = p.PortHead, p.PortBody
	r.IconHead, r.IconBody = p.IconHead, p.IconBody
	r.IconSize = p.IconSize
	bytesOf(r.IconColour[:], p.IconColour[:])
	r.ItemCount = p.ItemCount
	r.Hands = p.Hands
	r.SaveBonus = p.SaveBonus
	r.Weight = p.Weight
	r.Status = p.Status
	r.Okay = p.Okay
	r.Side = p.Side
	r.THAC0 = p.THAC0
	bytesOf(r.AC[:], p.AC[:])
	bytesOf(r.Attacks[:], p.Attacks[:])
	bytesOf(r.Dice[:], p.Dice[:])
	bytesOf(r.Sides[:], p.Sides[:])
	bytesOf(r.Bonus[:], p.Bonus[:])
	r.HP = p.HP
	r.Move = p.Move
	return r
}
