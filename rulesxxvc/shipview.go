package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// The ship as the front reads it: the panel's and the sheet's figures (charhlp.pas DrawShipDetails, DrawShipWeapons).

// mountLines are the ship's mounts in slot order, a slot with no weapon an empty line.
func (r *Rules) mountLines(s xxvc.ShipRecord) []rules.ShipMount {
	lines := make([]rules.ShipMount, len(s.Mounts))
	for i, m := range s.Mounts {
		if m.Kind != 0 {
			lines[i] = rules.ShipMount{Short: r.weapon(m).Short, Count: int32(m.MaxStatus), Working: int32(m.Status), Ammo: int32(m.Ammo), Reloads: int32(m.Reloads)}
		}
	}
	return lines
}

// ShipState is the fight's snapshot, outside a fight the party's ship alone.
func (s shipRules) ShipState() rules.ShipFight {
	r := s.r
	if r.space != nil {
		return r.shipSnapshot()
	}
	own := r.partyShip()
	sh := rules.Ship{ID: xxvc.PartyShip, Name: own.Name, Systems: map[rules.ShipSystem]rules.Gauge{}, Mounts: r.mountLines(own)}
	for n, sys := range xxvc.ShipSystems {
		sh.Systems[sys] = rules.Gauge{Now: int32(own.Status[n]), Max: int32(own.Maximum[n])}
	}
	return rules.ShipFight{Own: sh}
}

// ShipSheetOf is the party's ship's record, or the foe's while a fight is on (CharHlp.ShipSpeed for the speed).
func (s shipRules) ShipSheetOf(foe bool) rules.ShipSheet {
	r := s.r
	side, rec := partySide, r.partyShip()
	f := &shipFight{ships: [2]xxvc.ShipRecord{rec, {}}}
	if r.space != nil {
		f = r.space
		rec = f.ships[partySide]
	}
	if foe {
		if r.space == nil {
			return rules.ShipSheet{}
		}
		side, rec = foeSide, f.ships[foeSide]
	}
	return rules.ShipSheet{Name: rec.Name, Type: rec.Type, Race: raceNames[rec.Race],
		Tonnage: int32(rec.Tonnage), Length: int32(rec.Length), Width: int32(rec.Width), Cargo: int32(rec.Cargo),
		Armor: int32(rec.ArmorClass), Crew: int32(rec.Crew), Speed: int32(f.speed(side)), Reaction: int32(rec.Reaction),
		Mounts: r.mountLines(rec)}
}
