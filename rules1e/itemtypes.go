package rules1e

//go:generate sh -c "go run ./idgen -spec ids.txt -dir . -spells SpellIDs -items ItemIDs -itemtype ItemType > ids.go"

// ItemType is an item's base type: the Type of an item record and the
// index into the ITEMS table (itemBase). Weapons 0x01..0x2F follow the
// order of their names in itemWords; the rest are named after the items of
// that type in the game's ITEM and MONnITM blocks. Only the types the code
// tests are named.
type ItemType byte

const (
	ItemHandAxe           ItemType = 0x02
	ItemBillGuisarme      ItemType = 0x05
	ItemClub              ItemType = 0x07
	ItemDagger            ItemType = 0x08
	ItemDart              ItemType = 0x09
	ItemHammer            ItemType = 0x14
	ItemJavelin           ItemType = 0x15
	ItemQuarrel           ItemType = 0x1C
	ItemSpear             ItemType = 0x1F
	itemQuarterStaff      ItemType = 0x21
	itemBastardSword      ItemType = 0x22
	itemBroadSword        ItemType = 0x23
	ItemLongSword         ItemType = 0x24
	itemShortSword        ItemType = 0x25
	ItemTwoHandedSword    ItemType = 0x26
	ItemCompositeLongBow  ItemType = 0x29
	itemCompositeShortBow ItemType = 0x2A
	itemLongBow           ItemType = 0x2B
	ItemShortBow          ItemType = 0x2C
	ItemHeavyCrossbow     ItemType = 0x2D // also the Fine Composite Long Bow's in the data (fixItemRecord)
	itemLightCrossbow     ItemType = 0x2E
	ItemSling             ItemType = 0x2F
	ItemLeatherArmor      ItemType = 0x32
	itemPaddedArmor       ItemType = 0x33
	itemStuddedLeather    ItemType = 0x34
	itemRingMail          ItemType = 0x35
	itemScaleMail         ItemType = 0x36
	ItemChainMail         ItemType = 0x37
	itemSplintMail        ItemType = 0x38
	itemBandedMail        ItemType = 0x39
	ItemPlateMail         ItemType = 0x3A
	ItemShield            ItemType = 0x3B
	ItemMageScroll        ItemType = 0x3D
	ItemClericScroll      ItemType = 0x3E
	ItemGauntlets         ItemType = 0x3F // of Ogre Power
	ItemCloak             ItemType = 0x43
	ItemRing              ItemType = 0x45
	ItemValuable          ItemType = 0x46 // jewellery, manuals, curios
	ItemPotion            ItemType = 0x47 // also holy symbols
	ItemArrow             ItemType = 0x49
	ItemBracers           ItemType = 0x4D
	ItemWand              ItemType = 0x4E
	itemWandOfMissiles    ItemType = 0x4F
	ItemStrengthPotion    ItemType = 0x54 // of Giant Strength
	ItemHolyWater         ItemType = 0x55
	ItemOil               ItemType = 0x56
	itemBoulder           ItemType = 0x57
	itemGiantBoulder      ItemType = 0x58 // 1d12+8, range 20; no item record has it
	ItemFeatherFallRing   ItemType = 0x5A
	ItemCloakDisplacement ItemType = 0x5C
	ItemProtectionRing    ItemType = 0x5D
)
