package xxvc

import "gobox/core"

// Weapons is a hit on the weapon mounts (C_Sys.WEAPONS): a target, no gauge.
const Weapons core.ShipSystem = "weapons"

// PartyShip is the SHIPS record the party's ship starts as (GameInit
// PARTY_SHIP_ID 10, record 40 + 10).
const PartyShip core.ShipID = 50

// ShipWeaponKind is a ship weapon, shipWeapons' index (1..8); 0 an empty mount.
type ShipWeaponKind uint8

// ShipWeapon is one shipWeapons row (types.pas TShipWeaponDef).
type ShipWeapon struct {
	MaxRange    uint8
	ReloadAmmo  uint8
	Damage      uint8
	TargetBonus int8
	Name        string // ShipCbt's messages
	Short       string // weaponShortNames: the menus' label
}

// ShipMount is one weapon mount (TShipWeaponSlot).
type ShipMount struct {
	Kind      ShipWeaponKind
	Crew      bool // CREW_MOUNT: the crew fires it, not the pilot
	MaxStatus uint8
	Status    uint8
	Ammo      uint8
	Reloads   uint8
}

// ShipRecord is a ship (types.pas TShip, SHIPS' $6C-byte records), typed.
type ShipRecord struct {
	Name           string
	Tonnage        int16
	Length, Width  int16
	Status         [6]int16 // by ShipSystems
	Maximum        [6]int16
	Cargo          int16
	Crew           int16
	ArmorClass     int8
	AlternateSkill uint8 // Pilot Rocket when no crew member checks
	SystemSkill    uint8 // Repair Engine, Jury Rig
	SystemValue    uint8 // gunnery when no crew member fires
	BaseSpeed      int8
	Reaction       int8
	Maneuver       uint8
	Type           string // shipTypeNames[shipType]
	Race           core.Race
	Mounts         [8]ShipMount
}

// Explosion is DamageCrew's damage to each member when the party's hull goes.
type Explosion int8

const (
	ExplosionRolled Explosion = iota // BR1: 5d20 + hit point maximum
	ExplosionFixed                   // BR2: hit point maximum + 30
)
