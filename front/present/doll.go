package present

import (
	_ "embed"
	"image"
)

// The paper doll every game's inventory shows: the doll and the slot boxes (the item
// icons: the series' shared sheet, sprites.ItemIcons), by DollSlots (Inventory.Doll's order).
var (
	//go:embed dollart/doll.png
	defaultDoll []byte
)

// defaultDollBoxes are the doll's slot boxes (18x18 with the frame, the
// rings' 10x10) in doll pixels, by DollSlots.
var defaultDollBoxes = []image.Rectangle{
	image.Rect(27, 62, 45, 80),  // weapon
	image.Rect(75, 62, 93, 80),  // shield
	image.Rect(21, 22, 39, 40),  // armor
	image.Rect(22, 42, 40, 60),  // gauntlets
	image.Rect(95, 1, 113, 19),  // helm
	image.Rect(97, 40, 115, 58), // belt
	image.Rect(97, 58, 115, 76), // robe
	image.Rect(84, 21, 102, 39), // cloak
	image.Rect(74, 83, 92, 101), // boots
	image.Rect(25, 82, 35, 92),  // ring
	image.Rect(37, 82, 47, 92),  // ring2
	image.Rect(22, 2, 40, 20),   // arrows: the Missiles slot
	image.Rect(5, 62, 23, 80),   // launcher
}

var defaultSlotNames = []string{"Weapon", "Shield", "Armour", "Gauntlets", "Helm", "Belt", "Robe", "Cloak",
	"Boots", "Ring", "Ring", "Missiles", "Launcher"}
