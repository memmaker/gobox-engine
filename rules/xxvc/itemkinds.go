package xxvc

import "gobox/rules"

// The Buck Rogers series' base item kinds (the buck-rogers itemicons
// items.rec is keyed by them). BR1 and BR2 share the item types 1..38
// (ItemTemplate.Type, the ITEMS file); a kind is an item type, never a name:
// the Martian Sword Gun is a Heat Gun (type 14), the Mercurian Mono Bolt a
// Mono Sword (type 8).
const (
	ItemIDKnife                rules.ItemID = 1
	ItemIDMonoKnife            rules.ItemID = 2
	ItemIDCutlass              rules.ItemID = 3
	ItemIDSword                rules.ItemID = 4
	ItemIDCrossbow             rules.ItemID = 5 // D.R. X-bow
	ItemIDNeedleGun            rules.ItemID = 6
	ItemIDBoltGun              rules.ItemID = 7
	ItemIDMonoSword            rules.ItemID = 8
	ItemIDLaserPistol          rules.ItemID = 9
	ItemIDRocketPistol         rules.ItemID = 10
	ItemIDRocketRifle          rules.ItemID = 11
	ItemIDMicrowaveGun         rules.ItemID = 12
	ItemIDLaserRifle           rules.ItemID = 13
	ItemIDHeatGun              rules.ItemID = 14
	ItemIDSonicStunner         rules.ItemID = 15
	ItemIDGrenadeLauncher      rules.ItemID = 16
	ItemIDPlasmaThrower        rules.ItemID = 17
	ItemIDRocketLauncher       rules.ItemID = 18
	ItemIDPolearm              rules.ItemID = 19
	ItemIDSpacesuit            rules.ItemID = 20
	ItemIDSmartsuit            rules.ItemID = 21
	ItemIDHeavyBodyArmor       rules.ItemID = 22
	ItemIDBattleArmor          rules.ItemID = 23
	ItemIDBattleArmorFields    rules.ItemID = 24
	ItemIDDemoCharge           rules.ItemID = 25
	ItemIDECMPackage           rules.ItemID = 26
	ItemIDGoggles              rules.ItemID = 27
	ItemIDBreathingMask        rules.ItemID = 28
	ItemIDRope                 rules.ItemID = 29
	ItemIDJetpack              rules.ItemID = 30
	ItemIDPoisonAntidote       rules.ItemID = 31
	ItemIDMiniExplosiveGrenade rules.ItemID = 32
	ItemIDExplosiveGrenade     rules.ItemID = 33
	ItemIDGasGrenade           rules.ItemID = 34
	ItemIDStunGrenade          rules.ItemID = 35
	ItemIDDazzleGrenade        rules.ItemID = 36
	ItemIDChaffGrenade         rules.ItemID = 37
	ItemIDAerosolMistGrenade   rules.ItemID = 38
)

// ItemKinds is the base kind of each item type (ItemTemplate.Type), BR1's
// and BR2's alike (both ITEMS files and ITEM0 templates, 2026-11-14).
var ItemKinds = map[int32]rules.ItemID{
	1: ItemIDKnife, 2: ItemIDMonoKnife, 3: ItemIDCutlass, 4: ItemIDSword, 5: ItemIDCrossbow,
	6: ItemIDNeedleGun, 7: ItemIDBoltGun, 8: ItemIDMonoSword, 9: ItemIDLaserPistol, 10: ItemIDRocketPistol,
	11: ItemIDRocketRifle, 12: ItemIDMicrowaveGun, 13: ItemIDLaserRifle, 14: ItemIDHeatGun, 15: ItemIDSonicStunner,
	16: ItemIDGrenadeLauncher, 17: ItemIDPlasmaThrower, 18: ItemIDRocketLauncher, 19: ItemIDPolearm,
	20: ItemIDSpacesuit, 21: ItemIDSmartsuit, 22: ItemIDHeavyBodyArmor, 23: ItemIDBattleArmor, 24: ItemIDBattleArmorFields,
	25: ItemIDDemoCharge, 26: ItemIDECMPackage, 27: ItemIDGoggles, 28: ItemIDBreathingMask, 29: ItemIDRope,
	30: ItemIDJetpack, 31: ItemIDPoisonAntidote, 32: ItemIDMiniExplosiveGrenade, 33: ItemIDExplosiveGrenade,
	34: ItemIDGasGrenade, 35: ItemIDStunGrenade, 36: ItemIDDazzleGrenade, 37: ItemIDChaffGrenade, 38: ItemIDAerosolMistGrenade,
}
