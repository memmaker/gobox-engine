package rulesxxvc

import "gobox/rules"

// Looks in flight (rules.FlightFact), as cbtacts.pas AnimateMissile picks a shot's picture by the weapon's item type.
// The engine and the adventure's art say which picture each kind is.

// The ranged weapons' itemTypes rows (cbtacts.pas CROSSBOW_TYPE.., c_item.pas).
const (
	crossbowType     = 5
	needleGunType    = 6
	boltGunType      = 7
	laserPistolType  = 9
	rocketPistolType = 10
	rocketRifleType  = 11
	microwaveGunType = 12
	laserRifleType   = 13
	heatGunType      = 14
)

// itemFlights is AnimateMissile's picture by item type; any other type (a grenade, the grenade launcher) is FlightDefault.
var itemFlights = map[int32]rules.FlightKind{
	plasmaThrowerType:  rules.FlightHeatBolt,
	heatGunType:        rules.FlightHeatBolt,
	laserPistolType:    rules.FlightLaser,
	laserRifleType:     rules.FlightLaser,
	crossbowType:       rules.FlightCrossbowBolt,
	boltGunType:        rules.FlightGunBolt,
	needleGunType:      rules.FlightGunBolt,
	rocketLauncherType: rules.FlightRocket,
	rocketPistolType:   rules.FlightRocket,
	rocketRifleType:    rules.FlightRocket,
	microwaveGunType:   rules.FlightWave,
	sonicStunnerType:   rules.FlightWave,
}

// flight is the kind of what a shot of item flies as. XXVc has no spells.
func (r *Rules) flight(item rules.ItemID) rules.FlightKind {
	return itemFlights[r.tables.Items[item].Type]
}
