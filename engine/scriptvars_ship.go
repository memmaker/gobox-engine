package engine

import (
	"gobox/rules"
)

// shipCells are the script fields of the party's ship and the last foe (BR1 layout 4D16..4D46: the record Area[512..619], types.pas
// TShip), each the ShipCellFact the ruleset answers; the engine keeps no copy. Systems are in the ruleset's order
// (Hull, Sensors, Control, Life support, Fuel, Engines); a mount's cells are its status, ammo and reloads.
var shipCells = map[string]rules.ShipCellFact{
	"ShipHull":          {Part: rules.ShipSystemStatus, Index: 0},
	"ShipSensors":       {Part: rules.ShipSystemStatus, Index: 1},
	"ShipControl":       {Part: rules.ShipSystemStatus, Index: 2},
	"ShipLifeSupport":   {Part: rules.ShipSystemStatus, Index: 3},
	"ShipFuel":          {Part: rules.ShipSystemStatus, Index: 4},
	"ShipEngines":       {Part: rules.ShipSystemStatus, Index: 5},
	"ShipMount0Status":  {Part: rules.ShipMountStatus, Index: 0},
	"ShipMount0Ammo":    {Part: rules.ShipMountAmmo, Index: 0},
	"ShipMount0Reloads": {Part: rules.ShipMountReloads, Index: 0},
	"ShipMount1Status":  {Part: rules.ShipMountStatus, Index: 1},
	"ShipMount1Ammo":    {Part: rules.ShipMountAmmo, Index: 1},
	"ShipMount1Reloads": {Part: rules.ShipMountReloads, Index: 1},
	"ShipMount2Status":  {Part: rules.ShipMountStatus, Index: 2},
	// the last foe's summary (Area[620..642], types.pas TShip's status, crew and weapons)
	"FoeHull":        {Foe: true, Part: rules.ShipSystemStatus, Index: 0},
	"FoeSensors":     {Foe: true, Part: rules.ShipSystemStatus, Index: 1},
	"FoeControl":     {Foe: true, Part: rules.ShipSystemStatus, Index: 2},
	"FoeLifeSupport": {Foe: true, Part: rules.ShipSystemStatus, Index: 3},
	"FoeEngines":     {Foe: true, Part: rules.ShipSystemStatus, Index: 5},
	"FoeCrew":        {Foe: true, Part: rules.ShipCrew},
	"FoeWeaponCount": {Foe: true, Part: rules.ShipWeaponCount},
}

// shipFields binds shipCells: a read asks the ruleset, a write sets its fact.
func (e *Engine) shipFields() map[string]engineField {
	m := make(map[string]engineField, len(shipCells))
	for name, f := range shipCells {
		m[name] = engineField{
			get: func(int) (int64, bool) { return int64(uint16(e.fact(rules.NoChar, f).(rules.Number))), true },
			set: func(_ int, v int64) {
				if err := e.ruleset().SetFact(rules.NoChar, f, rules.Number(v)); err != nil {
					panic("script: " + err.Error())
				}
			},
		}
	}
	return m
}
